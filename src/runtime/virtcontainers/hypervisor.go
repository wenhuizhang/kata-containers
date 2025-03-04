// Copyright (c) 2016 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package virtcontainers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/kata-containers/kata-containers/src/runtime/pkg/device/config"
	"github.com/kata-containers/kata-containers/src/runtime/pkg/govmm"
	hv "github.com/kata-containers/kata-containers/src/runtime/pkg/hypervisors"
	"github.com/kata-containers/kata-containers/src/runtime/virtcontainers/types"

	"github.com/sirupsen/logrus"
)

// HypervisorType describes an hypervisor type.
type HypervisorType string

type Operation int

const (
	AddDevice Operation = iota
	RemoveDevice
)

const (
	// FirecrackerHypervisor is the FC hypervisor.
	FirecrackerHypervisor HypervisorType = "firecracker"

	// QemuHypervisor is the QEMU hypervisor.
	QemuHypervisor HypervisorType = "qemu"

	// AcrnHypervisor is the ACRN hypervisor.
	AcrnHypervisor HypervisorType = "acrn"

	// ClhHypervisor is the ICH hypervisor.
	ClhHypervisor HypervisorType = "clh"

	// DragonballHypervisor is the Dragonball hypervisor.
	DragonballHypervisor HypervisorType = "dragonball"

	// RemoteHypervisor is the Remote hypervisor.
	RemoteHypervisor HypervisorType = "remote"

	// VirtFrameworkHypervisor is the Darwin Virtualization.framework hypervisor
	VirtframeworkHypervisor HypervisorType = "virtframework"

	// MockHypervisor is a mock hypervisor for testing purposes
	MockHypervisor HypervisorType = "mock"

	procCPUInfo = "/proc/cpuinfo"

	defaultVCPUs = 1
	// 2 GiB
	defaultMemSzMiB = 2048

	defaultBridges = 1

	defaultBlockDriver = config.VirtioSCSI

	// port numbers below 1024 are called privileged ports. Only a process with
	// CAP_NET_BIND_SERVICE capability may bind to these port numbers.
	vSockPort = 1024

	// Port where the agent will send the logs. Logs are sent through the vsock in cases
	// where the hypervisor has no console.sock, i.e firecracker
	vSockLogsPort = 1025

	// MinHypervisorMemory is the minimum memory required for a VM.
	MinHypervisorMemory = 256

	defaultMsize9p = 8192

	defaultDisableGuestSeLinux = true
)

var (
	hvLogger                   = logrus.WithField("source", "virtcontainers/hypervisor")
	noGuestMemHotplugErr error = errors.New("guest memory hotplug not supported")
)

// In some architectures the maximum number of vCPUs depends on the number of physical cores.
// TODO (dcantah): Find a suitable value for darwin/vfw. Seems perf degrades if > number of host
// cores.
var defaultMaxVCPUs = govmm.MaxVCPUs()

// agnostic list of kernel root parameters for NVDIMM
var commonNvdimmKernelRootParams = []Param{ //nolint: unused, deadcode, varcheck
	{"root", "/dev/pmem0p1"},
	{"rootflags", "dax,data=ordered,errors=remount-ro ro"},
	{"rootfstype", "ext4"},
}

// agnostic list of kernel root parameters for NVDIMM
var commonNvdimmNoDAXKernelRootParams = []Param{ //nolint: unused, deadcode, varcheck
	{"root", "/dev/pmem0p1"},
	{"rootflags", "data=ordered,errors=remount-ro ro"},
	{"rootfstype", "ext4"},
}

// agnostic list of kernel root parameters for virtio-blk
var commonVirtioblkKernelRootParams = []Param{ //nolint: unused, deadcode, varcheck
	{"root", "/dev/vda1"},
	{"rootflags", "data=ordered,errors=remount-ro ro"},
	{"rootfstype", "ext4"},
}

// DeviceType describes a virtualized device type.
type DeviceType int

