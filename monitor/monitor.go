package monitor

import (
	"fmt"
)

func monitorInit() fileLocationMonitor {

	// Initialize default file paths for CPU and memory monitoring
	var cpuFilePaths cpuFilePaths = cpuFilePaths{
		cpuInfo:   "/proc/stat",               // aggregate CPU counters for utilization
		coresInfo: "/sys/devices/system/cpu/", // default Linux core info file paths
	}

	var locations fileLocationMonitor = fileLocationMonitor{
		cpuFilePaths: cpuFilePaths,
		// memFilePaths: memFilePaths,
	}

	return locations

}

func StartMonitor() {
	var locations fileLocationMonitor = monitorInit()

	cpuInfo := getCPUInfo(locations.cpuFilePaths)
	// memInfo := getMemInfo(locations.memFilePath)

	fmt.Printf("CPU Info: %+v\n", cpuInfo)

}
