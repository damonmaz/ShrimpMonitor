package monitor

import (
	"log"
	"runtime"

	"github.com/damonmaz/shrimp-monitor/api"
)

func monitorInit() CPU {

	var operatingSystem string = runtime.GOOS

	var cpu CPU = initCPUMonitor(operatingSystem)

	return cpu

}

// Starts the monitoring process by initializing monitor and calls submonitors (CPU, Memory, etc.)
func StartMonitor() {

	var cpu CPU = monitorInit()
	go startCPUMonitor(&cpu)
	var err error = api.StartAPICPU(func() any {
		return cpu.apiSnapshot()
	})

	if err != nil {
		log.Printf("CPU API server stopped: %v", err)
	}

	// Will eventually want to create a thread for each submonitor
	// startCPUMonitor(locations.cpuFilePaths)
	// memInfo := getMemInfo(locations.memFilePath)

}
