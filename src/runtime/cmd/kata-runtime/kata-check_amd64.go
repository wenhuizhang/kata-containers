// Copyright (c) 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	vc "github.com/kata-containers/kata-containers/src/runtime/virtcontainers"
	"github.com/sirupsen/logrus"
)

const (
	cpuFlagsTag               = genericCPUFlagsTag
	archCPUVendorField        = genericCPUVendorField
	archCPUModelField         = genericCPUModelField
	archGenuineIntel          = "GenuineIntel"
	archAuthenticAMD          = "AuthenticAMD"
	msgKernelVM               = "Kernel-based Virtual Machine"
	msgKernelVirtio           = "Host kernel accelerator for virtio"
	msgKernelVirtioNet        = "Host kernel accelerator for virtio network"
	msgKernelVirtioVhostVsock = "Host Support for Linux VM Sockets"
	cpuFlagVMX                = "vmx"
	cpuFlagLM                 = "lm"
	cpuFlagSVM                = "svm"
	cpuFlagSSE4_1             = "sse4_1"
	kernelModvhost            = "vhost"
	kernelModvhostnet         = "vhost_net"
	kernelModvhostvsock       = "vhost_vsock"
	kernelModkvm              = "kvm"
	kernelModkvmintel         = "kvm_intel"
	kernelModkvmamd           = "kvm_amd"
)

// CPU types
const (
	cpuTypeIntel   = 0
	cpuTypeAMD     = 1
	cpuTypeUnknown = -1
)

const acrnDevice = "/dev/acrn_hsm"

// ioctl_ACRN_CREATE_VM is the IOCTL to create VM in ACRN.
// Current Linux mainstream kernel doesn't have support for ACRN.
// Due to this several macros are not defined in Linux headers.
// Until the support is available, directly use the value instead
// of macros.
// https://github.com/kata-containers/runtime/issues/1784
const ioctl_ACRN_CREATE_VM = 0xC030A210 //nolint
const ioctl_ACRN_PAUSE_VM = 0xA213      //nolint
const ioctl_ACRN_DESTROY_VM = 0xA211    //nolint

type acrn_vm_creation struct { //nolint
	vmid         uint16 //nolint
	reserved0    uint16 //nolint
	vcpu_num     uint16 //nolint
	reserved1    uint16 //nolint
	name         [16]uint8
	vm_flag      uint64 //nolint
	ioreq_buf    uint64 //nolint
	cpu_affinity uint64 //nolint
}

var io_request_page [4096]byte

type acrn_io_request struct { // nolint
	io_type            uint32     // nolint
	completion_polling uint32     // nolint
	reserved0          [14]uint32 // nolint
	data               [8]uint64  // nolint
	reserved1          uint32     // nolint
	kernel_handled     uint32     // nolint
	processed          uint32     // nolint
}

// cpuType save the CPU type
var cpuType int

// archRequiredCPUFlags maps a CPU flag value to search for and a
// human-readable description of that value.
var archRequiredCPUFlags map[string]string

// archRequiredCPUAttribs maps a CPU (non-CPU flag) attribute value to search for
// and a human-readable description of that value.
var archRequiredCPUAttribs map[string]string

// archRequiredKernelModules maps a required module name to a human-readable
// description of the modules functionality and an optional list of
// required module parameters.
var archRequiredKernelModules map[string]kernelModule

func setCPUtype(hypervisorType vc.HypervisorType) error {
	cpuType = getCPUtype()

	if cpuType == cpuTypeUnknown {
		return fmt.Errorf("Unknow CPU Type")
	} else if cpuType == cpuTypeIntel {
		var kvmIntelParams map[string]string
		onVMM, err := vc.RunningOnVMM(procCPUInfo)
		if err != nil && !onVMM {
			kvmIntelParams = map[string]string{
				// "VMX Unrestricted mode support". This is used
				// as a heuristic to determine if the system is
				// "new enough" to run a Kata Container
				// (atleast a Westmere).
				"unrestricted_guest": "Y",
			}
		}

		switch hypervisorType {
		case "firecracker":
			fallthrough
		case "clh":
			fallthrough
		case "dragonball":
			fallthrough
		case "qemu":
			archRequiredCPUFlags = map[string]string{
				cpuFlagVMX:    "Virtualization support",
				cpuFlagLM:     "64Bit CPU",
				cpuFlagSSE4_1: "SSE4.1",
			}
			archRequiredCPUAttribs = map[string]string{
				archGenuineIntel: "Intel Architecture CPU",
			}
			archRequiredKernelModules = map[string]kernelModule{
				kernelModkvm: {
					desc:     msgKernelVM,
					required: true,
				},
				kernelModkvmintel: {
					desc:       "Intel KVM",
					parameters: kvmIntelParams,
					required:   true,
				},
				kernelModvhost: {
					desc:     msgKernelVirtio,
					required: true,
				},
				kernelModvhostnet: {
					desc:     msgKernelVirtioNet,
					required: true,
				},
				kernelModvhostvsock: {
					desc:     msgKernelVirtioVhostVsock,
					required: false,
				},
			}
		case "acrn":
			archRequiredCPUFlags = map[string]string{
				cpuFlagLM:     "64Bit CPU",
				cpuFlagSSE4_1: "SSE4.1",
			}
			archRequiredCPUAttribs = map[string]string{
				archGenuineIntel: "Intel Architecture CPU",
			}
			archRequiredKernelModules = map[string]kernelModule{
				kernelModvhost: {
					desc:     msgKernelVirtio,
					required: false,
				},
				kernelModvhostnet: {
					desc:     msgKernelVirtioNet,
					required: false,
				},
				kernelModvhostvsock: {
					desc:     msgKernelVirtioVhostVsock,
					required: false,
				},
			}
		case "mock":
			archRequiredCPUFlags = map[string]string{
				cpuFlagVMX:    "Virtualization support",
				cpuFlagLM:     "64Bit CPU",
				cpuFlagSSE4_1: "SSE4.1",
			}
			archRequiredCPUAttribs = map[string]string{
				archGenuineIntel: "Intel Architecture CPU",
			}

		default:
			return fmt.Errorf("setCPUtype: Unknown hypervisor type %s", hypervisorType)
		}

	} else if cpuType == cpuTypeAMD {
		archRequiredCPUFlags = map[string]string{
			cpuFlagSVM:    "Virtualization support",
			cpuFlagLM:     "64Bit CPU",
			cpuFlagSSE4_1: "SSE4.1",
		}
		archRequiredCPUAttribs = map[string]string{
			archAuthenticAMD: "AMD Architecture CPU",
		}
		archRequiredKernelModules = map[string]kernelModule{
			kernelModkvm: {
				desc:     msgKernelVM,
				required: true,
			},
			kernelModkvmamd: {
				desc:     "AMD KVM",
				required: true,
			},
			kernelModvhost: {
				desc:     msgKernelVirtio,
				required: true,
			},
			kernelModvhostnet: {
				desc:     msgKernelVirtioNet,
				required: true,
			},
			kernelModvhostvsock: {
				desc:     msgKernelVirtioVhostVsock,
				required: false,
			},
		}
	}

	return nil
}

