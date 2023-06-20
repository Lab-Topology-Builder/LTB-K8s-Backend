# User Guide

## Installation (Pre-requisites)

| Tool | Version | Installation | Description |
| ---- | ------- | ------------ | ----------- |
|[Kubernetes](https://kubernetes.io/)| ^1.26.0 | [Installation](https://kubernetes.io/docs/setup/)| Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications. |
|[Kubevirt](https://kubevirt.io/) | 0.59.0 | [Installation](https://kubevirt.io/user-guide/#/installation/installation) | Kubevirt is a Kubernetes add-on to run virtual machines on Kubernetes. |
|[Multus-CNI](https://github.com/k8snetworkplumbingwg/multus-cni)| 3.9.0 |  [Installation](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md)| Multus-CNI is a plugin for K8s to attach multiple network interfaces to pods. |
|[Operator Lifecycle Manager](https://olm.operatorframework.io/)| ^0.24.0 | [Installation](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md) | Operator Lifecycle Manager (OLM) helps users install, update, and manage the lifecycle of all Operators and their associated services running across their Kubernetes clusters. |

Alternative OLM installation:

```sh
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.25.0/install.sh | bash -s v0.25.0
```

Change the version to the desired one.

## Installation of the LTB K8s operator

1. Install the operator by creating a catalog source and a subscription.

```sh
kubectl apply -f https://raw.githubusercontent.com/Lab-Topology-Builder/LTB-K8s-Backend/main/install/catalogsource.yaml -f https://raw.githubusercontent.com/Lab-Topology-Builder/LTB-K8s-Backend/main/install/subscription.yaml
```

1. Wait for the operator to be installed

```sh
kubectl get csv -n operators -w
```

## Uninstall

1. Delete the subscription

```sh
kubectl delete subscriptions.operators.coreos.com -n operators ltb-subscription
```

1. Delete the CSV

```sh
kubectl delete csv -n operators ltb-operator.<version>
```

1. Delete the CRDs

```sh
kubectl delete crd labinstances.ltb-backend.ltb labtemplates.ltb-backend.ltb nodetypes.ltb-backend.ltb
```

1. Delete operator

```sh
kubectl delete operator ltb-operator.operators
```

1. Delete the CatalogSource

```sh
kubectl delete catalogsource.operators.coreos.com -n operators ltb-catalog
```

## Example Node Type

This is an example of a VM node type, which you can use as a starting point for your own node types.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: NodeType
metadata:
  name: nodetypeubuntuvm
spec:
  kind: vm
  nodeSpec: |
    running: true
    template:
      spec:
        domain:
          resources:
            requests:
              memory: 4096M
          cpu:
            cores: 2
          devices:
            disks:
              - name: containerdisk
                disk:
                  bus: virtio
              - name: cloudinitdisk
                disk:
                  bus: virtio
        terminationGracePeriodSeconds: 0
        volumes:
          - name: containerdisk
            containerDisk:
              image: quay.io/containerdisks/ubuntu:22.04
          - name: cloudinitdisk
            cloudInitNoCloud:
              userDataBase64: {{ .Config }}
```

This is an example of a pod node type, which you can use as a starting point for your own node types.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: NodeType
metadata:
  name: genericpod
spec:
  kind: pod
  nodeSpec: |
    containers:
      - name: {{ .Name }}
        image: {{ .NodeTypeRef.Image}}:{{ .NodeTypeRef.Version }}
        command: {{ .Config }}
        ports:
          {{- range $index, $port := .Ports }}
          - name: {{ $port.Name }}
            containerPort: {{ $port.Port }}
            protocol: {{ $port.Protocol }}
          {{- end }}
```

## Example Lab Template

This is an example of lab template, which you can use as a starting point for your own labs.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: LabTemplate
metadata:
  name: labtemplate-sample
spec:
  nodes:
  - name: "sample-node-1"
    nodetyperef:
      type: "nodetypeubuntuvm"
      image: "ubuntu"
      version: "22.04"
    config: "I2Nsb3VkLWNvbmZpZwpwYXNzd29yZDogdWJ1bnR1CmNocGFzc3dkOiB7IGV4cGlyZTogRmFsc2UgfQpzc2hfcHdhdXRoOiBUcnVlCnBhY2thZ2VzOgogLSBxZW11LWd1ZXN0LWFnZW50CiAtIGNtYXRyaXgKcnVuY21kOgogLSBbIHN5c3RlbWN0bCwgc3RhcnQsIHFlbXUtZ3Vlc3QtYWdlbnQgXQo="
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
  - name: "sample-node-2"
    nodetyperef:
      type: "genericpod"
      image: "ghcr.io/insrapperswil/network-ninja"
      version: "latest"
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
    config: '["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]'
  - name: "sample-node-3"
    nodetyperef:
      type: "genericpod"
      image: "ubuntu"
      version: "22.04"
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
    config: '["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]'
  neighbors:
  - "TestHost1:1,TestHost2:1"
  - "TestHost1:2,TestHost3:1"

```

The above lab template will define three nodes and one connection between two of the nodes.

## Example Lab Instance

This is an example of lab instance, which you can use as a starting point for your own labs.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: LabInstance
metadata:
  name: labinstance-sample
spec:
  labTemplateReference: "labtemplate-sample"
  dnsAddress: "labinstance-sample.example.com"
```

The above lab instance will create a lab instance called labinstance-sample using the data from the referenced resource labtemplate-sample, which is provided at the beginning as an example.
