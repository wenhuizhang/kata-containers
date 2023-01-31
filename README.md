<img src="https://www.openstack.org/assets/kata/kata-vertical-on-white.png" width="150">

# Kata Containers


Kata Containers is an open source project and community working to build a
standard implementation of lightweight Virtual Machines (VMs) that feel and
perform like containers, but provide the workload isolation and security
advantages of VMs.

Host artifacts for Kata

* `cloud-hypervisor`, `firecracker`, `qemu`, and supporting binaries
* `containerd-shim-kata-v2` (go runtime and rust runtime)
* `kata-collect-data.sh`
* `kata-runtime`


Virtual Machine artifacts for Kata

* `kata-containers.img` and `kata-containers-initrd.img`
* `vmlinuz.container` and `vmlinuz-virtiofs.container`


## 1. Hardware requirements

The [Kata Containers runtime](src/runtime) provides a command to
determine if your host system is capable of running and creating a
Kata Container:

```bash
$ kata-runtime check
```


## 2. Getting started

### 1. install dependencies 


1. general dependencies 

```
apt-get install netbase libnftnl11

apt-get install -y --no-install-recommends \
	    bc \
	    bison \
	    build-essential \
	    ca-certificates \
	    curl \
	    flex \
	    git \
	    iptables \
	    libelf-dev \
	    patch
```


2. install golang
```
wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
rm -rf /usr/bin/go && tar -C /usr/bin -xzf go1.19.5.linux-amd64.tar.gz 

go version

vim ~/.bashrc 
export PATH=$PATH:/usr/bin/go/bin
export GOPATH=/usr/bin/go
source ~/.bashrc
```



3. install protobuf and seccomp support

```
apt install -y protobuf-compiler
apt-get install -y seccomp
apt -y install libseccomp-dev

```

### 2. Compile and Install Kata



1. Install libs and agent
```
# libs 
cd ~/kata-containers/src/libs
make 
make install


# Kata runtime
cd ~/kata-containers/src/runtime
make 
make install

# Kata agent 
cd ~/kata-containers/src/agent
make 
make install
```



2. Install kata runtime explicity
```
cd ~/kata-containers/src/runtime
make && sudo -E PATH=$PATH make install

mkdir -p /usr/local/bin/data/
cp /usr/local/bin/kata-collect-data.sh /usr/local/bin/data/kata-collect-data.sh
```
This installation creates the following
```
runtime binary: /usr/local/bin/kata-runtime
configuration file: /usr/share/defaults/kata-containers/configuration.toml
```

Install location

```
	binary installation path (BINDIR) : /usr/local/bin
	binaries to install :
	 - /usr/local/bin/kata-runtime
	 - /usr/local/bin/containerd-shim-kata-v2
	 - /usr/local/bin/kata-monitor
	 - /usr/local/bin/data/kata-collect-data.sh
	configs to install (CONFIGS) :
	 - config/configuration-acrn.toml
 	 - config/configuration-clh-tdx.toml
 	 - config/configuration-clh.toml
 	 - config/configuration-fc.toml
 	 - config/configuration-qemu-sev.toml
 	 - config/configuration-qemu-tdx.toml
 	 - config/configuration-qemu.toml
	install paths (CONFIG_PATHS) :
	 - /usr/share/defaults/kata-containers/configuration-acrn.toml
 	 - /usr/share/defaults/kata-containers/configuration-clh-tdx.toml
 	 - /usr/share/defaults/kata-containers/configuration-clh.toml
 	 - /usr/share/defaults/kata-containers/configuration-fc.toml
 	 - /usr/share/defaults/kata-containers/configuration-qemu-sev.toml
 	 - /usr/share/defaults/kata-containers/configuration-qemu-tdx.toml
 	 - /usr/share/defaults/kata-containers/configuration-qemu.toml
	alternate config paths (SYSCONFIG_PATHS) : 
	 - /etc/kata-containers/configuration-acrn.toml
 	 - /etc/kata-containers/configuration-clh-tdx.toml
 	 - /etc/kata-containers/configuration-clh.toml
 	 - /etc/kata-containers/configuration-fc.toml
 	 - /etc/kata-containers/configuration-qemu-sev.toml
 	 - /etc/kata-containers/configuration-qemu-tdx.toml
 	 - /etc/kata-containers/configuration-qemu.toml
	default install path for qemu (CONFIG_PATH) : /usr/share/defaults/kata-containers/configuration.toml
	default alternate config path (SYSCONFIG) : /etc/kata-containers/configuration.toml
	qemu hypervisor path (QEMUPATH) : /usr/bin/qemu-system-x86_64
	cloud-hypervisor hypervisor path (CLHPATH) : /usr/bin/cloud-hypervisor
	firecracker hypervisor path (FCPATH) : /usr/bin/firecracker
	acrn hypervisor path (ACRNPATH) : /usr/bin/acrn-dm
	assets path (PKGDATADIR) : /usr/share/kata-containers
	shim path (PKGLIBEXECDIR) : /usr/libexec/kata-containers

```



