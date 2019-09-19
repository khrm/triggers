package builder

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinetb "github.com/tektoncd/pipeline/test/builder"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventListenerOp is an operation which modifies the EventListener.
type EventListenerOp func(*v1alpha1.EventListener)

// EventListenerSpecOp is an operation which modifies the EventListenerSpec.
type EventListenerSpecOp func(*v1alpha1.EventListenerSpec)

// EventListenerTriggerOp is an operation which modifies the EventListenerSpec.Trigger.
type EventListenerTriggerOp func(*v1alpha1.Trigger)

// EventListenerTriggerValidateOp is an operation which modifies the TriggerValidate.
type EventListenerTriggerValidateOp func(*v1alpha1.TriggerValidate)

// EventListener creates an EventListener with default values.
// Any number of EventListenerOp modifiers can be passed to transform it.
func EventListener(name, namespace string, ops ...EventListenerOp) *v1alpha1.EventListener {
	e := &v1alpha1.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, op := range ops {
		op(e)
	}

	return e
}

// EventListenerMeta sets the Meta structs of the EventListener.
// Any number of MetaOp modifiers can be passed.
func EventListenerMeta(ops ...MetaOp) EventListenerOp {
	return func(e *v1alpha1.EventListener) {
		for _, op := range ops {
			switch o := op.(type) {
			case ObjectMetaOp:
				o(&e.ObjectMeta)
			case TypeMetaOp:
				o(&e.TypeMeta)
			}
		}
	}
}

// EventListenerSpec sets the specified spec of the EventListener.
// Any number of EventListenerSpecOp modifiers can be passed to create/modify it.
func EventListenerSpec(ops ...EventListenerSpecOp) EventListenerOp {
	return func(e *v1alpha1.EventListener) {
		for _, op := range ops {
			op(&e.Spec)
		}
	}
}

// EventListenerServiceAccount sets the specified ServiceAccount of the EventListener.
func EventListenerServiceAccount(saName string) EventListenerSpecOp {
	return func(spec *v1alpha1.EventListenerSpec) {
		spec.ServiceAccountName = saName
	}
}

// EventListenerTrigger adds a Trigger to the EventListenerSpec Triggers.
func EventListenerTrigger(apiVersion string, ops ...EventListenerTriggerOp) EventListenerSpecOp {
	return func(spec *v1alpha1.EventListenerSpec) {
		trigger := &v1alpha1.Trigger{}
		for _, op := range ops {
			op(trigger)
		}
		spec.Triggers = append(spec.Triggers, *trigger)
	}
}

// EventListenerTriggerBindingRef adds a TriggerBinding to the Triger in EventListenerSpec Triggers.
func EventListenerTriggerBindingRef(tbName, apiVersion string) EventListenerTriggerOp {
	return func(trigger *v1alpha1.Trigger) {
		trigger.TriggerBinding = v1alpha1.TriggerBindingRef{
			Name:       tbName,
			APIVersion: apiVersion,
		}
	}
}

// EventListenerTriggerBindingRef adds a TriggerTemplate to the Triger in EventListenerSpec Triggers.
func EventListenerTriggerTemplateRef(ttName, apiVersion string) EventListenerTriggerOp {
	return func(trigger *v1alpha1.Trigger) {
		trigger.TriggerTemplate = v1alpha1.TriggerTemplateRef{
			Name:       ttName,
			APIVersion: apiVersion,
		}
	}
}

// EventListenerTriggerValidate adds a TrjggerValidate to the Triger in EventListenerSpec Triggers.
func EventListenerTriggerValidate(ops ...EventListenerTriggerValidateOp) EventListenerTriggerOp {
	return func(trigger *v1alpha1.Trigger) {
		validate := &v1alpha1.TriggerValidate{}
		for _, op := range ops {
			op(validate)
		}
		trigger.TriggerValidate = validate
	}
}

// EventListenerTriggerValidateTaskRef adds a TaskRef to the TrigerValidate.
func EventListenerTriggerValidateTaskRef(taskName, apiVersion string, kind pipelinev1.TaskKind) EventListenerTriggerValidateOp {
	return func(validate *v1alpha1.TriggerValidate) {
		validate.TaskRef = pipelinev1.TaskRef{
			Name:       taskName,
			Kind:       kind,
			APIVersion: apiVersion,
		}
	}
}

// EventListenerTriggerServiceAccount adds a service account name to the TrigerValidate.
func EventListenerTriggerValidateServiceAccount(serviceAccount string) EventListenerTriggerValidateOp {
	return func(validate *v1alpha1.TriggerValidate) {
		validate.ServiceAccountName = serviceAccount
	}
}

// EventListenerTriggerValidateParam adds a param name to the TrigerValidate.
func EventListenerTriggerValidateParam(name, value string) EventListenerTriggerValidateOp {
	return func(validate *v1alpha1.TriggerValidate) {
		validate.Params = append(validate.Params,
			pipelinev1.Param{
				Name:  name,
				Value: *pipelinetb.ArrayOrString(value),
			},
		)
	}
}
