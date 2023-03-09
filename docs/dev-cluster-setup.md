# Development cluster setup

Setup for a development Kubernetes cluster.

## Prerequisites

WSL version 0.67.6 and higher

Active Ubuntu WSL needed:

```bash
sudo vim /etc/wsl.conf
```

```bash
# /etc/wsl.conf
[boot]
systemd=true
```

Restart WSL with `wsl.exe --shutdown` and opening the WSL again.

## Option 1: K3S install

```bash
sudo mkdir -p /etc/rancher/k3s/
sudo vim /etc/rancher/k3s/config.yaml
```

```yaml
# /etc/rancher/k3s/config.yaml
write-kubeconfig-mode: "0644"
```

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.26.1+k3s1" sudo -E sh -
ln -s /etc/rancher/k3s/k3s.yaml ~/.kube/k3s.yaml
echo "export KUBECONFIG=${KUBECONFIG}:${HOME}/.kube/k3s.yaml" >> ~/.bashrc
```

## Option 2: K3D install

Prerequisites:
Docker Desktop installed and running or Docker installed in WSL

```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.4.7 bash
```

## Option 3: K0s install

```bash
curl -sSfL https://get.k0s.sh | K0S_VERSION=v1.26.1+k0s.0 sudo -E sh
```

## Operator SDK install

### [Prerequisites](https://v1-11-x.sdk.operatorframework.io/docs/installation/#prerequisites)

- [curl](https://curl.haxx.se/)
- [gpg](https://gnupg.org/) version 2.0+

### 1. Download the release binary

Set platform information:

```sh
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
```

Download the binary for your platform:

```sh
export OPERATOR_SDK_Version=1.26.1
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_Version}
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
```

### 2. [Verify the downloaded binary](https://v1-11-x.sdk.operatorframework.io/docs/installation/#2-verify-the-downloaded-binary) (Optional)

Import the operator-sdk release GPG key from `keyserver.ubuntu.com`:

```sh
gpg --keyserver keyserver.ubuntu.com --recv-keys 052996E2A20B5C7E
```

Download the checksums file and its signature, then verify the signature (optional):

```sh
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt.asc
gpg -u "Operator SDK (release) <cncf-operator-sdk@cncf.io>" --verify checksums.txt.asc
```

You should see something similar to the following:

```console
gpg: assuming signed data in 'checksums.txt'
gpg: Signature made Fri 30 Oct 2020 12:15:15 PM PDT
gpg:                using RSA key ADE83605E945FA5A1BD8639C59E5B47624962185
gpg: Good signature from "Operator SDK (release) <cncf-operator-sdk@cncf.io>" [ultimate]
```

Make sure the checksums match:

```sh
grep operator-sdk_${OS}_${ARCH} checksums.txt | sha256sum -c -
```

You should see something similar to the following:

```console
operator-sdk_linux_amd64: OK
```

### 3. [Install the release binary in your PATH](https://v1-11-x.sdk.operatorframework.io/docs/installation/#3-install-the-release-binary-in-your-path)

```sh
chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

Verify the installation:

```sh
operator-sdk version
```

## Install Go

Step 1 - Downloading Go lang binary files
Visit official downloads page and grab file using either wget command or curl command:

```bash
# let us download a file with curl on Linux command line #
GO_VERSION="1.20.1" # go version
ARCH="amd64" # go archicture
wget -L "https://golang.org/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz
```

Step 2  - Add to PATH

```bash
echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
source ~/.bashrc
```

Step 3 - Verify that you've installed Go by opening a command prompt and typing the following command:

```bash
go version
```
