# Interaction with Operator

## Context and Problem Statement

There are multiple ways to interact with the operator, such as a GitOps approach, using the frontend via the API or using `kubectl`.

## Considered Options

* GitOps
* Use frontend
* Use kubectl
* All

## Decision Outcome

Chosen option: "All", because these options are not mutually exclusive and can be used together, we want to support all of them.
