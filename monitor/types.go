package monitor

import (
	"time"
)

// ** File Paths ** //

type cpuFilePaths struct {
	cpuStaticPath  string
	cpuDynamicPath string
	cpuCoresPath   string
}

type memFilePaths struct {
}

// ** Monitor Data ** //
// CPU
type CPU struct {
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
	error error
}

type Monitor struct {
	cpu CPU
	mem Memory
}

// Constants
var TICKER_TIME = 500 * time.Millisecond
