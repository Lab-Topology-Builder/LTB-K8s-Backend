# DevContainer

## Context and Problem Statement

Every team member could set up their development environment manually or they can create an automated and same development environment for everyone by using a DevContainer.

## Considered Options

* DevContainer
* Manual Setup

## Decision Outcome

Chosen option: "DevContainer", because DevContainer setup lets you create the same development environment for all team members, which ensures consistency. It also provides a completely isolated development environment, which helps to avoid software incompatibility issues, such as operator-sdk not working on Windows. Moreover, DevContainer is easily portable and can be shared between team members irrespective of the operating system they use.

## More Information

* [DevContainer](http://bit.ly/3TQ8zhx)
