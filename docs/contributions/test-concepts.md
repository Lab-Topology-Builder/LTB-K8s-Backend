# Test Concept

This document outlines the approaches, methodologies, and types of tests that ensure that the LTB Operator components are functioning as expected.

## Test categories

The tests primarily focus on Functionality and Logic.
Security and Performance tests should be added in the future as the project matures.

## Tools

The following tools are used to test the LTB Operator, you can find more information why they were chosen in the [Testing Framework Decision](../decisions/0012-testing-framework.md):

- [Testing](https://pkg.go.dev/testing): The default Go testing library that provides support for automated testing of Go packages. It is intended to be used in concert with the "go test" command, which automates execution of any function of the form
- [Ginkgo](https://onsi.github.io/ginkgo/): a Go testing framework for Go to help you write expressive, readable, and maintainable tests. It is best used with the [Gomega](https://onsi.github.io/gomega/) matcher library.
- [Gomega](https://onsi.github.io/gomega/): a Go matcher library that provides a set of matchers to perform assertions in tests. It is best used with the [Ginkgo](https://onsi.github.io/ginkgo/) testing framework.

## Strategies: Test Approach

We focused on unit tests for the LTB Operator, as they were easy to implement and provide a good coverage of the code.
At a later stage of the project, we could add integration tests to ensure that the LTB Operator works as expected with other components like the LTB API.

We aspire to achieve approximately 90% test coverage, to increase the maintainability and stability of the LTB Operator.

### Unit Tests

Unit rely on the [Fake Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/fake) package from the [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) library to create a fake client that can be used to mock interactions with the Kubernetes API.
This allows us to test functions that interact with the Kubernetes API without mocking the complete API or using a real Kubernetes cluster.

### Integration Tests

The integration tests could be implemented by using the [EnvTest](https://https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) package from the [controller-runtime](https://https://pkg.go.dev/sigs.k8s.io/controller-runtime) library.
