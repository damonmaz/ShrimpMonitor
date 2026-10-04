package monitor

import (
	"fmt"
	"log"
	"runtime"

	"github.com/damonmaz/shrimp-monitor/api"
)

func monitorInit() (CPU, Memory) {

	var operatingSystem string = runtime.GOOS

	var cpu CPU = initCPUMonitor(operatingSystem)
	var mem Memory = initMemMonitor(operatingSystem)

	return cpu, mem

}

// Starts the monitoring process by initializing monitor and calls submonitors (CPU, Memory, etc.)
func StartMonitor() {

	cpu, mem := monitorInit()
	go startCPUMonitor(&cpu)
	go startMemMonitor(&mem)
	fmt.Println("monitors started")

	var errCPU error = api.StartAPI(
		func() any {
			return cpu.apiSnapshot()
		}, func() any {
			return mem.apiSnapshot()
		})

	if errCPU != nil {
		log.Printf("CPU API server stopped: %v", errCPU)
	}
}
