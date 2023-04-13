# Firecracker with Kata

1. Install firecracker and jailer

To install Firecracker we need to get the firecracker and jailer binaries:

```
$ release_url="https://github.com/firecracker-microvm/firecracker/releases"
$ version="v0.23.1"
$ arch=`uname -m`
$ curl ${release_url}/download/${version}/firecracker-${version}-${arch} -o firecracker
$ curl ${release_url}/download/${version}/jailer-${version}-${arch} -o jailer
$ chmod +x jailer firecracker
```

Move them to /usr/local/bin:
```

$ cp $(pwd)/firecracker /usr/local/bin
$ cp $(pwd)/jailer /usr/local/bin
```


2. Install device mapper
```
$ ctr plugins ls |grep devmapper
io.containerd.snapshotter.v1    devmapper                linux/amd64    ok
```


Stop Docker.

```
 sudo systemctl stop docker
```
 
 
Edit /etc/docker/daemon.json. 
```

root@n192-191-015:~/opt/kata/share/defaults/kata-containers# cat /etc/docker/daemon.json
{
    "runtimes": {
        "runsc": {
            "path": "/root/gvisor/bin/runsc"
        },
        "runsc-debug": {
            "path": "/usr/local/bin/runsc",
            "runtimeArgs": [
                "--debug",
                "--debug-log=/tmp/runsc-debug.log",
                "--strace",
                "--log-packets"
            ]
        },
	"kata": {
	     "path": "/usr/local/bin/kata-runtime"
        }
    },
  "storage-driver": "devicemapper"
}
```


```
DATA_DIR=/var/lib/containerd/devmapper
POOL_NAME=devpool

mkdir -p ${DATA_DIR}

# Create data file
sudo touch "${DATA_DIR}/data"
sudo truncate -s 100G "${DATA_DIR}/data"

# Create metadata file
sudo touch "${DATA_DIR}/meta"
sudo truncate -s 10G "${DATA_DIR}/meta"

# Allocate loop devices
DATA_DEV=$(sudo losetup --find --show "${DATA_DIR}/data")
META_DEV=$(sudo losetup --find --show "${DATA_DIR}/meta")

# Define thin-pool parameters.
# See https://www.kernel.org/doc/Documentation/device-mapper/thin-provisioning.txt for details.
SECTOR_SIZE=512
DATA_SIZE="$(sudo blockdev --getsize64 -q ${DATA_DEV})"
LENGTH_IN_SECTORS=$(bc <<< "${DATA_SIZE}/${SECTOR_SIZE}")
DATA_BLOCK_SIZE=128
LOW_WATER_MARK=32768

# Create a thin-pool device
sudo dmsetup create "${POOL_NAME}" \
    --table "0 ${LENGTH_IN_SECTORS} thin-pool ${META_DEV} ${DATA_DEV} ${DATA_BLOCK_SIZE} ${LOW_WATER_MARK}"

```


