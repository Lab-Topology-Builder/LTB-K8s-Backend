# Concepts

## Kubernetes Cluster

Kubernetes Cluster is a set of nodes that run containerized applications managed by Kubernetes.
<!-- TODO: add image -->

## Lab Topology Builder (LTB)

The Lab Topology Builder (LTB) is a tool that allows you to build a topology of virtual machines and containers, which are connected to each other according to the network topology you have defined.
<!-- TODO: add LTB image -->

## LTB Kubernetes Operator

The LTB Kubernetes Operator  custom Kubernetes controller that allows you to deploy and manage applications and their components on Kubernetes using custom resources.
<!-- TODO: add image -->

## Kubernetes operator

A Kubernetes operator is an application-specific controller that extends the functionality of the Kubernetes API to create, configure, and manage instances of complex applications on behalf of a Kubernetes user.
It builds upon the basic Kubernetes resource and controller concepts but includes domain or application-specific knowledge to automate the entire life cycle of the software it manages.

## Custom Resource (CR)

A [custom resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#custom-controllers) (CR) is an extension of the Kubernetes API that allows you to define and manage your own API objects.
It provides a way to store and retrieve structured data and can be used with a custom controller to provide a declarative API.
Custom resources can be defined as a Kubernetes API extension using Custom Resource Definitions (CRDs) or via API aggregation.

## Custom Resource Definition (CRD)

A [custom resource definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) (CRD) is a Kubernetes native resource.
Defining a CRD object creates a new custom resource with a name and schema that you specify.
The custom resource created from a CRD object can be either namespaced or cluster-scoped.
CustomResourceDefinitions themselves are non-namespaced and are available to all namespaces.
<!-- TODO: add code example -->

## Lab Template

Lab Template is a YAML file that defines the topology of the lab. It contains information about the devices that are part of the lab, as well as the network topology.
<!-- TODO: add code example -->

## Node

In a network, a node represents any device that is connected.
Within LTB, a node can be either a KubeVirt virtual machine or a container.
Each node is characterized by its type, version, and name.

## Network Topology

The arrangement or pattern in which all nodes on a network are connected together is referred to as the network’s topology.

## Lab

A lab defines a set of nodes that are connected together according to a network topology.

## Lab Instance

A lab instance is a custom resource and describes a lab that is deployed in a Kubernetes cluster.
It defines the name, which lab template to use and also has a status field that is updated by the operator.
<!-- TODO: add example -->
