# Remote-Access

## Context and Problem Statement

For the lab instances to be useful for the students, they need to be able to access the pods (containers) and VMs.
Access to pods/VMs should only be granted to user with the appropriate access rights.
It should be possible to access the console of the pods/VMs and employ various out-of-band (OOB) protocols such as SSH, RDP, VNC, and more.

## Considered Options

* Kubernetes Service
* Gotty
* ttyd

## Decision Outcome

Chosen option: "ttyd and Kubernetes Service", because ttyd can be used as a jump host to access the pods/VMs' console. A Kubernetes service can be used to allow access the pods/VMs via OOB protocols.
Security for the console access will be easy to implement.
Secure access for OOB protocols was considered, but needs to be researched further. Currently, it depends on the chosen OOB protocol and the security features it provides.