const (
	// ImgDev is the image device type.
	ImgDev DeviceType = iota

	// FsDev is the filesystem device type.
	FsDev

	// NetDev is the network device type.
	NetDev

	// BlockDev is the block device type.
	BlockDev

	// SerialPortDev is the serial port device type.
	SerialPortDev

	// VSockPCIDev is the vhost vsock PCI device type.
	VSockPCIDev

	// VFIODevice is VFIO device type
	VfioDev

	// VhostuserDev is a Vhost-user device type
	VhostuserDev

	// CPUDevice is CPU device type
	CpuDev

	// MemoryDev is memory device type
	MemoryDev

	// HybridVirtioVsockDev is a hybrid virtio-vsock device supported
	// only on certain hypervisors, like firecracker.
	HybridVirtioVsockDev
)

type MemoryDevice struct {
	Slot   int
	SizeMB int
	Addr   uint64
	Probe  bool
}

// SetHypervisorLogger sets up a logger for the hypervisor part of this pkg
func SetHypervisorLogger(logger *logrus.Entry) {
	fields := hvLogger.Data
	hvLogger = logger.WithFields(fields)
}

// Set sets an hypervisor type based on the input string.
func (hType *HypervisorType) Set(value string) error {
	switch value {
	case "qemu":
		*hType = QemuHypervisor
		return nil
	case "firecracker":
		*hType = FirecrackerHypervisor
		return nil
	case "acrn":
		*hType = AcrnHypervisor
		return nil
	case "clh":
		*hType = ClhHypervisor
		return nil
	case "dragonball":
		*hType = DragonballHypervisor
		return nil
	case "remote":
		*hType = RemoteHypervisor
		return nil
	case "virtframework":
		*hType = VirtframeworkHypervisor
		return nil
	case "mock":
		*hType = MockHypervisor
		return nil
	default:
		return fmt.Errorf("Unknown hypervisor type %s", value)
	}
}

// String converts an hypervisor type to a string.
func (hType *HypervisorType) String() string {
	switch *hType {
	case QemuHypervisor:
		return string(QemuHypervisor)
	case FirecrackerHypervisor:
		return string(FirecrackerHypervisor)
	case AcrnHypervisor:
		return string(AcrnHypervisor)
	case ClhHypervisor:
		return string(ClhHypervisor)
	case RemoteHypervisor:
		return string(RemoteHypervisor)
	case MockHypervisor:
		return string(MockHypervisor)
	default:
		return ""
	}
}

// GetHypervisorSocketTemplate returns the full "template" path to the
// hypervisor socket. If the specified hypervisor doesn't use a socket,
// an empty string is returned.
//
// The returned value is not the actual socket path since this function
// does not create a sandbox. Instead a path is returned with a special
// template value "{ID}" which would be replaced with the real sandbox
// name sandbox creation time.
func GetHypervisorSocketTemplate(hType HypervisorType, config *HypervisorConfig) (string, error) {
	hypervisor, err := NewHypervisor(hType)
	if err != nil {
		return "", err
	}

	if err := hypervisor.setConfig(config); err != nil {
		return "", err
	}

	// Tag that is used to represent the name of a sandbox
	const sandboxID = "{ID}"

	socket, err := hypervisor.GenerateSocket(sandboxID)
	if err != nil {
		return "", err
	}

	var socketPath string

	if hybridVsock, ok := socket.(types.HybridVSock); ok {
		socketPath = hybridVsock.UdsPath
	}

	return socketPath, nil
}

// Param is a key/value representation for hypervisor and kernel parameters.
type Param struct {
	Key   string
	Value string
}

