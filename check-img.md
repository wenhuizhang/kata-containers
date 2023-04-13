

1. Check Kata-container img

```
sudo mount -o loop,offset=3145728 ./kata-containers.img /mnt
umount /mnt
```


2. Check FC img

```
sudo mount -o loop,offset=0 ./bionic.rootfs.ext4_firecracker_2G /mnt
umount /mnt
```


3. Check qemu image


```
# Step 1 - Enable NBD on the Host**
    
    modprobe nbd max_part=8

# Step 2 - Connect the QCOW2 as network block device**

    qemu-nbd --connect=/dev/nbd0 /root/ubuntu-jammy-amd64-uefi-bm.qcow2

# Step 3 - Find The Virtual Machine Partitions**

    fdisk /dev/nbd0 -l

# Step 4 - Mount the partition from the VM**

    mount /dev/nbd0p2 /mnt/

# Step 5 - After you done, unmount and disconnect**

    umount /mnt/
    qemu-nbd --disconnect /dev/nbd0
    rmmod nbd
```


