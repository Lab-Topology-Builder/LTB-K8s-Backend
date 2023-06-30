# Replace KVM/Docker-based LTB Backend

## Context and Problem Statement

The LTB K8s Backend could replace the KVM/Docker-based LTB Backend fully or partially by reusing parts, such as

## Considered Options

* Replace KVM/Docker-based LTB Backend fully
* Replace KVM/Docker-based LTB Backend partially

## Decision Outcome

Chosen option: "Replace KVM/Docker-based LTB Backend fully", because huge parts of the current LTB Backend would need to be rewritten to be compatible with the new LTB K8s operator and it would be easier to rewrite the whole backend. Additionally, the same programming language can be used throughout the whole Backend, Go.