// HypervisorConfig is the hypervisor configuration.
type HypervisorConfig struct {
	customAssets                   map[types.AssetType]*types.Asset
	SeccompSandbox                 string
	KernelPath                     string
	ImagePath                      string
	InitrdPath                     string
	FirmwarePath                   string
	FirmwareVolumePath             string
	MachineAccelerators            string
	CPUFeatures                    string
	HypervisorPath                 string
	HypervisorCtlPath              string
	GuestPreAttestationKeyset      string
	BlockDeviceDriver              string
	HypervisorMachineType          string
	GuestPreAttestationProxy       string
	DevicesStatePath               string
	EntropySource                  string
	SharedFS                       string
	SharedPath                     string
	VirtioFSDaemon                 string
	VirtioFSCache                  string
	FileBackedMemRootDir           string
	VhostUserStorePath             string
	GuestMemoryDumpPath            string
	GuestHookPath                  string
	VMid                           string
	VMStorePath                    string
	RunStorePath                   string
	SELinuxProcessLabel            string
	JailerPath                     string
	MemoryPath                     string
	GuestPreAttestationSecretGuid  string
	GuestPreAttestationSecretType  string
	SEVCertChainPath               string
	BlockDeviceAIO                 string
	User                           string
	RemoteHypervisorSocket         string
	SandboxName                    string
	SandboxNamespace               string
	JailerPathList                 []string
	EntropySourceList              []string
	VirtioFSDaemonList             []string
	VirtioFSExtraArgs              []string
	EnableAnnotations              []string
	FileBackedMemRootList          []string
	PFlash                         []string
	VhostUserStorePathList         []string
	HypervisorCtlPathList          []string
	KernelParams                   []Param
	Groups                         []uint32
	HypervisorPathList             []string
	HypervisorParams               []Param
	DiskRateLimiterBwOneTimeBurst  int64
	DiskRateLimiterOpsMaxRate      int64
	DiskRateLimiterOpsOneTimeBurst int64
	SGXEPCSize                     int64
	DefaultMaxMemorySize           uint64
	NetRateLimiterBwMaxRate        int64
	NetRateLimiterBwOneTimeBurst   int64
	NetRateLimiterOpsMaxRate       int64
	NetRateLimiterOpsOneTimeBurst  int64
	MemOffset                      uint64
	TxRateLimiterMaxRate           uint64
	DiskRateLimiterBwMaxRate       int64
	RxRateLimiterMaxRate           uint64
	MemorySize                     uint32
	DefaultMaxVCPUs                uint32
	DefaultBridges                 uint32
	Msize9p                        uint32
	MemSlots                       uint32
	VirtioFSCacheSize              uint32
	VirtioFSQueueSize              uint32
	Uid                            uint32
	Gid                            uint32
	SEVGuestPolicy                 uint32
	PCIeRootPort                   uint32
	NumVCPUs                       uint32
	RemoteHypervisorTimeout        uint32
	IOMMUPlatform                  bool
	EnableIOThreads                bool
	Debug                          bool
	MemPrealloc                    bool
	HugePages                      bool
	VirtioMem                      bool
	IOMMU                          bool
	DisableBlockDeviceUse          bool
	DisableNestingChecks           bool
	DisableImageNvdimm             bool
	HotplugVFIOOnRootBus           bool
	GuestMemoryDumpPaging          bool
	ConfidentialGuest              bool
	SevSnpGuest                    bool
	GuestPreAttestation            bool
	BlockDeviceCacheNoflush        bool
	BlockDeviceCacheDirect         bool
	BlockDeviceCacheSet            bool
	BootToBeTemplate               bool
	BootFromTemplate               bool
	DisableVhostNet                bool
	EnableVhostUserStore           bool
	GuestSwap                      bool
	Rootless                       bool
	DisableSeccomp                 bool
	DisableSeLinux                 bool
	DisableGuestSeLinux            bool
	LegacySerial                   bool
	EnableVCPUsPinning             bool
}

// vcpu mapping from vcpu number to thread number
type VcpuThreadIDs struct {
	vcpus map[int]int
}

func (conf *HypervisorConfig) CheckTemplateConfig() error {
	if conf.BootToBeTemplate && conf.BootFromTemplate {
		return fmt.Errorf("Cannot set both 'to be' and 'from' vm tempate")
	}

	if conf.BootToBeTemplate || conf.BootFromTemplate {
		if conf.MemoryPath == "" {
			return fmt.Errorf("Missing MemoryPath for vm template")
		}

		if conf.BootFromTemplate && conf.DevicesStatePath == "" {
			return fmt.Errorf("Missing DevicesStatePath to Load from vm template")
		}
	}

	return nil
}

