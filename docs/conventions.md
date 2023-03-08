# Conventions

## Naming

The following naming conventions are used in the project:

### [Naming conventions in Go](https://golang.org/doc/effective_go#names)

- *camelCase* for variables and functions
- *PascalCase* for types and functions that need to be exported

#### Examples

- **labStatus**: a status of a lab instance (running, stopped, etc.)
- **device**: a device that is part of a lab network
- **deviceType**: a type of a network device
- **deviceVersion**: a version of a network device
- **deviceName**: a name of a network device
- **deviceGroup**: a group the network device belongs to
- **labReservation**: a reservation of resources for a lab instance
<!-- TODO: add more -->

## Coding

- The Go extension in VSCode has a linting capability, so that will be used for linting.
