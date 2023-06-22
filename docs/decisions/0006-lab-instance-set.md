# Lab Instance Set

## Context and Problem Statement

We could create a CR called LabInstanceSet, and tell the operator we want e.g. 10 LabInstances by providing one LabInstance CR and a generator like a list of names or we could provide the operator 10 CRs to create 10 LabInstances.

## Considered Options

* With LabInstanceSet
* Without LabInstanceSet

## Decision Outcome

Chosen option: "Without LabInstanceSet", because we currently don't see a need for it. This could change in the future, but for now we will not implement it.