Edit config for container and restart container with `sudo systemctl restart containerd` and `sudo systemctl start docker`, then use `docker info` to double check. 
```
root@n192-191-015:~/opt/kata/share/defaults/kata-containers# cat /etc/containerd/config.toml
disabled_plugins = []
imports = []
oom_score = 0
plugin_dir = ""
required_plugins = []
root = "/var/lib/containerd"
state = "/run/containerd"
temp = ""
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
  tcp_tls_ca = ""
  tcp_tls_cert = ""
  tcp_tls_key = ""
  uid = 0

[metrics]
  address = ""
  grpc_histogram = false

[plugins]
#  [plugins.devmapper]
#    pool_name = "devpool"
#    root_path = "/var/lib/containerd/devmapper"
#    base_image_size = "10GB"
#    discard_blocks = true

  [plugins."io.containerd.gc.v1.scheduler"]
    deletion_threshold = 0
    mutation_threshold = 100
    pause_threshold = 0.02
    schedule_delay = "0s"
    startup_delay = "100ms"

  [plugins."io.containerd.grpc.v1.cri"]
    cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
    device_ownership_from_security_context = false
    disable_apparmor = false
    disable_cgroup = false
    disable_hugetlb_controller = true
    disable_proc_mount = false
    disable_tcp_service = true
    drain_exec_sync_io_timeout = "0s"
    enable_cdi = false
    enable_selinux = false
    enable_tls_streaming = false
    enable_unprivileged_icmp = false
    enable_unprivileged_ports = false
    ignore_image_defined_volumes = false
    image_pull_progress_timeout = "1m0s"
    max_concurrent_downloads = 3
    max_container_log_line_size = 16384
    netns_mounts_under_state_dir = false
    restrict_oom_score_adj = false
    sandbox_image = "registry.k8s.io/pause:3.8"
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
      ip_pref = ""
      max_conf_num = 1
      setup_serially = false

    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      disable_snapshot_annotations = true
      discard_unpacked_layers = false
      ignore_blockio_not_enabled_errors = false
      ignore_rdt_not_enabled_errors = false
      no_pivot = false
      snapshotter = "overlayfs"

      [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime]
        base_runtime_spec = ""
        cni_conf_dir = ""
        cni_max_conf_num = 0
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        privileged_without_host_devices_all_devices_allowed = false
        runtime_engine = ""
        runtime_path = ""
        runtime_root = ""
        runtime_type = ""
        sandbox_mode = ""
        snapshotter = ""

        [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
          runtime_type = "io.containerd.kata.v2"

	[plugin."io.containerd.grpc.v1.cri".containerd.runtimes.kata-fc]
            runtime_type = "io.containerd.kata-fc.v2"

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = ""
          cni_conf_dir = ""
          cni_max_conf_num = 0
          container_annotations = []
          pod_annotations = []
          privileged_without_host_devices = false
          privileged_without_host_devices_all_devices_allowed = false
          runtime_engine = ""
          runtime_path = ""
          runtime_root = ""
          runtime_type = "io.containerd.runc.v2"
          sandbox_mode = "podsandbox"
          snapshotter = ""

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

      [plugins."io.containerd.grpc.v1.cri".containerd.untrusted_workload_runtime]
        base_runtime_spec = ""
        cni_conf_dir = ""
        cni_max_conf_num = 0
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        privileged_without_host_devices_all_devices_allowed = false
        runtime_engine = ""
        runtime_path = ""
        runtime_root = ""
        runtime_type = ""
        sandbox_mode = ""
        snapshotter = ""

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

  [plugins."io.containerd.internal.v1.tracing"]
    sampling_ratio = 1.0
    service_name = "containerd"

  [plugins."io.containerd.metadata.v1.bolt"]
    content_sharing_policy = "shared"

  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false

  [plugins."io.containerd.nri.v1.nri"]
    disable = true
    disable_connections = false
    plugin_config_path = "/etc/nri/conf.d"
    plugin_path = "/opt/nri/plugins"
    plugin_registration_timeout = "5s"
    plugin_request_timeout = "2s"
    socket_path = "/var/run/nri/nri.sock"

  [plugins."io.containerd.runtime.v1.linux"]
    no_shim = false
    runtime = "runc"
    runtime_root = ""
    shim = "containerd-shim"
    shim_debug = false

  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = false

  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]

  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""

  [plugins."io.containerd.snapshotter.v1.aufs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.btrfs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.devmapper"]
    base_image_size = "10G"
    discard_blocks = true
    fs_options = ""
    fs_type = ""
     pool_name = "devpool"
     root_path = "/var/lib/containerd/devmapper"

  [plugins."io.containerd.snapshotter.v1.native"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.overlayfs"]
    root_path = ""
    upperdir_label = false

  [plugins."io.containerd.snapshotter.v1.zfs"]
    root_path = ""

  [plugins."io.containerd.tracing.processor.v1.otlp"]
    endpoint = ""
    insecure = false
    protocol = ""

  [plugins."io.containerd.transfer.v1.local"]

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
  "io.containerd.timeout.bolt.open" = "0s"
  "io.containerd.timeout.metrics.shimstats" = "2s"
  "io.containerd.timeout.shim.cleanup" = "5s"
  "io.containerd.timeout.shim.load" = "5s"
  "io.containerd.timeout.shim.shutdown" = "3s"
  "io.containerd.timeout.task.state" = "2s"

[ttrpc]
  address = ""
  gid = 0
  uid = 0

```

