# Namespace Per LabInstance

## Context and Problem Statement

LabInstances could be created in separate namespaces or one namespace for all LabInstances.

## Considered Options

* One Namespace for all LabInstances
* Namespace per LabInstance

## Decision Outcome

Chosen option: "Namespace per LabInstance", because it will be easier in the future to implement features like RBAC and resource quotas and limits.
