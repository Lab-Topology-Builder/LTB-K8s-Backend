# Concepts

## Lab Topology Builder (LTB)

LTB is a tool that allows you to build a topology of virtual machines and containers, which are connected to each other according to the network topology you have defined.
<!-- TODO: add LTB image -->

## LTB Kubernetes Operator

The LTB Kubernetes Operator  custom Kubernetes controller that allows you to deploy and manage applications and their components on Kubernetes using custom resources.
<!-- TODO: add image -->

## Kubernetes operator

A Kubernetes operator is an application-specific controller that extends the functionality of the Kubernetes API to create, configure, and manage instances of complex applications on behalf of a Kubernetes user.
It builds upon the basic Kubernetes resource and controller concepts but includes domain or application-specific knowledge to automate the entire life cycle of the software it manages.

## Custom Resource (CR)

A custom resource is an extension of the Kubernetes API that allows you to define and manage your own API objects.
It provides a way to store and retrieve structured data and can be used with a custom controller to provide a declarative API.
Custom resources can be defined as a Kubernetes API extension using Custom Resource Definitions (CRDs) or via API aggregation.

## Custom Resource Definition (CRD)


Custom Resource Definition (CRD) is a Kubernetes API that allows you to define your own custom resources.
<!-- TODO: add code example -->

## Kubernetes Cluster

Kubernetes Cluster is a set of nodes that run containerized applications managed by Kubernetes.
<!-- TODO: add image -->

## Lab Template

Lab Template is a YAML file that defines the topology of the lab. It contains information about the devices that are part of the lab, as well as the network topology.
<!-- TODO: add code example -->

## Device

A device is either a KubeVirt VM or a container. It has a device type, a device version, and a device name.

## Lab

Lab is a YAML file (CR) that defines a lab instance. It holds the details from the lab template, and some additional information, such as the lab name.
It also has a status field that is updated by the Kubernetes operator.
<!-- TODO: add code example -->

## Lab Instance (Lab) (LTB)

Lab Instance is a deployment of the lab template. It is deployed by the Kubernetes Operator in the Kubernetes cluster.
<!-- TODO: add image -->
