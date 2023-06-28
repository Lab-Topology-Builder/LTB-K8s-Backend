# Comparison to other similar projects

The Lab Topology Builder is just one among many open-source projects available for building emulated network topologies.
The aim of this project comparison is to provide a concise overview of the key features offered by some of the most well-known projects, assisting users in selecting the optimal solution for their use case.

## vrnetlab - VR Network Lab

[vrnetlab](https://github.com/vrnetlab/vrnetlab) is a network emulator that runs virtual routers using KVM and Docker. It is similar to the KVM/Docker-based LTB, but is more simple and only provides the deployment functionality.

## Containerlab

[Containerlab](https://github.com/srl-labs/containerlab) is a tool to deploy network labs with virtual routers, firewalls, load balancers, and more, using Docker containers. It is based on [vrnetlab](#vrnetlab---vr-network-lab) and provides a declarative way to define the lab topology using a YAML file.
Containerlab is not capable of deploying lab topologies over multiple host nodes, which is a key feature that the K8s-based LTB aims to provide in the future.

## Netlab

[Netlab](https://github.com/ipspace/netlab) is an abstraction layer to deploy network topologies based on [containerlab](#containerlab) or Vagrant. It provides a declarative way to define the lab topology using a YAML file.
It mainly provides an abstracted way to define labs topologies with preconfigured lab nodes.

## Kubernetes Network Emulator

[Kubernetes Network Emulator](https://github.com/openconfig/kne) is a network emulator that aims to provide a standard interface so that vendors can produce a standard container implementation of their network operating system that can be used in a network emulation environment.
Currently, it currently does not seem to support many network operating systems and additional operators are required to support different vendors.

## Mininet

[Minitnet](https://github.com/mininet/mininet) is a network emulator that runs a collection of end-hosts, switches, routers, and links on a single Linux kernel. It is mainly used for testing SDN controllers and can not deploy a lab with a specific vendor image.

## GNS3

[GNS3](https://www.gns3.com/software) is a network emulator that can run network devices as virtual machines or Docker containers.
It primarily focuses on providing a emulated network environment for a single user and its deployment and usage can be quite complex.
Additionally, it does not provide a way to scale labs over multiple host nodes.
