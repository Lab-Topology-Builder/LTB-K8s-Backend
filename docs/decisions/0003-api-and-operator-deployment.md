# API and Operator Deployment

## Context and Problem Statement

The operator and the API could be separated and deployed as two services/containers or they could be deployed as one service/container.

## Considered Options

* One container
* Separate containers

## Decision Outcome

Chosen option: "Separate containers", because it provides more flexibility and scalability. It also makes it easier to update the operator and the API separately. Additionally, it is easy to separate the API and the operator into two different services, as they talk to each other via the Kubernetes API.
