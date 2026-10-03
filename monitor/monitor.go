package monitor

import (
	"fmt"
	"runtime"
)

// Initializes the monitor by determining the operating system and setting up the appropriate file paths for CPU and memory monitoring.
// Returns a fileLocationMonitor struct containing the file paths for CPU and memory monitoring.
func monitorInit(operatingSystem string) fileLocationMonitor {

	var cpuInfo string
	var coresInfo string

	// Get file paths for monitoring depending on OS
	switch operatingSystem {
	case "linux":
		cpuInfo = "/proc/stat"                 // aggregate CPU counters for utilization
		coresInfo = "/sys/devices/system/cpu/" // default Linux core info file paths
	default:
		fmt.Printf("Operating system %s is not supported\n", operatingSystem)
	}

	// Initialize default file paths for CPU and memory monitoring
	var cpuFilePaths cpuFilePaths = cpuFilePaths{
		cpuInfo:   cpuInfo,
		coresInfo: coresInfo,
	}

	var locations fileLocationMonitor = fileLocationMonitor{
		cpuFilePaths: cpuFilePaths,
		// memFilePaths: memFilePaths,
	}

	return locations

}

// Starts the monitoring process by initializing monitor and calls submonitors (CPU, Memory, etc.)
func StartMonitor() {

	var operatingSystem string = runtime.GOOS
	var locations fileLocationMonitor = monitorInit(operatingSystem)

	// Will eventually want to create a thread for each submonitor
	startCPUMonitor(locations.cpuFilePaths)
	// memInfo := getMemInfo(locations.memFilePath)

}
