# Operator SDK

## Context and Problem Statement

It's best practice to use an SDK to build operators for Kubernetes. The SDK provides a higher level of abstraction for creating Kubernetes operators, making it easier to write and manage operators.
There are multiple SDKs available for building operators.
We need a SDK that's flexible and easy to use and can be used with Go.

## Considered Options

* Operator SDK (Operator Framework)
* KubeBuilder
* Kopf
* KUDO
* Metacontroller

## Decision Outcome

Chosen option: "Operator SDK", because it provides a high level of abstraction for creating Kubernetes operators, making it easier to write and manage operators.
Additionally, the Operator SDK incorporates tools and libraries for building, testing and packaging operators, offering a user-friendly experience and is compatible with Go.

## Links

* [Operator SDK](https://sdk.operatorframework.io/)
* [Tools to build an operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
