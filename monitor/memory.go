package monitor

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/damonmaz/shrimp-monitor/lib"
)

////////////////////////////////////
// ** Memory monitor functions ** //
////////////////////////////////////

// Create a Memory struct and initialize it with static information about the memory
func initMemMonitor(operatingSystem string) Memory {
	var mem Memory = Memory{mu: &sync.RWMutex{}}

	switch operatingSystem {
	case "linux":
		mem.memFilePaths.memPath = "/proc/meminfo" // static memory info file path
	default:
		mem.error = fmt.Errorf("operating system %s is not supported", operatingSystem)
		return mem
	}

	mem.getMemInfo()

	return mem
}

// Starts the memory monitoring process
func startMemMonitor(mem *Memory) {
	lib.Ticker(func() any {
		mem.getMemInfo()
		return nil
	})
}

////////////////////////////////////
// ** Memory Struct functions ** //
///////////////////////////////////

// Reads the memory information from the file specified in memPath.
// Populates the Memory struct with the information read from the file.
func (mem *Memory) getMemInfo() {
	file, err := lib.GetFile(mem.memFilePaths.memPath)
	if err != nil {
		mem.mu.Lock()
		mem.error = err
		mem.mu.Unlock()
		return
	}
	defer file.Close()

	var totalMB uint64
	var availableMB uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// e.g. "MemTotal:       16337180 kB"
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "MemTotal":
			totalMB = value / 1024 // Convert kB to MB
		case "MemAvailable":
			availableMB = value / 1024 // Convert kB to MB
		}
	}

	if err := scanner.Err(); err != nil {
		mem.mu.Lock()
		mem.error = err
		mem.mu.Unlock()
		return
	}

	// Update all memory values together after parsing is complete.
	mem.mu.Lock()
	mem.totalMB = totalMB
	mem.availableMB = availableMB
	mem.usedMB = totalMB - availableMB
	mem.error = nil
	mem.mu.Unlock()
}

// Returns a snapshot of the current memory usage for API response.
func (mem *Memory) apiSnapshot() any {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var snapshot memSnapshot = memSnapshot{
		Total:     mem.totalMB,
		Used:      mem.usedMB,
		Available: mem.availableMB,
	}

	if mem.error != nil {
		snapshot.Error = mem.error.Error()
	}

	return snapshot
}
