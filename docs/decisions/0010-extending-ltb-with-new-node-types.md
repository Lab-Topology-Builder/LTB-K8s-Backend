# Extending LTB with New Node Types

## Context and Problem Statement

It should be possible to create, update and delete node types (e.g. Ubuntu, XRD, XR, IOS, Cumulus, etc.).
Node types should be used inside lab templates and expose a way to provide a node with configuration (cloud-init, zero-touch, etc.)
The amount of available network interfaces is dynamic and depends on how many connections a node has according to a specific lab template.

## Decision Drivers

* Certain operating systems' images like XR, and XRD need a specific interface configuration which depends on how many interfaces a certain node will receive.
* The chosen solution should support multiple versions of a type in an easy to use way (e.g. Ubuntu 22.04, 20.04, ...).
* For XRd images, interfaces need to have environment variables set for each interface they use, and the interface count needs to be dynamically set according to the lab template.
* For XR virtual machine images, the first interface is the management interface and then there are two empty interfaces that need a special configuration.
* For mount from config might be different
* Cumulus VX images need a privileged container
* XRd need additional privileges

## Considered Options

* Custom Resources
* Go

## Decision Outcome

Chosen option: "Custom Resources", because it will be possible to support all the cases mentioned in the decision drivers using Go templates and CRs. Implementing the types in Go does not seem to bring any major advantages, whereas using CRs will be easier for external users to extend the system with new node types.

### Positive Consequences

* Easy to extend during runtime
* Easy to extend for external users
* All decision drivers will be supported

### Negative Consequences

* Go templates are not as powerful as Go, which could make it harder to implement certain node types.
* A Custom Resource is also a little bit less flexible than a Go type, but this should not be a problem for the use cases we have.

## Links

* [Example of similar solution using Go node types](https://github.com/vrnetlab/vrnetlab)
