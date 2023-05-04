# User Guide

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
    image:
      type: "ubuntu"
      version: "22.04"
      kind: "vm"
    config: |-
              #cloud-config
              password: ubuntu
              chpasswd: { expire: False }
              ssh_authorized_keys:
                - <your-ssh-pub-key>
              packages:
                - qemu-guest-agent
              runcmd:
                - [ systemctl, start, qemu-guest-agent ]
  - name: "sample-node-2"
    image:
      type: "ghcr.io/insrapperswil/network-ninja"
      version: "latest"
  - name: "sample-node-3"
    image:
      type: "ubuntu"
      version: "latest"
      kind: "pod"
  connections:
  - neighbors: "TestHost1:1,TestHost2:1"
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
```

The above lab instance will create a lab instance called labinstance-sample using the data from the referenced resource labtemplate-sample, which is provided at the beginning as an example.
