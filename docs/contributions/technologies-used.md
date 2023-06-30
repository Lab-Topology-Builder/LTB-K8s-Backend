# Tools and Frameworks

The tools and frameworks used in the project are listed below.

## Go-based Operator SDK framework

To create the LTB Operator, we used the [Go-based Operator-SDK](https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/) framework. It provides a set of tools to simplify the process of building, testing and packaging our operator.

## Kubevirt

[Kubevirt](https://kubevirt.io/) is a tool that provides a virtual machine management layer on top of Kubernetes. It allows us to deploy virtual machines on Kubernetes.

## Kubernetes

We use [Kubernetes](https://kubernetes.io/) as the container orchestration platform for the LTB application.

## Multus CNI

To create multiple network interfaces for the pods, [Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni) is used.
