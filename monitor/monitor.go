package monitor

import (
	"runtime"
)

func monitorInit() CPU {

	var operatingSystem string = runtime.GOOS

	var cpu CPU = initCPUMonitor(operatingSystem)

	return cpu

}

// Starts the monitoring process by initializing monitor and calls submonitors (CPU, Memory, etc.)
func StartMonitor() {

	var cpu CPU = monitorInit()
	startCPUMonitor(&cpu)
	// Will eventually want to create a thread for each submonitor
	// startCPUMonitor(locations.cpuFilePaths)
	// memInfo := getMemInfo(locations.memFilePath)

}
