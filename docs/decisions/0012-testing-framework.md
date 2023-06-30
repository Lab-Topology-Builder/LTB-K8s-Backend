# Testing Framework

## Context and Problem Statement

Every project needs to be tested.
There are multiple testing libraries and frameworks to test Go applications, which can be used in addition to the default Go testing library.

## Considered Options

* Testify
* Ginkgo/Gomega
* GoSpec
* GoConvey

## Decision Outcome

Chosen option: "Ginkgo/Gomega", because it is widely used in the Kubernetes community to test Kubernetes operators. It is also used in the [Kubevirt](https://kubevirt.io/) project, that is used by the LTB Operator. Additionally, tests written with Ginkgo/Gomega are easy to read and understand.
