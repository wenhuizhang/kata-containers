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

cd ~/kata-containers/tools/osbuilder/image-builder
./image_builder.sh ../rootfs-builder/rootfs-ubuntu

cd ~/kata-containers/tools/osbuilder/initrd-builder
./initrd_builder.sh ../rootfs-builder/rootfs-ubuntu/
```


### 4.2 install two images

1. Install rootfs image

```
cd ~/kata-containers/tools/osbuilder/image-builder

commit=$(git log --format=%h -1 HEAD)
date=$(date +%Y-%m-%d-%T.%N%z)
image="kata-containers-${date}-${commit}"

sudo install -o root -g root -m 0640 -D ./kata-containers.img "/usr/share/kata-containers/${image}"

cd /usr/share/kata-containers
sudo ln -sf "$image" kata-containers.img

root@n223-247-005:/usr/share/kata-containers# ls /usr/share/kata-containers/
kata-containers-2023-01-31-06:04:20.789226969+0800-e6dbe0a9a  kata-containers.img
```

2. Install initrd image

depends on agent (located at "/usr/bin/kata-agent")

```
cd ~/kata-containers/tools/osbuilder/initrd-builder

sudo install -o root -g root -m 0550 -T ../rootfs-builder/rootfs-ubuntu/usr/bin/kata-agent ../rootfs-builder/rootfs-ubuntu/sbin/init
cp /usr/bin/kata-agent ../rootfs-builder/rootfs-ubuntu/usr/bin/kata-agent

commit=$(git log --format=%h -1 HEAD)
date=$(date +%Y-%m-%d-%T.%N%z)
image="kata-containers-initrd-${date}-${commit}"

sudo install -o root -g root -m 0640 -D kata-containers-initrd.img "/usr/share/kata-containers/${image}"

cd /usr/share/kata-containers 
sudo ln -sf "$image" kata-containers-initrd.img

root@n223-247-005:/usr/share/kata-containers# ls /usr/share/kata-containers/
kata-containers-2023-01-31-06:04:20.789226969+0800-e6dbe0a9a
kata-containers.img
kata-containers-initrd-2023-02-01-06:56:17.056503879+0800-ea1c79a24
kata-containers-initrd.img

```

## 5. Build and Install Kernel

1. setup kernel

```
cd ~/kata-containers/tools/packaging/kernel

./build-kernel.sh -v 5.15 -g intel -f -d setup

Note
-v 5.15: Specify the guest kernel version.
-b: To enable BPF related features in a guest kernel.
-g nvidia/intel: To build a guest kernel supporting Nvidia/Intel GPU.
-f: The .config file is forced to be generated even if the kernel directory already exists.
-d: Enable bash debug mode.
```


2. change kernel config

```
cd ~/kata-containers/tools/packaging/kernel

cd kata-linux--f-5.15-*

make menuconfig

