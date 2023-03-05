# Conventions

## Naming

The following naming conventions are used in the project:

**Go**: *camelCase for variables and functions, and PascalCase for types and methods should be used.*

- **LTBKubernetesOperator**: a Kubernetes operator that manages lab instances of the LTB application
- **LTBKubernetesCluster**: a Kubernetes cluster that is used to deploy lab instances of the LTB application
- **labTemplate**: lab definition that will be used as a custom resource definition (CRD)
- **admin**: privileged user
- **user**: non-privileged user
- **labInstance**: a lab that is deployed in a Kubernetes cluster
- **labStatus**: a status of a lab instance (running, stopped, etc.)
- **networkDevice**: a device that is part of a lab network
- **networkDeviceType**: a type of a network device
- **networkDeviceVersion**: a version of a network device
- **networkDeviceName**: a name of a network device
- **networkDeviceGroup**: a group the network device belongs to
- **labReservation**: a reservation of resources for a lab instance
- TODO: add more

## Coding

- The Go extension in VSCode has a linting capability, so that will be used for linting.

## Git workflow

- Branch **main** should be used for merging finished features.
- A new branch should be created for each work item and deleted after the work item is done. And it should be named after the work item in Jira.
- A merge request should be created after the work item is done and the work is reviewed by the other team member. If the merge request is approved, it can be merged into the main branch and the team member who approved the merge request should set the work item in Jira to done.
- If a merge request is set to draft, it means that the work is not done yet and it should not be merged into the main branch.
