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