cd ..
./build-kernel.sh -v 5.15 -b  -g intel -f -d build
```

```
...
Kernel: arch/x86/boot/bzImage is ready  (#1)
 Line 412: '[' '' == sev ']'
 Line 415: '[' x86_64 '!=' powerpc ']'
 Line 415: '[' -e arch/x86_64/boot/bzImage ']'
 Line 416: '[' -e vmlinux ']'
 Line 417: '[' '' == firecracker ']'
 Line 417: '[' '' == cloud-hypervisor ']'
 Line 418: popd
```

```
root@n223-247-005:~/qemu-6.2.0# ls ~/kata-containers/tools/packaging/kernel/kata-linux-5.15-98/
arch     crypto         init     lib          modules.builtin          scripts     usr
block    Documentation  ipc      LICENSES     modules.builtin.modinfo  security    virt
certs    drivers        Kbuild   MAINTAINERS  net                      sound       vmlinux
COPYING  fs             Kconfig  Makefile     README                   System.map  vmlinux.o
CREDITS  include        kernel   mm           samples                  tools       vmlinux.symvers

```

3. install kernel

Installation default locations: `/usr/share/kata-containers/`

```
./build-kernel.sh -v 5.15 -b   -g intel -f -d install
```


log
```
root@n223-247-005:~/kata-containers/tools/packaging/kernel# ./build-kernel.sh -v 5.15 -g intel -f -d install
 Line 479: getopts a:b:c:deEfg:hk:p:t:u:v:x: opt
 Line 538: shift 6
 Line 540: subcmd=install
 Line 542: '[' -z install ']'
 Line 544: [[ '' == \e\x\p\e\r\i\m\e\n\t\a\l ]]
 Line 555: '[' -z 5.15 ']'
 Line 580: kernel_version=5.15
 Line 582: '[' -z '' ']'
  Line 583: get_config_version
  Line 117: get_config_and_patches
  Line 111: '[' -z '' ']'
  Line 112: patches_path=/root/kata-containers/tools/packaging/kernel/patches
  Line 118: config_version_file=/root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
  Line 119: '[' -f /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version ']'
  Line 120: cat /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
 Line 583: config_version=98
 Line 584: [[ '' != '' ]]
 Line 587: kernel_path=/root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 589: info 'Config version: 98'
 Line 60: echo 'INFO: Config version: 98'
INFO: Config version: 98
 Line 592: info 'Kernel version: 5.15'
 Line 60: echo 'INFO: Kernel version: 5.15'
INFO: Kernel version: 5.15
 Line 594: case "${subcmd}" in
 Line 599: build_kernel /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 405: local kernel_path=/root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 406: '[' -n /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
 Line 407: '[' -d /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
 Line 408: '[' -n '' ']'
  Line 408: uname -m
 Line 408: arch_target=x86_64
  Line 409: arch_to_kernel x86_64
  Line 113: local -r arch=x86_64
  Line 115: case "$arch" in
  Line 119: echo x86_64
 Line 409: arch_target=x86_64
 Line 410: pushd /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
  Line 411: nproc
 Line 411: make -j 8 ARCH=x86_64
  DESCEND objtool
  CALL    scripts/atomic/check-atomics.sh
  CALL    scripts/checksyscalls.sh
  CHK     include/generated/compile.h
Kernel: arch/x86/boot/bzImage is ready  (#1)
 Line 412: '[' '' == sev ']'
 Line 415: '[' x86_64 '!=' powerpc ']'
 Line 415: '[' -e arch/x86_64/boot/bzImage ']'
 Line 416: '[' -e vmlinux ']'
 Line 417: '[' '' == firecracker ']'
 Line 417: '[' '' == cloud-hypervisor ']'
 Line 418: popd
 Line 600: install_kata /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 422: local kernel_path=/root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 423: '[' -n /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
 Line 424: '[' -d /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
 Line 425: pushd /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
  Line 426: get_config_version
  Line 117: get_config_and_patches
  Line 111: '[' -z '' ']'
  Line 112: patches_path=/root/kata-containers/tools/packaging/kernel/patches
  Line 118: config_version_file=/root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
  Line 119: '[' -f /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version ']'
  Line 120: cat /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
 Line 426: config_version=98
 Line 427: '[' -n 98 ']'
  Line 428: readlink -m ///usr/share/kata-containers
 Line 428: install_path=/usr/share/kata-containers
 Line 430: suffix=
 Line 431: [[ '' != '' ]]
 Line 434: [[ intel != '' ]]
 Line 435: suffix=-intel-gpu
 Line 438: [[ '' != '' ]]
 Line 442: vmlinuz=vmlinuz-5.15-98-intel-gpu
 Line 443: vmlinux=vmlinux-5.15-98-intel-gpu
 Line 445: '[' -e arch/x86_64/boot/bzImage ']'
 Line 446: bzImage=arch/x86_64/boot/bzImage
 Line 454: '[' x86_64 = powerpc ']'
 Line 457: install --mode 0644 -D arch/x86_64/boot/bzImage /usr/share/kata-containers/vmlinuz-5.15-98-intel-gpu
 Line 461: '[' x86_64 = arm64 ']'
 Line 463: '[' x86_64 = s390 ']'
 Line 466: install --mode 0644 -D vmlinux /usr/share/kata-containers/vmlinux-5.15-98-intel-gpu
 Line 469: install --mode 0644 -D ./.config /usr/share/kata-containers/config-5.15
 Line 471: ln -sf vmlinuz-5.15-98-intel-gpu /usr/share/kata-containers/vmlinuz-intel-gpu.container
 Line 472: ln -sf vmlinux-5.15-98-intel-gpu /usr/share/kata-containers/vmlinux-intel-gpu.container
 Line 473: ls -la /usr/share/kata-containers/vmlinux-intel-gpu.container
lrwxrwxrwx 1 root root 25 Feb  1 08:09 /usr/share/kata-containers/vmlinux-intel-gpu.container -> vmlinux-5.15-98-intel-gpu
 Line 474: ls -la /usr/share/kata-containers/vmlinuz-intel-gpu.container
lrwxrwxrwx 1 root root 25 Feb  1 08:09 /usr/share/kata-containers/vmlinuz-intel-gpu.container -> vmlinuz-5.15-98-intel-gpu
```

verify 

```
root@n223-247-005:~/kata-containers/tools/packaging/kernel# ls /usr/share/kata-containers/
config-5.15                                                          vmlinux-5.15-98-intel-gpu
kata-containers-2023-01-31-06:04:20.789226969+0800-e6dbe0a9a         vmlinux-intel-gpu.container
kata-containers.img                                                  vmlinuz-5.15-98-intel-gpu
kata-containers-initrd-2023-02-01-06:56:17.056503879+0800-ea1c79a24  vmlinuz-intel-gpu.container
kata-containers-initrd.img


```




## 6. Setup kata-containers

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


