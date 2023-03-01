# Kata QEMU Installation


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


2. install golang and rust

golang
```
wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
rm -rf /usr/bin/go && tar -C /usr/bin -xzf go1.19.5.linux-amd64.tar.gz 

go version

vim ~/.bashrc 
export PATH=$PATH:/usr/bin/go/bin
export GOPATH=/usr/bin/go
source ~/.bashrc
```
rust
```
curl https://sh.rustup.rs -sSf | sh
vim ~/.bashrc 
export PATH=$PATH:/usr/bin/go/bin:$HOME/.cargo/bin
. "$HOME/.cargo/env"
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
sudo kata-collect-data.sh > /tmp/kata.log
cd ~/kata-containers/src/agent
sudo apt install musl-tools
arch=$(uname -m)
rustup target add "${arch}-unknown-linux-musl"
sudo ln -s /usr/bin/g++ /bin/musl-g++
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
sudo apt-get install multistrap
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

## 5. Build and Install Kernel vmlinux-container

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

mv kata-linux-*** kata-linux--***

//make menuconfig

./build-kernel.sh build -v 5.15 -b  -g intel -f -d 
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
./build-kernel.sh install -v 5.15 -b   -g intel -f -d 
```


log
```
root@n223-247-005:~/kata-containers/tools/packaging/kernel# ./build-kernel.sh -v 5.15 -g intel -f -d install
...
  Line 112: patches_path=/root/kata-containers/tools/packaging/kernel/patches
  Line 118: config_version_file=/root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
  Line 119: '[' -f /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version ']'
  Line 120: cat /root/kata-containers/tools/packaging/kernel/patches/../kata_config_version
 Line 583: config_version=98
 Line 584: [[ '' != '' ]]
 Line 587: kernel_path=/root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
...
 Line 594: case "${subcmd}" in
 Line 599: build_kernel /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 405: local kernel_path=/root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98
 Line 406: '[' -n /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
 Line 407: '[' -d /root/kata-containers/tools/packaging/kernel/kata-linux-5.15-98 ']'
...
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
...
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

## 7. Test with QEMU

You need a machine with KVM to test.
```
egrep -c '(vmx|svm)' /proc/cpuinfo
2

sudo kvm-ok
sudo apt install cpu-checker
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils
```


virtio setup : https://virtio-fs.gitlab.io/howto-qemu.html

```
apt-get install dracut live-tools
lsinitrd /boot/initramfs-$(uname -r).img | grep virtio
lsinitrd /boot/initrd.img-5.15.56.bsk.1-amd64 | grep virtio

vim /etc/dracut.conf
add_drivers+="virtio_blk virtio_scsi virtio_net virtio_pci virtio_ring virtio"

dracut -f

vim /etc/initramfs-tools/modules
virtio_blk
virtio_scsi
virtio_net
virtio_pci
virtio_ring
virtio


update-initramfs -u
lsinitrd /boot/initramfs-$(uname -r).img | grep virtio
lsinitrd /boot/initrd.img-5.15.56.bsk.1-amd64 | grep virtio

# https://gitlab.com/virtio-fs/virtiofsd
ls /usr/libexec/virtiofsd
```

To use virtio-mem
```
echo 1 > /proc/sys/vm/overcommit_memory
```


setup the containerd
```
containerd config default > /etc/containerd/config.toml
vim config.toml 

# add kata runtime option after the runtimes
# [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
  runtime_type = "io.containerd.kata.v2"


systemctl restart containerd
```


- Sample containerd configuration

We need to change configuration of containerd, a sample is shown below. 

```
disabled_plugins = []
imports = []
oom_score = 0
plugin_dir = ""
required_plugins = []
root = "/var/lib/containerd"
state = "/run/containerd"
version = 2

[cgroup]
  path = ""

[debug]
  address = ""
  format = ""
  gid = 0
  level = ""
  uid = 0

[grpc]
  address = "/run/containerd/containerd.sock"
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216
  tcp_address = ""
  tcp_tls_cert = ""
  tcp_tls_key = ""
  uid = 0

[metrics]
  address = ""
  grpc_histogram = false