// AddKernelParam allows the addition of new kernel parameters to an existing
// hypervisor configuration.
func (conf *HypervisorConfig) AddKernelParam(p Param) error {
	if p.Key == "" {
		return fmt.Errorf("Empty kernel parameter")
	}

	conf.KernelParams = append(conf.KernelParams, p)

	return nil
}

func (conf *HypervisorConfig) AddCustomAsset(a *types.Asset) error {
	if a == nil || a.Path() == "" {
		// We did not get a custom asset, we will use the default one.
		return nil
	}

	if !a.Valid() {
		return fmt.Errorf("Invalid %s at %s", a.Type(), a.Path())
	}

	hvLogger.Debugf("Using custom %v asset %s", a.Type(), a.Path())

	if conf.customAssets == nil {
		conf.customAssets = make(map[types.AssetType]*types.Asset)
	}

	conf.customAssets[a.Type()] = a

	return nil
}

func (conf *HypervisorConfig) assetPath(t types.AssetType) (string, error) {
	// Custom assets take precedence over the configured ones
	a, ok := conf.customAssets[t]
	if ok {
		return a.Path(), nil
	}

	// We could not find a custom asset for the given type, let's
	// fall back to the configured ones.
	switch t {
	case types.KernelAsset:
		return conf.KernelPath, nil
	case types.ImageAsset:
		return conf.ImagePath, nil
	case types.InitrdAsset:
		return conf.InitrdPath, nil
	case types.HypervisorAsset:
		return conf.HypervisorPath, nil
	case types.HypervisorCtlAsset:
		return conf.HypervisorCtlPath, nil
	case types.JailerAsset:
		return conf.JailerPath, nil
	case types.FirmwareAsset:
		return conf.FirmwarePath, nil
	case types.FirmwareVolumeAsset:
		return conf.FirmwareVolumePath, nil
	default:
		return "", fmt.Errorf("Unknown asset type %v", t)
	}
}

func (conf *HypervisorConfig) isCustomAsset(t types.AssetType) bool {
	_, ok := conf.customAssets[t]
	return ok
}

// KernelAssetPath returns the guest kernel path
func (conf *HypervisorConfig) KernelAssetPath() (string, error) {
	return conf.assetPath(types.KernelAsset)
}

// CustomKernelAsset returns true if the kernel asset is a custom one, false otherwise.
func (conf *HypervisorConfig) CustomKernelAsset() bool {
	return conf.isCustomAsset(types.KernelAsset)
}

// ImageAssetPath returns the guest image path
func (conf *HypervisorConfig) ImageAssetPath() (string, error) {
	return conf.assetPath(types.ImageAsset)
}

// CustomImageAsset returns true if the image asset is a custom one, false otherwise.
func (conf *HypervisorConfig) CustomImageAsset() bool {
	return conf.isCustomAsset(types.ImageAsset)
}

// InitrdAssetPath returns the guest initrd path
func (conf *HypervisorConfig) InitrdAssetPath() (string, error) {
	return conf.assetPath(types.InitrdAsset)
}

// CustomInitrdAsset returns true if the initrd asset is a custom one, false otherwise.
func (conf *HypervisorConfig) CustomInitrdAsset() bool {
	return conf.isCustomAsset(types.InitrdAsset)
}

// HypervisorAssetPath returns the VM hypervisor path
func (conf *HypervisorConfig) HypervisorAssetPath() (string, error) {
	return conf.assetPath(types.HypervisorAsset)
}

func (conf *HypervisorConfig) IfPVPanicEnabled() bool {
	return conf.GuestMemoryDumpPath != ""
}

// HypervisorCtlAssetPath returns the VM hypervisor ctl path
func (conf *HypervisorConfig) HypervisorCtlAssetPath() (string, error) {
	return conf.assetPath(types.HypervisorCtlAsset)
}

// CustomHypervisorAsset returns true if the hypervisor asset is a custom one, false otherwise.
func (conf *HypervisorConfig) CustomHypervisorAsset() bool {
	return conf.isCustomAsset(types.HypervisorAsset)
}

