# Remote-Access

## Context and Problem Statement

For the lab instances to be useful for the students, they need to be able to access the pods (containers) and VMs.
Access to pods/VMs should only be granted to user with the appropriate access rights.
It should be possible to access the pods/VMs console and or access it via multiple OOB protocols (SSH, RDP, VNC, etc.).

## Considered Options

* Kubernetes Service
* Gotty
* ttyd

## Decision Outcome

Chosen option: "ttyd and Kubernetes Service", ttyd will be used as a jump host to access the pods/VMs console. A Kubernetes service will be used to allow access the pods/VMs via OOB protocols.
Security for the console access will likely be easy to implement.
Secure access via OOB protocols was considered, but will need to be researched further, currently it would depend on the OOB protocol used and the security features it provides.
