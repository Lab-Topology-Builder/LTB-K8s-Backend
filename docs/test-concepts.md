# Test Concept

## Introduction

The approaches and methodologies used to test the LTB K8s Backend components, including the scope, tools, strategies and environment are described in this document. Moreover, the types of tests that are going to be performed are also identified.

## Scope

The scope of the tests is to ensure that the LTB K8s Backend components are working as expected. The following tests are going to be performed on the LTB K8s Backend components:

- **Functionality and Logic**: Automated integration tests, the LTB K8s Backend will be tested to see how it behaves when integrated with other components of the LTB application. For example, how the operator functions in a Kubernetes cluster with a K8s API server and other resources.

The tests listed below will be performed if the time permits and how they are going to be performed will be decided later:

- Security
- Performance
- Availability
 
## Tools

The tools that are going to be used to perform the tests are:

- [Testify](https://github.com/stretchr/testify): a go package that provides a set of features to perform unit tests, such as assertions, mocks, etc.
- [EnvTest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest): a Go library that helps write integration tests for Kubernetes controllers by setting up an instance of etcd and a Kubernetes API server, without kubelet, controller-manager, or other components.
- [Ginkgo](https://onsi.github.io/ginkgo/): a Go testing framework for Go to help you write epxressive, readable, and maintainable tests. It is best used with the [Gomega](https://onsi.github.io/gomega/) matcher library.
- [Gomega](https://onsi.github.io/gomega/): a Go matcher library that provides a set of matchers to perform assertions in tests. It is best used with the [Ginkgo](https://onsi.github.io/ginkgo/) testing framework.


## Strategies: Test Approach

The following test approaches are going to be used to test the LTB K8s Backend components:

### Unit Tests

Unit tests are going to be used to test small pieces of code, such as functions, which don't involve setting up testing Kubernetes environment with a K8s API server and other resources.

### Integration Tests

Integration tests are going to be used to test the different components of the LTB K8s Backend, such as the operator, the controllers, etc., and how they interact with each other.

## Environment

[EnvTest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) is going to be used to set up a testing Kubernetes environment with a K8s API server and other resources. The testing environment is going to be set up in a Docker container.