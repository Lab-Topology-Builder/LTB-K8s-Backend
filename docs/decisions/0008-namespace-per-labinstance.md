# Namespace Per LabInstance

## Context and Problem Statement

Lab instances could be created in separate namespaces or one namespace for all lab instances.

## Considered Options

* One Namespace for all lab instances
* Namespace per lab instance

## Decision Outcome

Chosen option: "Namespace per lab instance", because it will be easier in the future to implement features like network policies and resource quotas and limits.
This approach ensures easier management and isolation of each lab instance within its dedicated namespace.
