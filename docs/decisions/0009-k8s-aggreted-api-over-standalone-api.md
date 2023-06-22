# K8s Aggreted API Over Standalone API

## Context and Problem Statement

We want a declarative way of creating LTB labs inside Kubernetes using Kubernetes native pods and KubeVirt virtual machines.
We could either create a standalone API which interacts with the Kubernetes API and does not follow the Kubernetes API conventions. And therefore is not compatible with Kubernetes tools, such as dashboards or `kubectl`, but would allow complete control over the API design.
Or we could create an aggregated API which uses the Kubernetes aggregation layer to extend the Kubernetes API, which would allow us to use Kubernetes tools, such as dashboards or `kubectl`, but would limit the control over the API design.

## Considered Options

* Standalone API
* Aggregated API

## Decision Outcome

Chosen option: "Aggregated API", because Is better suited for declerative API, our new types will be readable and writeable using kubectl/Kubernetes tools, such as dashboards.
We also can leverage Kuberbetes API support features this way.
Additionally, our resources are scoped to a cluster or namespaces of a cluster.
Finally, the Operator Pattern is simpler to implement this way.

## Links

* [Standalone vs Aggregated API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#should-i-add-a-custom-resource-to-my-kubernetes-cluster)
* [Aggregated API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)
* [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
