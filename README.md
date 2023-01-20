<img src="https://www.openstack.org/assets/kata/kata-vertical-on-white.png" width="150">

# Kata Containers



## Introduction

Kata Containers is an open source project and community working to build a
standard implementation of lightweight Virtual Machines (VMs) that feel and
perform like containers, but provide the workload isolation and security
advantages of VMs.


### Hardware requirements

The [Kata Containers runtime](src/runtime) provides a command to
determine if your host system is capable of running and creating a
Kata Container:

```bash
$ kata-runtime check
```


## Getting started

1. install dependencies 
```
wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
rm -rf /usr/bin/go && tar -C /usr/bin -xzf go1.19.5.linux-amd64.tar.gz 

go version

vim ~/.bashrc 
export PATH=$PATH:/usr/bin/go/bin
export GOPATH=/usr/bin/go
source ~/.bashrc
```


## Documentation

See the [official documentation](docs) including:

- [Installation guides](docs/install)
- [Developer guide](docs/Developer-Guide.md)
- [Design documents](docs/design)
  - [Architecture overview](docs/design/architecture)
  - [Architecture 3.0 overview](docs/design/architecture_3.0/)

## Configuration

Kata Containers uses a single
[configuration file](src/runtime/README.md#configuration)
which contains a number of sections for various parts of the Kata
Containers system including the [runtime](src/runtime), the
[agent](src/agent) and the [hypervisor](#hypervisors).

## Hypervisors

See the [hypervisors document](docs/hypervisors.md) and the
[Hypervisor specific configuration details](src/runtime/README.md#hypervisor-specific-configuration).


### Main components

The table below lists the core parts of the project:

| Component | Type | Description |
|-|-|-|
| [runtime](src/runtime) | core | Main component run by a container manager and providing a containerd shimv2 runtime implementation. |
| [runtime-rs](src/runtime-rs) | core | The Rust version runtime. |
| [agent](src/agent) | core | Management process running inside the virtual machine / POD that sets up the container environment. |
| [`dragonball`](src/dragonball) | core | An optional built-in VMM brings out-of-the-box Kata Containers experience with optimizations on container workloads |
| [documentation](docs) | documentation | Documentation common to all components (such as design and install documentation). |
| [tests](https://github.com/kata-containers/tests) | tests | Excludes unit tests which live with the main code. |

### Additional components

The table below lists the remaining parts of the project:

| Component | Type | Description |
|-|-|-|
| [packaging](tools/packaging) | infrastructure | Scripts and metadata for producing packaged binaries<br/>(components, hypervisors, kernel and rootfs). |
| [kernel](https://www.kernel.org) | kernel | Linux kernel used by the hypervisor to boot the guest image. Patches are stored [here](tools/packaging/kernel). |
| [osbuilder](tools/osbuilder) | infrastructure | Tool to create "mini O/S" rootfs and initrd images and kernel for the hypervisor. |
| [`agent-ctl`](src/tools/agent-ctl) | utility | Tool that provides low-level access for testing the agent. |
| [`kata-ctl`](src/tools/kata-ctl) | utility | Tool that provides advanced commands and debug facilities. |
| [`trace-forwarder`](src/tools/trace-forwarder) | utility | Agent tracing helper. |
| [`runk`](src/tools/runk) | utility | Standard OCI container runtime based on the agent. |
| [`ci`](https://github.com/kata-containers/ci) | CI | Continuous Integration configuration files and scripts. |
| [`katacontainers.io`](https://github.com/kata-containers/www.katacontainers.io) | Source for the [`katacontainers.io`](https://www.katacontainers.io) site. |


