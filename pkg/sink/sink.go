/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sink

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelineclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersclientset "github.com/tektoncd/triggers/pkg/client/clientset/versioned"

	"github.com/tektoncd/triggers/pkg/template"
	"github.com/tidwall/gjson"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	discoveryclient "k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
)

type Resource struct {
	TriggersClient         triggersclientset.Interface
	DiscoveryClient        discoveryclient.DiscoveryInterface
	RESTClient             restclient.Interface
	PipelineClient         pipelineclientset.Interface
	EventListenerName      string
	EventListenerNamespace string
}

const (
	validateTaskWait = 20 * time.Second
)

func (r Resource) HandleEvent(response http.ResponseWriter, request *http.Request) {
	el, err := r.TriggersClient.TektonV1alpha1().EventListeners(r.EventListenerNamespace).Get(r.EventListenerName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting EventListener %s in Namespace %s: %s", r.EventListenerName, r.EventListenerNamespace, err)
		return
	}

	event, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading event body: %s", err)
		return
	}

	// Execute each Trigger
	for _, trigger := range el.Spec.Triggers {
		// Secure Endpoint
		if trigger.TriggerValidate != nil {
			if err := r.secureEndpoint(trigger, request.Header, event); err != nil {
				log.Printf("Error securing Endpoint for TriggerBinding %s in Namespace %s: %s", trigger.TriggerBinding.Name, r.EventListenerNamespace, err)
				continue
			}
		}

		binding, err := template.ResolveBinding(trigger,
			r.TriggersClient.TektonV1alpha1().TriggerBindings(r.EventListenerNamespace).Get,
			r.TriggersClient.TektonV1alpha1().TriggerTemplates(r.EventListenerNamespace).Get)
		if err != nil {
			log.Print(err)
			continue
		}
		resources, err := template.NewResources(event, binding)
		if err != nil {
			log.Print(err)
			continue
		}
		err = createResources(resources, r.RESTClient, r.DiscoveryClient)
		if err != nil {
			log.Print(err)
		}
	}
}

func createResources(resources []json.RawMessage, restClient restclient.Interface, discoveryClient discoveryclient.DiscoveryInterface) error {
	for _, resource := range resources {
		if err := createResource(resource, restClient, discoveryClient); err != nil {
			return err
		}
	}
	return nil
}

// createResource uses the kubeClient to create the resource defined in the
// TriggerResourceTemplate and returns any errors with this process
func createResource(rt json.RawMessage, restClient restclient.Interface, discoveryClient discoveryclient.DiscoveryInterface) error {
	// Assume the TriggerResourceTemplate is valid (it has an apiVersion and Kind)
	apiVersion := gjson.GetBytes(rt, "apiVersion").String()
	kind := gjson.GetBytes(rt, "kind").String()
	namespace := gjson.GetBytes(rt, "metadata.namespace").String()
	namePlural, err := findAPIResourceNamePlural(discoveryClient, apiVersion, kind)
	if err != nil {
		return err
	}
	uri := createRequestURI(apiVersion, namePlural, namespace)
	result := restClient.Post().
		RequestURI(uri).
		Body([]byte(rt)).
		SetHeader("Content-Type", "application/json").
		Do()
	if result.Error() != nil {
		return result.Error()
	}
	return nil
}

// apiResourceName returns the plural resource name for the apiVersion and kind
func findAPIResourceNamePlural(discoveryClient discoveryclient.DiscoveryInterface, apiVersion, kind string) (string, error) {
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(apiVersion)
	if err != nil {
		return "", xerrors.Errorf("Error getting kubernetes server resources for apiVersion %s: %s", apiVersion, err)
	}
	for _, apiResource := range resourceList.APIResources {
		if apiResource.Kind == kind {
			return apiResource.Name, nil
		}
	}
	return "", xerrors.Errorf("Error could not find resource with apiVersion %s and kind %s", apiVersion, kind)
}

// createRequestURI returns the URI for a request to the kubernetes API REST endpoint
// given apiVersion, namePlural, and namespace. If namespace is an empty string,
// then namespace will be excluded from the URI
func createRequestURI(apiVersion, namePlural, namespace string) string {
	var uri string
	if apiVersion == "v1" {
		uri = "api/v1"
	} else {
		uri = path.Join(uri, "apis", apiVersion)
	}
	if namespace != "" {
		uri = path.Join(uri, "namespaces", namespace)
	}
	uri = path.Join(uri, namePlural)
	return uri
}

func (r Resource) secureEndpoint(trigger triggersv1.Trigger, headers http.Header, payload []byte) error {

	params := []pipelinev1.Param{}
	params = append(params, pipelinev1.Param{
		Name: "Payload",
		Value: pipelinev1.ArrayOrString{
			Type:      pipelinev1.ParamTypeArray,
			StringVal: string(payload),
		},
	})

	for key := range headers {
		params = append(params, pipelinev1.Param{
			Name: key,
			Value: pipelinev1.ArrayOrString{
				Type:      pipelinev1.ParamTypeArray,
				StringVal: headers.Get(key),
			},
		})
	}

	tr := &pipelinev1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    r.EventListenerNamespace,
			GenerateName: trigger.TriggerValidate.TaskRef.Name,
		},
		Spec: pipelinev1.TaskRunSpec{
			Inputs: pipelinev1.TaskRunInputs{
				Params: params,
			},
			TaskRef:        trigger.TriggerValidate.TaskRef,
			ServiceAccount: trigger.TriggerValidate.ServiceAccount,
		},
	}

	tr, err := r.PipelineClient.TektonV1alpha1().TaskRuns(r.EventListenerNamespace).Create(tr)
	if err != nil {
		return err
	}

	for {
		tr, err := r.PipelineClient.TektonV1alpha1().TaskRuns(r.EventListenerNamespace).Get(tr.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if tr.IsSuccessful() {
			break
		}

		time.Sleep(validateTaskWait)

		if tr.IsDone() && !tr.IsSuccessful() {
			return errors.New("validation taskrun failed")
		}
	}
	return nil
}
