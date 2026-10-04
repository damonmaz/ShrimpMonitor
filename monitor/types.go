package monitor

import (
	"sync"
)

// ** File Paths ** //

type cpuFilePaths struct {
	cpuStaticPath  string
	cpuDynamicPath string
	cpuCoresPath   string
}

type memFilePaths struct {
	memPath string
}

// ** Monitor Data ** //
// CPU
type CPU struct {
	mu           *sync.RWMutex
	cpuFilePaths cpuFilePaths
	cpuStatic    cpuStatic
	cpuDynamic   cpuDynamic
	error        error
}

type cpuStatic struct {
	name       string
	cores      uint
	threads    uint
	maxFreqMHz float64
}

type cpuDynamic struct {
	utilization    float64
	cpuUtilSampler cpuUtilSampler
	cores          []cpuDynamicCore
}

type cpuDynamicCore struct {
	label          int
	utilization    float64
	frequencyMHz   float64
	cpuUtilSampler cpuUtilSampler
}

type cpuUtilSampler struct {
	previousTotal uint64
	previousIdle  uint64
	initialized   bool
}

// Memory
type Memory struct {
	mu           *sync.RWMutex
	memFilePaths memFilePaths
	totalMB      uint64
	usedMB       uint64
	availableMB  uint64
	error        error
}

type Monitor struct {
	cpu CPU
	mem Memory
}

// ** API Snapshot ** //
// CPU
type cpuSnapshot struct {
	Name         string            `json:"name"`
	Cores        uint              `json:"cores"`
	Threads      uint              `json:"threads"`
	Utilization  float64           `json:"utilization_percent"`
	LogicalCores []cpuCoreSnapshot `json:"logical_cores"`
	Error        string            `json:"error,omitempty"`
}

type cpuCoreSnapshot struct {
	Label       int     `json:"label"`
	Utilization float64 `json:"utilization_percent"`
}

// Memory
type memSnapshot struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Available uint64 `json:"available"`
	Error     string `json:"error,omitempty"`
}
