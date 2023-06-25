# Contributor Guide

Contributions are welcome and appreciated.

## How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

## Development Environment

You’ll need a Kubernetes cluster to run against. You can find instructions on how to setup your dev cluster in the [Dev Cluster Setup](./dev-cluster-setup.md) section.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

We recommend using VSCode with the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension. This will allow you to use our devcontainer, which has all the tools needed to develop and test the operator already installed.

### Prerequisites for recommended IDE setup

- [Docker](https://docs.docker.com/get-docker/)
- [VSCode](https://code.visualstudio.com/)
- [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension

### Getting Started

1. Clone the repository
2. Open the repository in VSCode
3. Click the popup or use the command palette to reopen the repository in a container (Dev Containers: Reopen in Container)

Now you are ready to start developing!

### Running the operator locally

You can run the operator locally on your machine, this is useful for quick testing and debugging. However, you will need to be aware that the operator will use the kubeconfig file on your machine, so you will need to make sure that the context is set to the cluster you want to run against. Therefore it also does not use the RBAC rules it would usually be deployed with.

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

#### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

#### Running the operator on the cluster

You can also run the operator on the cluster, this is useful for testing the operator in a more realistic environment.
However, you will first need to login to some container registry that the cluster can access, so that you can push the operator image to that registry.
This will allow you to test the operators RBAC rules.

Make sure to replace `<some-registry>` with the location of your container registry and `<tag>` with the tag you want to use.

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/ltb-operator:<tag>
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/ltb-operator:<tag>
```

##### Undeploy controller

Undeploy the controller from the cluster:

```sh
make undeploy
```

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

You can find more information on how to develop the operator in the [Operator SDK Documentation](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/) and the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