// FirmwareAssetPath returns the guest firmware path
func (conf *HypervisorConfig) FirmwareAssetPath() (string, error) {
	return conf.assetPath(types.FirmwareAsset)
}

// FirmwareVolumeAssetPath returns the guest firmware volume path
func (conf *HypervisorConfig) FirmwareVolumeAssetPath() (string, error) {
	return conf.assetPath(types.FirmwareVolumeAsset)
}

func appendParam(params []Param, parameter string, value string) []Param {
	return append(params, Param{parameter, value})
}

// SerializeParams converts []Param to []string
func SerializeParams(params []Param, delim string) []string {
	var parameters []string

	for _, p := range params {
		if p.Key == "" && p.Value == "" {
			continue
		} else if p.Key == "" {
			parameters = append(parameters, fmt.Sprint(p.Value))
		} else if p.Value == "" {
			parameters = append(parameters, fmt.Sprint(p.Key))
		} else if delim == "" {
			parameters = append(parameters, fmt.Sprint(p.Key))
			parameters = append(parameters, fmt.Sprint(p.Value))
		} else {
			parameters = append(parameters, fmt.Sprintf("%s%s%s", p.Key, delim, p.Value))
		}
	}

	return parameters
}

// DeserializeParams converts []string to []Param
func DeserializeParams(parameters []string) []Param {
	var params []Param

	for _, param := range parameters {
		if param == "" {
			continue
		}
		p := strings.SplitN(param, "=", 2)
		if len(p) == 2 {
			params = append(params, Param{Key: p[0], Value: p[1]})
		} else {
			params = append(params, Param{Key: p[0], Value: ""})
		}
	}

	return params
}

