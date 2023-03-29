# Operator SDK (Operator Framework)

## Context and Problem Statement

Other tools and libraries could be used to build the operator, such as [KubeBuilder](https://book.kubebuilder.io/), [Kopf](https://github.com/nolar/kopf), etc.

## Considered Options

* Operator SDK (Operator Framework)
* KubeBuilder

## Decision Outcome

Chosen option: "Operator SDK (Operator Framework)", because it provides a higher level of abstraction for creating Kubernetes operators, which makes it easier to write and manage operators. Additionally, there are tools and libraries for building and testing the operator included in Operator SDK.

## More Information

* [Operator SDK](https://sdk.operatorframework.io/)
* [Tools to build an operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
