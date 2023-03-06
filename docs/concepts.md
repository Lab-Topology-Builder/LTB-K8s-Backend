# Concepts

## Lab Topology Builder (LTB)

LTB is a tool that allows you to build a topology of virtual machines and containers, which are connected to each other according to the network topology you have defined.

## Kubernetes Operator

Kubernetes Operator is a custom Kubernetes controller that allows you to deploy and manage applications and their components on Kubernetes using custom resources.

## Custom Resource (CR)

Custom Resource is an extension of Kubernetes API that allows you to create your own objects and store them in the Kubernetes cluster.

## Custom Resource Definition (CRD)

Custom Resource Definition (CRD) is a Kubernetes API that allows you to define your own custom resources. In LTB, CRD is the lab template (YAML) that is provided by the user.

## Kubernetes Cluster

Kubernetes Cluster is a set of nodes that run containerized applications. It is managed by the Kubernetes operator.

## Lab Template

Lab Template is a YAML file that defines the topology of the lab. It contains information about the virtual machines and containers that are part of the lab, as well as the network topology.

## Lab Instance (Lab)

Lab Instance is a deployment of the lab template. It is deployed by the Kubernetes Operator in the Kubernetes cluster.

