# Operator Scope

## Context and Problem Statement

The Operator could be a namespace-scoped or cluster-scoped.

## Considered Options

* Namespace-scoped
* Cluster-scoped

## Decision Outcome

Chosen option: "Cluster-scoped", because cluster-scoped operators enables you to manage namespaces or resources in the entire cluster. They are also capable of managing infrastructure-level resources, such as nodes. Additionally, cluster-scoped operators provide us greater visibility and control over the entire cluster.

## More Information

* [Operator Scope](https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/)
