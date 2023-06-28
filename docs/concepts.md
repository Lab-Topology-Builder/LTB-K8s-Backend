# Concepts

## Lab Topology Builder

The Lab Topology Builder is a network emulator that allows you to build a topology of virtual machines and containers, which are connected to each other according to the network topology you have defined.

## Network Topology

The arrangement or pattern in which all nodes on a network are connected together is referred to as the network’s topology.

Here is an example of a network topology:

![LTB](./assets/images/Lab-Topology.png)

## Lab

In our context, a lab refers to a networking lab consisting of interconnected nodes following a specific network topology.

## Kubernetes

Kubernetes is a portable, extensible, open source platform for managing containerized workloads and services, that facilitates both declarative configuration and automation. It has a large, rapidly growing ecosystem. Kubernetes services, support, and tools are widely available. [^1]
[^1]: Excerpt from [kubernetes.io](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/)

## Kubernetes Cluster

Kubernetes Cluster is a set of nodes that run containerized applications managed by Kubernetes. A Kubernetes cluster consists of a control plane and one or more nodes. The control plane is responsible for maintaining the desired state of the cluster, such as which applications are running and which container images they use, etc. And the nodes run the applications and workloads.

## Kubernetes operator

A Kubernetes operator is an application-specific controller that extends the functionality of the Kubernetes API to create, configure, and manage instances of complex applications on behalf of a Kubernetes user.
It builds upon the basic Kubernetes resource and controller concepts but includes domain or application-specific knowledge to automate the entire life cycle of the software it manages.
More information can be found [here](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

## LTB Operator

The LTB Operator is a K8s Operator for the LTB application, which is responsible for creating, configuring, and managing the emulated network topologies of the LTB application inside a Kubernetes cluster.
It also automatically updates the status of the labs based on the current state of the associated containers and virtual machines, ensuring accurate and real-time lab information.

## Custom Resource Definition (CRD)

A [custom resource definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) (CRD) is a Kubernetes native resource.
By defining a Custom Resource Definition (CRD), you can create custom resources with the specified name and schema.
These custom resources can be either namespaced or cluster-scoped, depending on your requirements.
CustomResourceDefinitions themselves are cluster-scoped and therefore are available to all namespaces.

## Custom Resource (CR)

A [custom resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#custom-controllers) (CR) is an extension of the Kubernetes API that allows you to define and manage your own API objects.
It provides a way to store and retrieve structured data and can be used with a custom controller to provide a declarative API.
Custom resources can be defined as a Kubernetes API extension using Custom Resource Definitions (CRDs) or via API aggregation.

## Lab Template

Lab Template is a CR, which defines a template for a lab. It contains information about which nodes are part of the lab, their configuration and how they are connected to each other.

## Lab Instance

A lab instance is a CR that describes a lab that you want to deploy in a Kubernetes cluster.
It has a reference to the lab template you want to use and also has a status field that is updated by the operator, which shows how many pods and VMs are running in the lab and the status of the lab instance itself. In addition, it also has a dns address field, which will be used to access the nodes using the web-based terminal.

## Node Type

In a network, a node represents any device that is part of the lab. A NodeType is a CR that defines a type of node that can be part of a lab. You reference the node type you want to have in your lab in the lab template.
Within LTB, a node can be either a KubeVirt virtual machine or a regular Kubernetes pod.
