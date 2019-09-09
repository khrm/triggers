# EventListener
`EventListeners` connect `TriggerBindings` to `TriggerTemplates` and provide an addressable endpoint, which is where webhooks/events are directed. It also define an optional field called`validate` to validate event using a predefined task. This task shouldn't require any resource. Only Input params can be provided. During taskrun, payload is passed as param `Payload` and those headers which have been defined as param in task are also passed. Refer [`validate-event`](validate-event.md) 
`validate` requires `taskRef`, `serviceAccountName` and `params`.
Further, it is as this level that the service account is connected, which specifies what permissions the resources will be created (or at least attempted) with.
When an `EventListener` is successfully created, a service is created that references a listener pod. This listener pod accepts the incoming events and does what has been specified in the corresponding `TriggerBindings`/`TriggerTemplates`.

<!-- FILE: examples/eventlisteners/eventlistener.yaml -->
```YAML
apiVersion: tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: listener
  namespace: tekton-pipelines
spec:
  serviceAccountName: default
  triggers:
    - binding:
        name: pipeline-binding
      template:
        name: pipeline-template
      validate:
        taskRef:
          name: validateTaskName
        serviceAccountName: saName
        params:
        - name: paramName
          value: paramValue
```
