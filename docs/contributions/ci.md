# Continuos Integration

We use [GitHub Actions](https://github.com/features/actions) as our CI/CD tool.
Currently, we have two workflows:

- [Deploy docs](#deploy-docs)
- [Operator CI](#operator-ci)

## Deploy docs

This workflow is used to deploy the [mkdocs](https://www.mkdocs.org/) documentation to GitHub Pages.
It is triggered on every push to the `main` branch that affects the documentation. Specifically, it is triggered when a file in the `docs` directory or the `mkdocs.yaml` configuration file has changed.

Check the [deploy-docs-ci.yaml](https://github.com/Lab-Topology-Builder/LTB-K8s-Backend/edit/main/.github/workflows/deploy-docs-ci.yml) action for more details.

## Operator CI

This workflow is used to test, build and push the LTB Operator image, CatalogSource and Bundle to the GitHub Container Registry.
It is triggered for every push to a pull request and for every push to the `main` branch.

You can find this action here: [operator-ci.yaml](https://github.com/Lab-Topology-Builder/LTB-K8s-Backend/blob/main/.github/workflows/operator-ci.yml)

### Test

The test step runs the unit tests of the LTB Operator and fails the pipeline if any test fails.
Additionally, a coverage report is generated and uploaded to [Codecov](https://app.codecov.io/gh/Lab-Topology-Builder/LTB-K8s-Backend).
Pull requests are checked for the code coverage and will block the merge if the coverage drops below 80%. This is ensured by the Codecov GitHub integration and defined in the [codecov.yml](https://github.com/Lab-Topology-Builder/LTB-K8s-Backend/blob/main/codecov.yml) file.

### Build

We use the GitHub actions provided by [Docker](https://github.com/docker) to build and push the LTB Operator image to the GitHub Container Registry.
Additionally, we use [cosign](https://github.com/sigstore/cosign) to sign the images, so that users can verify the authenticity of the image.

### Additional deployment artifacts

To be able to deploy the LTB Operator with the [Operator Lifecycle Manager](https://olm.operatorframework.io/), a [Bundle](https://olm.operatorframework.io/docs/tasks/creating-operator-bundle/) and a [CatalogSource](https://olm.operatorframework.io/docs/tasks/creating-a-catalog/) must be created.
These artifacts are created with the [Operator-SDK](https://sdk.operatorframework.io/), to simplify the pipeline these tasks have been exported to the [Makefile](https://github.com/Lab-Topology-Builder/LTB-K8s-Backend/blob/main/Makefile)
