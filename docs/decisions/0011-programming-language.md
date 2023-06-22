# Programming Language

## Context and Problem Statement

We need to choose a programming language for the project.
The considered options are based on the supported languages of the Operator SDK.

## Considered Options

* Go
* Helm
* Ansible

## Decision Outcome

Chosen option: "Go", because many cloud native projects are written in Go, and it is a compiled language, which is more performant than interpreted languages. Also, Go is a statically typed language, which makes it easier to maintain and refactor the code. It is also easier to write complicated logic and tests in Go than in Helm or Ansible.
