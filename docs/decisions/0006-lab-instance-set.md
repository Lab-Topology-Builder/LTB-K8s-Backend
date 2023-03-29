# Lab Instance Set

## Context and Problem Statement

We could use one CR, LabInstanceSet, and tell the operator we want e.g. 10 LabInstances  by only providing one LabInstance CR or we could provide the operator 10 CRs to create 10 LabInstances.

## Considered Options

* With LabInstanceSet
* Without LabInstanceSet

## Decision Outcome

<!--TODO: update this documentation -->
Chosen option: "With LabInstanceSet", because it is easier to manage, and helps to avoid redundancy...