// CheckCmdline checks whether an option or parameter is present in the kernel command line.
// Search is case-insensitive.
// Takes path to file that contains the kernel command line, desired option, and permitted values
// (empty values to Check for options).
func CheckCmdline(kernelCmdlinePath, searchParam string, searchValues []string) (bool, error) {
	f, err := os.Open(kernelCmdlinePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Create Check function -- either Check for verbatim option
	// or Check for parameter and permitted values
	var check func(string, string, []string) bool
	if len(searchValues) == 0 {
		check = func(option, searchParam string, _ []string) bool {
			return strings.EqualFold(option, searchParam)
		}
	} else {
		check = func(param, searchParam string, searchValues []string) bool {
			// split parameter and value
			split := strings.SplitN(param, "=", 2)
			if len(split) < 2 || split[0] != searchParam {
				return false
			}
			for _, value := range searchValues {
				if strings.EqualFold(value, split[1]) {
					return true
				}
			}
			return false
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		for _, field := range strings.Fields(scanner.Text()) {
			if check(field, searchParam, searchValues) {
				return true, nil
			}
		}
	}
	return false, err
}

func CPUFlags(cpuInfoPath string) (map[string]bool, error) {
	flagsField := "flags"

	f, err := os.Open(cpuInfoPath)
	if err != nil {
		return map[string]bool{}, err
	}
	defer f.Close()

	flags := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Expected format: ["flags", ":", ...] or ["flags:", ...]
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		if !strings.HasPrefix(fields[0], flagsField) {
			continue
		}

		for _, field := range fields[1:] {
			flags[field] = true
		}

		return flags, nil
	}

	if err := scanner.Err(); err != nil {
		return map[string]bool{}, err
	}

	return map[string]bool{}, fmt.Errorf("Couldn't find %q from %q output", flagsField, cpuInfoPath)
}

// RunningOnVMM checks if the system is running inside a VM.
func RunningOnVMM(cpuInfoPath string) (bool, error) {
	if runtime.GOARCH == "amd64" {
		flags, err := CPUFlags(cpuInfoPath)
		if err != nil {
			return false, err
		}
		return flags["hypervisor"], nil
	}

	hvLogger.WithField("arch", runtime.GOARCH).Info("Unable to know if the system is running inside a VM")
	return false, nil
}

func GetHypervisorPid(h Hypervisor) int {
	pids := h.GetPids()
	if len(pids) == 0 {
		return 0
	}
	return pids[0]
}

// Kind of guest protection
type guestProtection uint8

const (
	noneProtection guestProtection = iota

	//Intel Trust Domain Extensions
	//https://software.intel.com/content/www/us/en/develop/articles/intel-trust-domain-extensions.html
	// Exclude from lint checking for it won't be used on arm64 code
	tdxProtection

	// AMD Secure Encrypted Virtualization
	// https://developer.amd.com/sev/
	// Exclude from lint checking for it won't be used on arm64 code
	sevProtection

	// AMD Secure Encrypted Virtualization - Secure Nested Paging (SEV-SNP)
	// https://developer.amd.com/sev/
	// Exclude from lint checking for it won't be used on arm64 code
	snpProtection

	// IBM POWER 9 Protected Execution Facility
	// https://www.kernel.org/doc/html/latest/powerpc/ultravisor.html
	// Exclude from lint checking for it won't be used on arm64 code
	pefProtection

	// IBM Secure Execution (IBM Z & LinuxONE)
	// https://www.kernel.org/doc/html/latest/virt/kvm/s390-pv.html
	// Exclude from lint checking for it won't be used on arm64 code
	seProtection
)

var guestProtectionStr = [...]string{
	noneProtection: "none",
	pefProtection:  "pef",
	seProtection:   "se",
	sevProtection:  "sev",
	snpProtection:  "snp",
	tdxProtection:  "tdx",
}

func (gp guestProtection) String() string {
	return guestProtectionStr[gp]
}

func genericAvailableGuestProtections() (protections []string) {
	return
}

func AvailableGuestProtections() (protections []string) {
	gp, err := availableGuestProtection()
	if err != nil || gp == noneProtection {
		return genericAvailableGuestProtections()
	}
	return []string{gp.String()}
}

// hypervisor is the virtcontainers hypervisor interface.
// The default hypervisor implementation is Qemu.
type Hypervisor interface {
	CreateVM(ctx context.Context, id string, network Network, hypervisorConfig *HypervisorConfig) error
	StartVM(ctx context.Context, timeout int) error
	AttestVM(ctx context.Context) error

	// If wait is set, don't actively stop the sandbox:
	// just perform cleanup.
	StopVM(ctx context.Context, waitOnly bool) error
	PauseVM(ctx context.Context) error
	SaveVM() error
	ResumeVM(ctx context.Context) error
	AddDevice(ctx context.Context, devInfo interface{}, devType DeviceType) error
	HotplugAddDevice(ctx context.Context, devInfo interface{}, devType DeviceType) (interface{}, error)
	HotplugRemoveDevice(ctx context.Context, devInfo interface{}, devType DeviceType) (interface{}, error)
	ResizeMemory(ctx context.Context, memMB uint32, memoryBlockSizeMB uint32, probe bool) (uint32, MemoryDevice, error)
	ResizeVCPUs(ctx context.Context, vcpus uint32) (uint32, uint32, error)
	GetTotalMemoryMB(ctx context.Context) uint32
	GetVMConsole(ctx context.Context, sandboxID string) (string, string, error)
	Disconnect(ctx context.Context)
	Capabilities(ctx context.Context) types.Capabilities
	HypervisorConfig() HypervisorConfig
	GetThreadIDs(ctx context.Context) (VcpuThreadIDs, error)
	Cleanup(ctx context.Context) error
	// getPids returns a slice of hypervisor related process ids.
	// The hypervisor pid must be put at index 0.
	setConfig(config *HypervisorConfig) error
	GetPids() []int
	GetVirtioFsPid() *int
	fromGrpc(ctx context.Context, hypervisorConfig *HypervisorConfig, j []byte) error
	toGrpc(ctx context.Context) ([]byte, error)
	Check() error

	Save() hv.HypervisorState
	Load(hv.HypervisorState)

	// generate the socket to communicate the host and guest
	GenerateSocket(id string) (interface{}, error)

	// check if hypervisor supports built-in rate limiter.
	IsRateLimiterBuiltin() bool
}
