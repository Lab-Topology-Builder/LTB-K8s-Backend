# Remote-Access

## Context and Problem Statement

For the labinstances to be usefull for the students, they need to be able to access the pods (containers) and VMs.
The access should be restricted to the pods/VMs the user is allowed to access.
It should be possible to access the pods/VMs console and or access it via multiple OOB protocols (SSH, RDP, VNC, ...).

## Considered Options

* Kubernetes Service
* Gotty
* ttyd

## Decision Outcome

Chosen option: "ttyd and Kubernetes Service", ttyd will be used as a jump host to access the pods/VMs console, and a Kubernetes service (LoadBalancer) will be used to access the pods/VMs via OOB protocols.
Security for the console access will likely be easy to implement.
Secure access via OOB protocols was considered, but will need to be reasearched further.
