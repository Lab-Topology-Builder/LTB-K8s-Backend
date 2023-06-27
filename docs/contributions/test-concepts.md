# Test Concept

## Overview

This document outlines the approaches, methodologies, and types of tests that will be used to ensure the LTB K8s Backend components are functioning as expected.

## Test categories

The tests will primarily focus on the following category:

<!-- - *Functionality and Logic*: This includes automated integration tests to evaluate how the LTB K8s Backend interacts with other components of the LTB application, such as the operator's function in a Kubernetes cluster with a K8s API server and other resources. -->

Testing in the other categories, such as Security and Performance, will be considered later time and their specifics will be determined accordingly.

## Tools

The tools listed below are going to be used to perform the tests mentioned above. Moreover, the tools are used in a suite test, which is created when a controller is scaffolded by the tool.

- [Testify](https://github.com/stretchr/testify): a go package that provides a set of features to perform unit tests, such as assertions, mocks, etc.
- [Ginkgo](https://onsi.github.io/ginkgo/): a Go testing framework for Go to help you write expressive, readable, and maintainable tests. It is best used with the [Gomega](https://onsi.github.io/gomega/) matcher library.
- [Gomega](https://onsi.github.io/gomega/): a Go matcher library that provides a set of matchers to perform assertions in tests. It is best used with the [Ginkgo](https://onsi.github.io/ginkgo/) testing framework.

## Strategies: Test Approach

<!--  -->

### Unit Tests

Unit tests are going to be used to test small pieces of code, such as functions, which don't involve setting up testing Kubernetes environment with a K8s API server and other resources.

### Integration Tests

<!--  -->
