# K8s Aggreted API Over Standalone API

## Context and Problem Statement

We want a declarative way of creating LTB labs inside Kubernetes using Kubernetes native pods and KubeVirt virtual machines.

## Considered Options

* Standalone API
* Aggregated API

## Decision Outcome

Chosen option: "Aggregated API", because Is better suited for declerative API, our new types will be readable and writeable using kubectl/Kubernetes tools, such as dashboards.
We also can leverage Kuberbetes API support features this way.
Our resources are scoped to a cluster or namespaces of a cluster.
