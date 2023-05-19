# Development cluster setup

Setup for a development Kubernetes cluster.

## Prepare Node

```bash
sudo apt update
sudo apt upgrade -y
sudo swapoff -a
sudo sed -ri '/\sswap\s/s/^#?/#/' /etc/fstab
```

## RKE2 Server Configuration

```bash
sudo mkdir -p /etc/rancher/rke2
sudo vim /etc/rancher/rke2/config.yaml
```

```yaml
# /etc/rancher/rke2/config.yaml
write-kubeconfig-mode: "0644"
kube-apiserver-arg: "allow-privileged=true"
cni: multus,cilium
disable-kube-proxy: true
```

### Cilium Configuration for Multus

```bash
sudo mkdir -p /var/lib/rancher/rke2/server/manifests
sudo vim /var/lib/rancher/rke2/server/manifests/rke2-cilium-config.yaml
```

```yaml
# /var/lib/rancher/rke2/server/manifests/rke2-cilium-config.yaml
# k8sServiceHost/Port IP of Control Plane node default Port 6443
---
apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: rke2-cilium
  namespace: kube-system
spec:
  valuesContent: |-
    cni:
      chainingMode: "none"
      exclusive: false
    kubeProxyReplacement: strict
    k8sServiceHost: "<NodeIP>"
    k8sServicePort: 6443
    operator:
      replicas: 1
```

### Install and start Server and check logs

```bash
curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.26.0+rke2r2 sudo -E sh -
sudo systemctl enable rke2-server.service
sudo systemctl start rke2-server.service
sudo journalctl -u rke2-server -f
```

### Add Kubernetes tools to path and set kubeconfig

Adds kubectl, crictl and ctr to path

```bash
echo 'export PATH="$PATH:/var/lib/rancher/rke2/bin"' >> ~/.bashrc
echo 'source <(kubectl completion bash)' >> ~/.bashrc
echo 'alias k=kubectl' >> ~/.bashrc
echo 'complete -o default -F __start_kubectl k' >>~/.bashrc
source ~/.bashrc
mkdir ~/.kube
ln -s /etc/rancher/rke2/rke2.yaml ~/.kube/config
```

### Get Token for Agent

```bash
sudo cat /var/lib/rancher/rke2/server/node-token
```

## RKE2 Agent Configuration (Optional)

```bash
sudo mkdir -p /etc/rancher/rke2
sudo vim /etc/rancher/rke2/config.yaml
```

```yaml
# /etc/rancher/rke2/config.yaml
---
server: https://<server>:9345
token: <token from server node>
```

### Install and start Agent and check logs

```bash
curl -sfL https://get.rke2.io | INSTALL_RKE2_TYPE="agent" INSTALL_RKE2_VERSION=v1.26.0+rke2r2 sudo -E sh -
sudo systemctl enable rke2-agent.service
sudo systemctl start rke2-agent.service
sudo journalctl -u rke2-agent -f
```

## Install Cluster Network Addons Operator

The [Cluster Network Addons Operator](https://github.com/kubevirt/cluster-network-addons-operator) can be used to deploy additional networking components.
Multus and Cilium are already installed via RKE2.
Open vSwitch CNI Plugin can be installed via this operator.

First install the operator itself:

```bash
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.85.0/namespace.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.85.0/network-addons-config.crd.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.85.0/operator.yaml
```

Then you need to create a configuration for the operator example CR:

```bash
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.85.0/network-addons-config-example.cr.yaml
```

Wait until the operator has finished the installation:

```bash
kubectl wait networkaddonsconfig cluster --for condition=Available
```

## Kubevirt

[Kubevirt](https://kubevirt.io/) is a Kubernetes add-on to run virtual machines.

### Validate Hardware Virtualization Support

```bash
sudo apt install libvirt-clients
sudo virt-host-validate qemu
```

### Install Kubevirt

Latest Release: ` export RELEASE=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt) `

```bash
export RELEASE=v0.58.1
# Deploy the KubeVirt operator
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml
# Create the KubeVirt CR (instance deployment request) which triggers the actual installation
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-cr.yaml
# wait until all KubeVirt components are up
kubectl -n kubevirt wait kv kubevirt --for condition=Available
```

### Install Containerized Data Importer

```bash
export CDI_VERSION=v1.55.2
kubectl create ns cdi
kubectl -n cdi apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$CDI_VERSION/cdi-operator.yaml
kubectl -n cdi apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$CDI_VERSION/cdi-cr.yaml
```

### Install virtctl via Krew

First install Krew and then install virtctl via Krew

```bash
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)
echo 'export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
kubectl krew install virt
kubectl virt help
```

## MetalLB

[MetalLB](https://metallb.universe.tf/) is a load-balancer implementation for bare metal Kubernetes clusters, using standard routing protocols.

### Install Operator Lifecycle Manager (OLM)

Install Operator Lifecycle Manager (OLM), a tool to help manage the Operators running on your cluster.

```bash
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.24.0/install.sh | bash -s v0.24.0
```

Install the operator by running the following command:

```bash
kubectl create -f https://operatorhub.io/install/metallb-operator.yaml
```

This Operator will be installed in the "operators" namespace and will be usable from all namespaces in the cluster.

After install, watch your operator come up using next command.

```bash
kubectl get csv -n operators
```

Now create a MetalLB IPAddressPool CR to configure the IP address range that MetalLB will use:

```bash
sudo vim metallb-ipaddresspool.yaml
```

```yaml
# metallb-ipaddresspool.yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: default
  namespace: operators
spec:
  addresses:
  - X.X.X.X/XX
```

Create a L2Advertisement to tell MetalLB to responde to ARP requests for all IP address pools (no named ip address pool, means all pools):

```bash
sudo vim l2advertisment.yaml
```

```yaml
# l2advertisment.yaml
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: default
  namespace: operators
spec:
  ipAddressPools:
  - default
```

Apply the configuration:

```bash
kubectl apply -f metallb-ipaddresspool.yaml
kubectl apply -f l2advertisment.yaml
```

## Trident

Check connectivity to NetApp Storage:

```bash
kubectl run -i --tty ping --image=busybox --restart=Never --rm -- \
  ping <NetApp Management IP>
```

Download and extract the Trident installer:

```bash
export TRIDENT_VERSION=23.01.0
wget https://github.com/NetApp/trident/releases/download/v$TRIDENT_VERSION/trident-installer-$TRIDENT_VERSION.tar.gz
tar -xf trident-installer-$TRIDENT_VERSION.tar.gz
cd trident-installer
mkdir setup
vim ./setup/backend.json
```

Configure the installer:

```bash
# ./backend.json

```
