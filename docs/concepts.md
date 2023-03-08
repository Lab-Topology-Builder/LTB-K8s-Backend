# Concepts

## Lab Topology Builder (LTB)

LTB is a tool that allows you to build a topology of virtual machines and containers, which are connected to each other according to the network topology you have defined.
<!-- TODO: add LTB image -->

## LTB Kubernetes Operator

LTB Kubernetes Operator is a custom Kubernetes controller that allows you to deploy and manage applications and their components on Kubernetes using custom resources.
<!-- TODO: add image -->

## Custom Resource (CR)

Custom Resource is an extension of LTB Kubernetes API that allows you to create your own objects and store them in the Kubernetes cluster.  In LTB, CR is the YAML file that is provided by the user.

## Custom Resource Definition (CRD)

Custom Resource Definition (CRD) is a Kubernetes API that allows you to define your own custom resources.
<!-- TODO: add code example -->

## Kubernetes Cluster

Kubernetes Cluster is a set of nodes that run containerized applications. It is managed by the Kubernetes operator.
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

## Lab Instance (Lab)

Lab Instance is a deployment of the lab template. It is deployed by the Kubernetes Operator in the Kubernetes cluster.
<!-- TODO: add image -->
