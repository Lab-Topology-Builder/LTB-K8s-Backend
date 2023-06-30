# Lab Instance Set

## Context and Problem Statement

One approach could involve creating a custom resource (CR) named `LabInstanceSet` and specifying our desired quantity of `LabInstances` to the LTB Operator. For instance, we could provide a single `LabInstance` along with a generator, such as a list of names, to indicate that we want 10 `LabInstances`. Alternatively, we could directly provide the operator with 10 separate `LabInstances` to create the desired quantity of 10.

## Considered Options

* With LabInstanceSet
* Without LabInstanceSet

## Decision Outcome

Chosen option: "Without LabInstanceSet", because we currently don't see a need for it. This could change in the future, but for now we will not implement it.