3. Other installations
```
# Kata RunD
cd ~/kata-containers/src/rumtime-rs
make 
make install

# Kata Dragonball
cd ~/kata-containers/src/dragonball
make 
make install

```

## 3. Build QEMU
```
git clone https://github.com/wenhuizhang/qemu-6.2.0.git
cd qemu-6.2.0

root@n223-247-005:~/qemu-6.2.0# ~/kata-containers/tools/packaging/scripts/configure-hypervisor.sh qemu > kata.cfg
rm -rf ./build
eval ./configure "$(cat kata.cfg)"
make -j $(nproc)
sudo -E make install
```

## 4. Build rootfs and images 

### 4.1 build rootfs and images
```
cd ~/kata-containers/tools/osbuilder/rootfs-builder
./rootfs.sh ubuntu
ls ./rootfs-ubuntu

cd ~/kata-containers/tools/osbuilder/initrd-builder
./initrd_builder.sh ../rootfs-builder/rootfs-ubuntu/

cd ~/kata-containers/tools/osbuilder/image-builder
./image_builder.sh ../rootfs-builder/rootfs-ubuntu
```


### 4.2 install images

Install the rootfs image

```
cd ~/kata-containers/tools/osbuilder/image-builder

commit=$(git log --format=%h -1 HEAD)
date=$(date +%Y-%m-%d-%T.%N%z)
image="kata-containers-${date}-${commit}"

sudo install -o root -g root -m 0640 -D ./kata-containers.img "/usr/share/kata-containers/${image}"

cd /usr/share/defaults/kata-containers
sudo ln -sf "$image" kata-containers.img

root@n223-247-005:/usr/share/kata-containers# ls /usr/share/kata-containers/
kata-containers-2023-01-31-06:04:20.789226969+0800-e6dbe0a9a  kata-containers.img
```

## 5. Setup kata-containers

1. sync config files
```
mkdir -p /etc/kata-containers
cp -R /usr/share/defaults/kata-containers/* /etc/kata-containers/


root@n223-247-005:~/kata-containers# ls -lash /usr/share/defaults/kata-containers/
total 176K
4.0K drwxr-xr-x 2 root root 4.0K Feb  1 05:20 .
4.0K drwxr-xr-x 4 root root 4.0K Dec  3 06:05 ..
 12K -rw-r--r-- 1 root root 9.6K Jan 31 03:14 configuration-acrn.toml
 20K -rw-r--r-- 1 root root  19K Jan 31 03:14 configuration-clh-tdx.toml
 20K -rw-r--r-- 1 root root  19K Jan 31 03:14 configuration-clh.toml
 12K -rw-r--r-- 1 root root  11K Jan 31 03:23 configuration-dragonball.toml
 16K -rw-r--r-- 1 root root  16K Jan 31 03:14 configuration-fc.toml
 28K -rw-r--r-- 1 root root  28K Jan 31 03:14 configuration-qemu-sev.toml
 28K -rw-r--r-- 1 root root  28K Jan 31 03:14 configuration-qemu-tdx.toml
 32K -rw-r--r-- 1 root root  29K Jan 31 03:14 configuration-qemu.toml
   0 lrwxrwxrwx 1 root root   59 Feb  1 05:20 configuration.toml -> /usr/share/defaults/kata-containers/configuration-qemu.toml

```


2. edit default config file base
```
rm -f /usr/share/defaults/kata-containers/configuration.toml
ln -s /usr/share/defaults/kata-containers/configuration-qemu.toml /usr/share/defaults/kata-containers/configuration.toml

rm -f /etc/kata-containers/configuration.toml
ln -s /etc/kata-containers/configuration-qemu.toml  /etc/kata-containers/configuration.toml
```

3. change configuration.toml
```
kernel = "/usr/share/kata-containers/vmlinuz-5.15"
image = "/usr/share/kata-containers/kata-containers.img"
enable_iommu = true
hotplug_vfio_on_root_bus = true
pcie_root_port = 1
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


