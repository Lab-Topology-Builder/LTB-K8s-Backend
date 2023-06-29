# User Guide

## Installation Pre-requisites

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

1. Install the operator by creating a catalog source and subscription.
```sh
kubectl apply -f https://raw.githubusercontent.com/Lab-Topology-Builder/LTB-K8s-Backend/main/install/catalogsource.yaml -f https://raw.githubusercontent.com/Lab-Topology-Builder/LTB-K8s-Backend/main/install/subscription.yaml
```

2. Wait for the operator to be installed
```sh
kubectl get csv -n operators -w
```

## Usage

To create a lab you'll need to create at least one node type and one lab template.
Node types define the basic properties of a node. For VMs this includes everything that can be defined in a [Kubevirt VirtualMachineSpec](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachinespec) and for pods everything that can be defined in a [Kubernetes PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podspec-v1-core).

In order to provide better reusability of node types, you can use [Go templating Syntax](https://golang.org/pkg/text/template/) to include information from the lab template (like configuration or node name) in the node type.
The following example node types show how this can be done. You can use them as a starting point for your own node types.

### Example Node Type

This is an example of a VM node type. It creates a VM with 2 vCPUs and 4GB of RAM, using the Ubuntu 22.04 container disk image from [quay.io/containerdisks/ubuntu](https://quay.io/repository/containerdisks/ubuntu?tab=tags) and the `cloudInitNoCloud` volume source to provide a cloud-init configuration to the VM.

Everything that is defined in the `node` field of the lab template is available to the node type via the `.` variable.
Example: `{{ .Name }}` will be replaced with the name of the node from the lab template.

Currently, you cannot provide the cloud-init configuration as a YAML string via the .Config field of the lab template. Instead, you have to encode it as base64 string and therefore use the `userDataBase64` field of the volume source, because of indentation issues while rendering configuration.

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

This is an example of a generic pod node type. It creates a pod with a single container. The container name, container image, command and ports to expose are taken from the lab template.

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

After you have defined some node types, you can create a lab template.
A lab template defines the nodes that should be created for a lab, how they should be configured and how they should be connected.

### Example Lab Template

This is an example of a lab template, which you can use as a starting point for your own labs.
It uses the previously defined node types to create a VM and two pods. They are referenced via the `nodeTypeRef` field.
The provided ports will be exposed to the host network and can be accessed via the node's IP address and the port number assigned by Kubernetes. You can retrieve the IP address of a node by running `kubectl get node -o wide` and the port number by running `kubectl get svc`.

Currently, there is no support for point to point connections between nodes. Instead, they are all connected to the same network.
In the future, we plan to add support for point to point connections, which will be able to be defined as neighbors in the lab template.
The syntax for this is not yet definite, but it will probably look something like this:

```yaml
  neighbors:
  - "sample-node-1:1,sample-node-2:1"
  - "sample-node-2:2-sample-node-3:1"
```

This would connect the first port of `sample-node-1` to the first port of `sample-node-2` and the second port of `sample-node-2` to the first port of `sample-node-3`.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: LabTemplate
metadata:
  name: labtemplate-sample
spec:
  nodes:
  - name: "sample-node-1"
    nodeTypeRef:
      type: "nodetypeubuntuvm"
    config: "I2Nsb3VkLWNvbmZpZwpwYXNzd29yZDogdWJ1bnR1CmNocGFzc3dkOiB7IGV4cGlyZTogRmFsc2UgfQpzc2hfcHdhdXRoOiBUcnVlCnBhY2thZ2VzOgogLSBxZW11LWd1ZXN0LWFnZW50CiAtIGNtYXRyaXgKcnVuY21kOgogLSBbIHN5c3RlbWN0bCwgc3RhcnQsIHFlbXUtZ3Vlc3QtYWdlbnQgXQo="
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
  - name: "sample-node-2"
    nodeTypeRef:
      type: "genericpod"
      image: "ghcr.io/insrapperswil/network-ninja"
      version: "latest"
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
    config: '["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]'
  - name: "sample-node-3"
    nodeTypeRef:
      type: "genericpod"
      image: "ubuntu"
      version: "22.04"
    ports:
    - name: "ssh"
      port: 22
      protocol: "TCP"
    config: '["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]'
```

With the lab template defined, you can create a lab instance.

## Example Lab Instance

This is an example of lab instance, which you can use as a starting point for your own labs.
The lab instance references the previously defined lab template with the `labTemplateReference` field.
You also need to provide a DNS address via the `dnsAddress` field. This address will be used to create routes for the web terminal to the lab nodes.
For example, if you use the address `example.com`, the console of a node called `sample-node-1` will be available at `https://labinstance-sample-sample-node-1.example.com/` via a web terminal.

Currently, there is no support to edit the lab instance after it has been created. If you want to change the lab, you have to delete the lab instance and create a new one.

```yaml
apiVersion: ltb-backend.ltb/v1alpha1
kind: LabInstance
metadata:
  name: labinstance-sample
spec:
  labTemplateReference: "labtemplate-sample"
  dnsAddress: "example.com"
```

## Uninstall

1. Delete the subscription
```sh
kubectl delete subscriptions.operators.coreos.com -n operators ltb-subscription
```

2. Delete the CSV
```sh
kubectl delete csv -n operators ltb-operator.<version>
```

3. Delete the CRDs
```sh
kubectl delete crd labinstances.ltb-backend.ltb labtemplates.ltb-backend.ltb nodetypes.ltb-backend.ltb
```

4. Delete operator
```sh
kubectl delete operator ltb-operator.operators
```

5. Delete the CatalogSource
```sh
kubectl delete catalogsource.operators.coreos.com -n operators ltb-catalog
```