3. Edit configuration for Kata


Add containerd-shim-kata-fc-v2

```
vim /usr/local/bin/containerd-shim-kata-fc-v2

root@n192-191-015:~/opt/kata/share/defaults/kata-containers# cat /usr/local/bin/containerd-shim-kata-fc-v2
#!/bin/bash
KATA_CONF_FILE=/etc/kata-containers/configuration-fc.toml /usr/local/bin/containerd-shim-kata-v2 $@

chmod 777 /usr/local/bin/containerd-shim-kata-fc-v2
```



Edit kata configuration, restart container with `sudo systemctl restart containerd` and `sudo systemctl start docker`
```
root@n192-191-015:~/opt/kata/share/defaults/kata-containers# cat /etc/kata-containers/configuration.toml 
[hypervisor.firecracker]
path = "/root/opt/kata/bin/firecracker"
kernel = "/usr/share/kata-containers/vmlinux.container"
image = "/root/opt/kata/share/kata-containers/kata-containers.img"

rootfs_type="ext4"
enable_annotations = ["enable_iommu"]
valid_hypervisor_paths = ["/root/opt/kata/bin/firecracker"]
jailer_path = "/root/opt/kata/bin/jailer"
valid_jailer_paths = ["/root/opt/kata/bin/jailer"]
kernel_params = ""
default_vcpus = 1
default_bridges = 1
default_memory = 2048
default_maxmemory = 0



disable_block_device_use = false
shared_fs = "virtio-fs"
virtio_fs_daemon = "/root/opt/kata/libexec/virtiofsd"
valid_virtio_fs_daemon_paths = ["/root/opt/kata/libexec/virtiofsd"]
virtio_fs_cache_size = 0
virtio_fs_queue_size = 1024
virtio_fs_extra_args = ["--thread-pool-size=1", "-o", "announce_submounts"]
virtio_fs_cache = "auto"
block_device_driver = "virtio-mmio"
block_device_aio = "io_uring"
enable_iothreads = false
enable_vhost_user_store = false
vhost_user_store_path = "/var/run/kata-containers/vhost-user"
valid_vhost_user_store_paths = ["/var/run/kata-containers/vhost-user"]
vhost_user_reconnect_timeout_sec = 0
valid_file_mem_backends = [""]
pflashes = []
valid_entropy_sources = ["/dev/urandom","/dev/random",""]
disable_selinux=false
disable_guest_selinux=true



[factory]

[agent.kata]
kernel_modules=[]

[runtime]
internetworking_model="tcfilter"
disable_guest_seccomp=true
sandbox_cgroup_only=false
static_sandbox_resource_mgmt=true
disable_guest_empty_dir=false
experimental=[]

```


4. test kata fc

```
root@n192-191-015:~/opt/kata/share/defaults/kata-containers# ctr run --snapshotter devmapper --runtime io.containerd.run.kata-fc.v2 -t --rm  --mount type=bind,src=/root/umk-bench,dst=/root/umk-bench,options=rbind:rw --mount type=bind,src=/usr,dst=/usr,options=rbind:ro  docker.io/library/ubuntu:latest hello12 sh
# uname -a
Linux localhost 5.15.0 #1 SMP Fri Mar 31 22:02:31 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux
# 

```
