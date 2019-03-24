package api

// Resources represents the system resources avaible for LXD
// API extension: resources
type Resources struct {
	CPU    ResourcesCPU    `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory ResourcesMemory `json:"memory,omitempty" yaml:"memory,omitempty"`

	// API extension: resources_gpu
	GPU ResourcesGPU `json:"gpu,omitempty" yaml:"gpu,omitempty"`
}

// ResourcesCPUSocket represents a cpu socket on the system
// API extension: resources
type ResourcesCPUSocket struct {
	Cores          uint64 `json:"cores" yaml:"cores"`
	Frequency      uint64 `json:"frequency,omitempty" yaml:"frequency,omitempty"`
	FrequencyTurbo uint64 `json:"frequency_turbo,omitempty" yaml:"frequency_turbo,omitempty"`
	Name           string `json:"name,omitempty" yaml:"name,omitempty"`
	Vendor         string `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	Threads        uint64 `json:"threads" yaml:"threads"`

	// API extension: resources_cpu_socket
	Socket uint64 `json:"socket" yaml:"socket"`

	// API extension: resources_numa
	NUMANode uint64 `json:"numa_node" yaml:"numa_node"`
}

// ResourcesCPU represents the cpu resources available on the system
// API extension: resources
type ResourcesCPU struct {
	Sockets []ResourcesCPUSocket `json:"sockets" yaml:"sockets"`
	Total   uint64               `json:"total" yaml:"total"`
}

// ResourcesGPUCardNvidia represents additional information for NVIDIA GPUs
// API extension: resources_gpu
type ResourcesGPUCardNvidia struct {
	CUDAVersion  string `json:"cuda_version" yaml:"cuda_version"`
	NVRMVersion  string `json:"nvrm_version" yaml:"nvrm_version"`
	Brand        string `json:"brand" yaml:"brand"`
	Model        string `json:"model" yaml:"model"`
	UUID         string `json:"uuid" yaml:"uuid"`
	Architecture string `json:"architecture" yaml:"architecture"`
}

// ResourcesGPUCard represents a GPU card on the system
// API extension: resources_gpu
type ResourcesGPUCard struct {
	Driver        string                  `json:"driver" yaml:"driver"`
	DriverVersion string                  `json:"driver_version" yaml:"driver_version"`
	ID            uint64                  `json:"id" yaml:"id"`
	Nvidia        *ResourcesGPUCardNvidia `json:"nvidia,omitempty" yaml:"nvidia,omitempty"`
	PCIAddress    string                  `json:"pci_address" yaml:"pci_address"`
	Vendor        string                  `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	VendorID      string                  `json:"vendor_id,omitempty" yaml:"vendor_id,omitempty"`
	Product       string                  `json:"product,omitempty" yaml:"product,omitempty"`
	ProductID     string                  `json:"product_id,omitempty" yaml:"product_id,omitempty"`

	// API extension: resources_numa
	NUMANode uint64 `json:"numa_node" yaml:"numa_node"`
}

// ResourcesGPU represents the GPU resources available on the system
// API extension: resources_gpu
type ResourcesGPU struct {
	Cards []ResourcesGPUCard `json:"cards" yaml:"cards"`
	Total uint64             `json:"total" yaml:"total"`
}

// ResourcesMemory represents the memory resources available on the system
// API extension: resources
type ResourcesMemory struct {
	Used  uint64 `json:"used" yaml:"used"`
	Total uint64 `json:"total" yaml:"total"`
}

// ResourcesStoragePool represents the resources available to a given storage pool
// API extension: resources
type ResourcesStoragePool struct {
	Space  ResourcesStoragePoolSpace  `json:"space,omitempty" yaml:"space,omitempty"`
	Inodes ResourcesStoragePoolInodes `json:"inodes,omitempty" yaml:"inodes,omitempty"`
}

// ResourcesStoragePoolSpace represents the space available to a given storage pool
// API extension: resources
type ResourcesStoragePoolSpace struct {
	Used  uint64 `json:"used,omitempty" yaml:"used,omitempty"`
	Total uint64 `json:"total" yaml:"total"`
}

// ResourcesStoragePoolInodes represents the inodes available to a given storage pool
// API extension: resources
type ResourcesStoragePoolInodes struct {
	Used  uint64 `json:"used" yaml:"used"`
	Total uint64 `json:"total" yaml:"total"`
}
