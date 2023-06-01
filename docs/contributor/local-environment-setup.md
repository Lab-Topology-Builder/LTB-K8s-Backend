# Test to run application locally using K3d

[K3d](https://k3d.io/) was used  to run a local Kubernetes cluster to test the application if it works as expected. K3d was preferred because of its simplicity and ease of use. It is a lightweight wrapper to run Kubernetes in Docker. Other tools such as [Minikube](https://minikube.sigs.k8s.io/docs/) could have been used as well, but K3d was chosen because of the reasons mentioned above.

After installing all the dependencies listed in the [User Guide](../user-guide.md) and [Development cluster setup](dev-cluster-setup.md) sections, deploying lab instances was possible, and the pods were running as expected, but the virtual machines had a pending status. The reason for this was, Kubevirt was crashing. Because of the lack of time, we didn't investigate the issue further, which means we can't say for sure if the issue was Kubevirt's incompatible with K3d or something else.
