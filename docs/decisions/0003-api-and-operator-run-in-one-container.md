# API and Operator Run in One Container

## Context and Problem Statement

The operator and the API could be separated and deployed as two services/containers.

## Considered Options

* One container
* Separate containers

## Decision Outcome

Chosen option: "One container", because it simplifies the deployment and implementation as everything will be written in one programming language, with the drawback that the components can't be exchanged easily.