func getCPUtype() int {
	content, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		kataLog.WithError(err).Error("failed to read file")
		return cpuTypeUnknown
	}
	str := string(content)
	if strings.Contains(str, archGenuineIntel) {
		return cpuTypeIntel
	} else if strings.Contains(str, archAuthenticAMD) {
		return cpuTypeAMD
	} else {
		return cpuTypeUnknown
	}
}

// kvmIsUsable determines if it will be possible to create a full virtual machine
// by creating a minimal VM and then deleting it.
func kvmIsUsable() error {
	return genericKvmIsUsable()
}

// acrnIsUsable determines if it will be possible to create a full virtual machine
// by creating a minimal VM and then deleting it.
func acrnIsUsable() error {
	flags := syscall.O_RDWR | syscall.O_CLOEXEC

	f, err := syscall.Open(acrnDevice, flags, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(f)
	kataLog.WithField("device", acrnDevice).Info("device available")

	var createVM acrn_vm_creation
	copy(createVM.name[:], "KataACRNVM")
	ioreq_buf := (*acrn_io_request)(unsafe.Pointer(&io_request_page))
	createVM.ioreq_buf = uint64(uintptr(unsafe.Pointer(ioreq_buf)))

	ret, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(f),
		uintptr(ioctl_ACRN_CREATE_VM),
		uintptr(unsafe.Pointer(&createVM)))
	if ret != 0 || errno != 0 {
		if errno == syscall.EBUSY {
			kataLog.WithField("reason", "another hypervisor running").Error("cannot create VM")
		}
		kataLog.WithFields(logrus.Fields{
			"ret":     ret,
			"errno":   errno,
			"VM_name": createVM.name,
		}).Info("Create VM Error")
		return errno
	}

	ret, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(f),
		uintptr(ioctl_ACRN_PAUSE_VM),
		0)
	if ret != 0 || errno != 0 {
		kataLog.WithFields(logrus.Fields{
			"ret":   ret,
			"errno": errno,
		}).Info("PAUSE VM Error")
		return errno
	}

	ret, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(f),
		uintptr(ioctl_ACRN_DESTROY_VM),
		0)
	if ret != 0 || errno != 0 {
		kataLog.WithFields(logrus.Fields{
			"ret":   ret,
			"errno": errno,
		}).Info("Destroy VM Error")
		return errno
	}

	kataLog.WithField("feature", "create-vm").Info("feature available")

	return nil
}

func archHostCanCreateVMContainer(hypervisorType vc.HypervisorType) error {

	switch hypervisorType {
	case "qemu":
		fallthrough
	case "clh":
		fallthrough
	case "firecracker":
		return kvmIsUsable()
	case "acrn":
		return acrnIsUsable()
	case "remote":
		return nil
	case "mock":
		return nil
	default:
		return fmt.Errorf("archHostCanCreateVMContainer: Unknown hypervisor type %s", hypervisorType)
	}
}

// hostIsVMContainerCapable checks to see if the host is theoretically capable
// of creating a VM container.
func hostIsVMContainerCapable(details vmContainerCapableDetails) error {
	return genericHostIsVMContainerCapable(details)
}

func archKernelParamHandler(onVMM bool, fields logrus.Fields, msg string) bool {
	return genericArchKernelParamHandler(onVMM, fields, msg)
}

func getCPUDetails() (vendor, model string, err error) {
	return genericGetCPUDetails()
}
