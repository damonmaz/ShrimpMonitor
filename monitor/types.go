package monitor

// File Paths
type fileLocationMonitor struct {
	cpuFilePaths cpuFilePaths
	memFilePaths memFilePaths
}

type cpuFilePaths struct {
	cpuInfo   string
	coresInfo string
}

type memFilePaths struct {
}

// Monitor Data
type CPU struct {
	error error
	util  float64
}

type Memory struct {
	error error
}

type Monitor struct {
	cpu CPU
	mem Memory
}

// Samplers
type CPUUtilSampler struct {
	total         uint64
	idle          uint64
	previousTotal uint64
	previousIdle  uint64
}
