
# Current (KVM/Docker)-base LTB Architecture

Currently the KVM/Docker based LTB is composed of the following containers:

- [Frontend](#frontend) built with React
- [Backend](#backend) built with Django
- [Deployment](#deployment) built with docker-compose

## Backend

The backend is accessible via API and a Admin Web UI.
It is responsible for the following tasks:

- parsing the yaml topology files
- deploying/destroying the containers and vms
- exposes status of lab deployments
- exposes information on how to access the deployed containers and vms
- provides remote ssh capabilities
- provides remote Wireshark capture capabilities
- managing reservations (create, delete, etc.)
- exposes node resource usage
- user management
- exposes information about a device (version, groups, etc.)

It is composed of the following components:

- [Reservations](#reservations)
- [Running lab store](#running-lab-store)
- [Template store](#template-store)
- [Authentication](#authentication)

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

### Reservations

The reservation component is responsible for reserving system resources in advance. It is responsible for the following tasks:

- Create a reservation
- Delete a reservation
- Update a reservation

### Running lab store

This component is responsible for storing information about running labs, such as:

- The devices taking part in the running lab, inclusive of the interfaces
- Connection information

### Template store

This component is responsible for storing lab templates.

### Authentication

This component is responsible for user authentication and management.

## Deployment

The deployment component is responsible for deploying the LTB backend and frontend components.

## Frontend

The frontend allows users to access their labs and their devices.
