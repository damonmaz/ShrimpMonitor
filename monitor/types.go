package monitor

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
