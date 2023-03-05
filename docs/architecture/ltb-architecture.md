# Lab Topology Builder Architecture

## LTBv1 Architecture

Currently the LTB is composed of two main components:

- [Frontend](#frontend) built with React
- [Backend](#backend) built with Django
- [Deployment](#deployment) built with docker-compose

### Backend

The backend is accessible via API and a Admin Web UI.
It is responsible for the following tasks:

- Parsing the YAML topology files
- Deploying/Destroying the containers and VMs
- Exposes status of lab deployments
- Exposes information on how to access the deployed containers and VMs
- Provides remote SSH capabilities
- Provides remote Wireshark capture capabilities
- Managing reservations (Create, Delete, etc.)
- Exposes node resource usage
- User Management
- Exposes information about a device (version, groups, etc.)
- Exposes device metrics

It is composed of the following components:

- [Lab Topology Builder Architecture](#lab-topology-builder-architecture)
  - [LTBv1 Architecture](#ltbv1-architecture)
    - [Backend](#backend)
      - [Orchestration](#orchestration)
      - [Reservations](#reservations)
      - [Running lab store](#running-lab-store)
      - [Template store](#template-store)
      - [Authentication](#authentication)
    - [Deployment](#deployment)
    - [Frontend](#frontend)

#### Orchestration

The orchestration component is responsible for creating different tasks using Celery and executing them on a remote host.
There are 4 different types of tasks:

- DeploymentTask
  - Deploys containers in docker
  - Deploys VMs using KVM
  - Creates connections between containers and VMs using an OVS bridge
- RemovalTask
  - Removes a running lab
- MirrorInterfaceTask
  - Creates a mirror interface on a connection
- SnapshotTask
  - Takes a snapshot of a running lab

#### Reservations

The reservation component is responsible for reserving system resources in advance. It is responsible for the following tasks:

- Create a reservation
- Delete a reservation
- Update a reservation

#### Running lab store

This component is responsible for storing information about running labs, such as:

- The devices taking part in the running lab, inclusive of the interfaces
- Connection information

#### Template store

This component is responsible for storing lab templates.

#### Authentication

This component is responsible for user authentication and management.

### Deployment

The deployment component is responsible for deploying the LTB backend and frontend components.

### Frontend
