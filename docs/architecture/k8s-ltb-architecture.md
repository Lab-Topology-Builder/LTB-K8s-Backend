# Kubernetes Lab Topology Builder Architecture

It is composed of a [Frontend](#frontend) and a [Backend](#backend).

![Architecture Overview](../assets/drawings/LTB-Architecture.drawio.svg)

## Frontend

The frontend is responsible for the following tasks:

- Providing a web UI for the user to interact with the labs.
- Providing a web UI for the admin to manage:
  - Users
  - Lab templates
  - Lab deployments
  - Reservations

## Backend

The backend is composed of the following components:

- Operator
- API

The backend is responsible for the following tasks:

- Parsing the yaml topology files
- Deploying/destroying the containers and vms
- Exposes status of lab instances
- Enables you to access the deployed containers and vms via different protocols
- Provides remote ssh capabilities
- Exposes information about a node (version, groups, etc.)
- Exposes node resource usage
- Provides remote Wireshark capture capabilities
- Managing reservations (create, delete, etc.)
- User management
- etc.

## C4 Model

### System Context Diagram

![C4 System Context](../assets/drawings/C4-System-Context.drawio.svg)

### Container Diagram

![C4 Container](../assets/drawings/C4-Container.drawio.svg)

### Component Diagram

![C4 Component](../assets/drawings/C4-Component.drawio.svg)

#### Legend

- <span style="color: #083f75">Dark blue</span>: represents Personas (User, Admin)
- <span style="color: #23a2d9">Blue</span>: represents Internal Components (Frontend Web UI, LTB K8s Backend)
- <span style="color: #63bef2">Light blue</span>: represents Components which will be implemented in this project (LTB Operator, LTB Operator API)
- <span style="color: #8c8496">Dark gray</span>: represents External Components (K8s, Keycloak)
