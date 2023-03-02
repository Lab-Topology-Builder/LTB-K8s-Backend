# Lab Topology Builder Architecture

## LTBv1 Architecture
Currently the LTB is composed of two main components:

- [Frontend](#frontend) built with React
- [Backend](#backend) built with Django

### Backend

The backend is accessible via API and a Admin Web UI.
It is responsible for the following tasks:

- Parsing the YAML topology files
- Deploying/Destroying the containers and VMs
- Exposes status of lab deployments
- Exposes 
- Expose information on how to access the deployed containers and VMs
- Provide remote SSH capabilities
- Provide remote Wireshark capture capabilities
- Managing reservations
- Exposes node resource usage
- User Management

It is composed of the following components:

- [Lab Topology Builder Architecture](#lab-topology-builder-architecture)
  - [LTBv1 Architecture](#ltbv1-architecture)
    - [Backend](#backend)
      - [Orchestration](#orchestration)
      - [Reservations](#reservations)
      - [Running lab store](#running-lab-store)
      - [Template store](#template-store)
      - [Authentication](#authentication)
    - [Frontend](#frontend)

#### Orchestration

The orchestration component is responsible for creating different tasks using Celery and executing them on a remote host.
There are 4 different types of tasks:

- DeploymentTask
    - Deploys containers in docker
    - Deploys VMs using KVM
    - Creates connections between containers and VMs using an OVS bridge
- RemovalTask
- MirrorInterfaceTask
- SnapshotTask

#### Reservations
The reservation component is responsible for reserving system resources in advance.

#### Running lab store
This component is responsible for storing information about running labs.
#### Template store
This component is responsible for storing lab templates.
#### Authentication
This component is responsible for user authentication and management.
### Frontend