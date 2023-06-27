# Coding Conventions

We are following the [Effective Go](https://golang.org/doc/effective_go) guidelines for coding conventions.
The following is a summary of the most important conventions.

## Naming

The following naming conventions are used in the project:

### [Naming conventions in Go](https://golang.org/doc/effective_go#names)

- *camelCase* for variables and functions, which are not exported
- *PascalCase* for types and functions that need to be exported

#### Examples

- **labInstanceStatus**: variable name for a status of a lab instance
- **UpdateLabInstanceStatus**: name for an exported function, starts with a capital letter

## [Formatting](https://golang.org/doc/effective_go#formatting)

We are using the [gofmt](https://golang.org/cmd/gofmt/) from the Go standard library to format our code.

[staticcheck](https://staticcheck.dev/) is used as a linter in addition to the formatting guidelines from Effective Go, because it is the default linter of the Go extension for VS Code.
