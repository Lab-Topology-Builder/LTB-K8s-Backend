# 00011-Remote-Access

## Context and Problem Statement

Remote access to pods (containers) could be done using kubernetes service (load-balancer or node-port) for every node or using a jump host, which redirects the traffic to the node the user wants to access.

## Considered Options

* Kubernetes Service
* Jump host

## Decision Outcome

Chosen option: "Jump host", because it will be easy to manage the traffic and implement access control. We still haven't fully decided, which option we are going to use, but the jump host option will likely be our choice. We will start implementing the first option (Kubernetes service) and extend it later.
