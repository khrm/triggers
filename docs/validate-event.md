# Validate Event Tekton Task
The validate event task provides a task to validate an incoming event to the addressable endpoint. No resource should be provided to this task. Also, it needs `Payload` forevent payload and required header parameter to be passed as Param. Additionally, if any Parameters are defined as part of `validate` under `event-listener`, they are also provided to taskrun during execution and needs to be added to Task also. Refer [`validate-event`](validate-event.md) 

Sample Task provided for [`validate-github-event`](validate-github-event.yaml) has been provided.