[plugins]

  [plugins."io.containerd.gc.v1.scheduler"]
    deletion_threshold = 0
    mutation_threshold = 100
    pause_threshold = 0.02
    schedule_delay = "0s"
    startup_delay = "100ms"

  [plugins."io.containerd.grpc.v1.cri"]
    disable_apparmor = false
    disable_cgroup = false
    disable_hugetlb_controller = true
    disable_proc_mount = false
    disable_tcp_service = true
    enable_selinux = false
    enable_tls_streaming = false
    ignore_image_defined_volumes = false
    max_concurrent_downloads = 3
    max_container_log_line_size = 16384
    netns_mounts_under_state_dir = false
    restrict_oom_score_adj = false
    sandbox_image = "k8s.gcr.io/pause:3.5"
    selinux_category_range = 1024
    stats_collect_period = 10
    stream_idle_timeout = "4h0m0s"
    stream_server_address = "127.0.0.1"
    stream_server_port = "0"
    systemd_cgroup = false
    tolerate_missing_hugetlb_controller = true
    unset_seccomp_profile = ""

    [plugins."io.containerd.grpc.v1.cri".cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
      conf_template = ""
      max_conf_num = 1

    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      disable_snapshot_annotations = true
      discard_unpacked_layers = false
      no_pivot = false
      snapshotter = "overlayfs"

      [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime]
        base_runtime_spec = ""
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        runtime_engine = ""
        runtime_root = ""
        runtime_type = ""

        [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = ""
          container_annotations = []
          pod_annotations = []
          privileged_without_host_devices = false
          runtime_engine = ""
          runtime_root = ""
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            BinaryName = ""
            CriuImagePath = ""
            CriuPath = ""
            CriuWorkPath = ""
            IoGid = 0
            IoUid = 0
            NoNewKeyring = false
            NoPivotRoot = false
            Root = ""
            ShimCgroup = ""
            SystemdCgroup = false
         [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
            runtime_type = "io.containerd.kata.v2"

      [plugins."io.containerd.grpc.v1.cri".containerd.untrusted_workload_runtime]
        base_runtime_spec = ""
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        runtime_engine = ""
        runtime_root = ""
        runtime_type = ""

        [plugins."io.containerd.grpc.v1.cri".containerd.untrusted_workload_runtime.options]

    [plugins."io.containerd.grpc.v1.cri".image_decryption]
      key_model = "node"

    [plugins."io.containerd.grpc.v1.cri".registry]
      config_path = ""

      [plugins."io.containerd.grpc.v1.cri".registry.auths]

      [plugins."io.containerd.grpc.v1.cri".registry.configs]

      [plugins."io.containerd.grpc.v1.cri".registry.headers]

      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]

    [plugins."io.containerd.grpc.v1.cri".x509_key_pair_streaming]
      tls_cert_file = ""
      tls_key_file = ""

  [plugins."io.containerd.internal.v1.opt"]
    path = "/opt/containerd"

  [plugins."io.containerd.internal.v1.restart"]
    interval = "10s"

  [plugins."io.containerd.metadata.v1.bolt"]
    content_sharing_policy = "shared"

  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false

  [plugins."io.containerd.runtime.v1.linux"]
    no_shim = false
    runtime = "runc"
    runtime_root = ""
    shim = "containerd-shim"
    shim_debug = false

  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]

  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]

  [plugins."io.containerd.snapshotter.v1.aufs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.devmapper"]
    async_remove = false
    base_image_size = ""
    pool_name = ""
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.native"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.overlayfs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.zfs"]
    root_path = ""

[proxy_plugins]

[stream_processors]

  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+encrypted"]
    args = ["--decryption-keys-path", "/etc/containerd/ocicrypt/keys"]
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/etc/containerd/ocicrypt/ocicrypt_keyprovider.conf"]
    path = "ctd-decoder"
    returns = "application/vnd.oci.image.layer.v1.tar"

  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar.gzip"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+gzip+encrypted"]
    args = ["--decryption-keys-path", "/etc/containerd/ocicrypt/keys"]
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/etc/containerd/ocicrypt/ocicrypt_keyprovider.conf"]
    path = "ctd-decoder"
    returns = "application/vnd.oci.image.layer.v1.tar+gzip"

[timeouts]
  "io.containerd.timeout.shim.cleanup" = "5s"
  "io.containerd.timeout.shim.load" = "5s"
  "io.containerd.timeout.shim.shutdown" = "3s"
  "io.containerd.timeout.task.state" = "2s"

[ttrpc]
  address = ""
  gid = 0
  uid = 0
```

pull the image and test 

```
ctr image pull docker.io/library/busybox:latest
ctr run --runtime io.containerd.run.kata.v2 -t --rm docker.io/library/busybox:latest hello sh
ctr run --runtime=io.containerd.kata.v2 --rm docker.io/library/busybox:latest kata-test uname -r

```

socket location 
```
root@n192-191-015:~/kata-containers# ls /run/containerd/containerd.sock
/run/containerd/containerd.sock
```

