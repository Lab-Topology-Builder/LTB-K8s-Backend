# Lab Topology Builder Architecture

Currently the LTB is composed of two main components:

- [Frontend](#frontend) built with React
- [Backend](#backend) built with Django

## Backend

The backend is accessible via API and a Admin Web UI.
It is responsible for the following tasks:

- Parsing the YAML topology files
- Deploying/Destroying the containers and VMs
- Expose information on how to access the deployed containers and VMs
- Provide remote Wireshark capture capabilities
- Managing reservations

It is composed of the following components:

- [Orchestration](#orchestration)
- [Reservations](#reservations)
- [Running lab store](#running-lab-store)
- [Template store](#template-store)

### Orchestration

The orchestration component is responsible for creating different tasks using Celery and executing them on a remote host.
There are 4 different types of tasks:

- DeploymentTask
    - Deploys containers in docker
    - Deploys VMs using KVM
    - Creates connections between containers and VMs using an OVS bridge
- RemovalTask
- MirrorInterfaceTask
- SnapshotTask
## Frontend
