# Dev Container

## Context and Problem Statement

Every team member could set up their development environment manually or we could use a dev container to provide a consistent development environment for all team members and future contributors.

## Considered Options

* Dev Container
* Manual Setup

## Decision Outcome

Chosen option: "Dev Container", because a dev container setup lets you create the same development environment for all team members to ensure consistency. It also provides a completely isolated development environment, which helps to avoid software incompatibility issues, such as Operator-SDK not working on Windows. Moreover, a dev container is easily portable and works on all operating systems that support Docker.
The only downside is that not all IDEs support dev containers, but at least two of the currently most popular IDEs, namely VS Code and Visual Studio support dev containers.

## Links

* [DevContainer](http://bit.ly/3TQ8zhx)
* [Most Popular IDEs](https://survey.stackoverflow.co/2023/#most-popular-technologies-new-collab-tools)
