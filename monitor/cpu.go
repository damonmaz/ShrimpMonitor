package monitor

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/damonmaz/shrimp-monitor/lib"
)

/////////////////////////////////
// ** CPU monitor functions ** //
///////////////////////////////

// Create a CPU struct and initialize it with static information about the CPU
// Returns the initialized CPU struct.
func initCPUMonitor(operatingSystem string) CPU {
	var cpu CPU = CPU{mu: &sync.RWMutex{}}

	// Get file paths for monitoring depending on OS
	switch operatingSystem {
	case "linux":
		cpu.cpuFilePaths.cpuStaticPath = "/proc/cpuinfo"           // static CPU info file path
		cpu.cpuFilePaths.cpuDynamicPath = "/proc/stat"             // dynamic CPU info file path
		cpu.cpuFilePaths.cpuCoresPath = "/sys/devices/system/cpu/" // default Linux core info file paths
	default:
		cpu.error = fmt.Errorf("Operating system %s is not supported", operatingSystem)
		return cpu
	}

	// Get static CPU info
	cpu.getCPUStaticInfo()

	return cpu
}

// Starts the CPU monitoring process
func startCPUMonitor(cpu *CPU) {
	// Record an initial baseline; subsequent ticker events produce utilization samples.
	cpu.getCPUDynamicInfo()

	ticker := time.NewTicker(lib.TICKER_TIME)
	defer ticker.Stop()

	// Start a loop that will run every 500 milliseconds to get dynamic CPU info
	for range ticker.C {
		cpu.getCPUDynamicInfo()
		fmt.Printf("CPU Info: %+v\n", cpu)
	}
}

/////////////////////////////////
// ** CPU Struct functions ** //
///////////////////////////////

// Reads the static information about the CPU from the file specified in cpuStaticPath.
// Populates the CPUStatic struct with the information read from the file.
func (cpu *CPU) getCPUStaticInfo() {
	file, err := lib.GetFile(cpu.cpuFilePaths.cpuStaticPath)
	if err != nil {
		cpu.error = err
		return
	}
	defer file.Close()

	physicalCores := make(map[string]struct{})
	physicalIDs := make(map[string]struct{})
	var fallbackCores uint
	cpu.cpuDynamic.cores = nil
	scanner := bufio.NewScanner(file)
	var physicalID string
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Process the key-value pairs from the CPU info file and populate the CPUStatic struct accordingly.
		switch key {
		case "model name":
			if cpu.cpuStatic.name == "" {
				cpu.cpuStatic.name = value
			}
		case "processor":
			label, err := strconv.Atoi(value)
			if err != nil {
				cpu.error = err
				return
			}
			cpu.cpuStatic.threads++
			// Create one dynamic core entry per logical processor, initializing only its label.
			cpu.cpuDynamic.cores = append(cpu.cpuDynamic.cores, cpuDynamicCore{label: label})
		case "physical id":
			physicalID = value
			physicalIDs[value] = struct{}{}
		case "core id":
			physicalCores[physicalID+":"+value] = struct{}{}
		case "cpu cores":
			count, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				cpu.error = err
				return
			}
			if uint(count) > fallbackCores {
				fallbackCores = uint(count)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		cpu.error = err
		return
	}

	// Prefer unique socket/core pairs; fall back to the reported per-socket core count.
	if len(physicalCores) > 0 {
		cpu.cpuStatic.cores = uint(len(physicalCores))
	} else if fallbackCores > 0 {
		socketCount := len(physicalIDs)
		if socketCount == 0 {
			socketCount = 1
		}
		cpu.cpuStatic.cores = fallbackCores * uint(socketCount)
	}
}

// Reads the dynamic information about the CPU from the files specified in cpuFilePaths.
func (cpu *CPU) getCPUDynamicInfo() {
	cpu.mu.Lock()
	defer cpu.mu.Unlock()

	var path string = cpu.cpuFilePaths.cpuDynamicPath
	var err error
	cpu.cpuDynamic.utilization, err = getCPUUtilization(path, &cpu.cpuDynamic.cpuUtilSampler)
	if err != nil {
		cpu.error = err
		return
	}

	// Sample each logical core using its own previous-counter baseline.
	for index := range cpu.cpuDynamic.cores {
		var core *cpuDynamicCore = &cpu.cpuDynamic.cores[index]
		core.utilization, err = getCPUCoreUtilization(path, core.label, &core.cpuUtilSampler)
		if err != nil {
			cpu.error = err
			return
		}
	}
	cpu.error = nil
}

// Copies monitor values into a JSON-friendly snapshot without exposing sampler state.
func (cpu *CPU) apiSnapshot() cpuSnapshot {
	cpu.mu.RLock()
	defer cpu.mu.RUnlock()

	var snapshot cpuSnapshot = cpuSnapshot{
		Name:         cpu.cpuStatic.name,
		Cores:        cpu.cpuStatic.cores,
		Threads:      cpu.cpuStatic.threads,
		Utilization:  cpu.cpuDynamic.utilization,
		LogicalCores: make([]cpuCoreSnapshot, 0, len(cpu.cpuDynamic.cores)), // Preallocate slice with capacity equal to the number of logical cores
	}
	// Populate the LogicalCores slice with snapshots of each logical core's utilization.
	for _, core := range cpu.cpuDynamic.cores {
		snapshot.LogicalCores = append(snapshot.LogicalCores, cpuCoreSnapshot{
			Label:       core.label,
			Utilization: core.utilization,
		})
	}
	if cpu.error != nil {
		snapshot.Error = cpu.error.Error()
	}
	return snapshot
}

// //////////////////////////////////
// ** CPU Utilization functions ** //
// //////////////////////////////////

// Calculates the CPU utilization from the file specified in cpuInfoPath.
// Returns the CPU utilization as a float64 value and an error if any occurred during the reading process.

// Calculates utilization from the current counters and the previous sample without waiting.
func getCPUUtilization(path string, sampler *cpuUtilSampler) (float64, error) {
	total, idle, err := readCPUUtilization(path)
	if err != nil {
		return 0, err
	}
	return calculateCPUUtilization(total, idle, sampler), nil
}

// Calculates utilization from one core's counters and its previous sample.
func getCPUCoreUtilization(path string, coreNumber int, sampler *cpuUtilSampler) (float64, error) {
	total, idle, err := readCPUCoreUtilization(path, coreNumber)
	if err != nil {
		return 0, err
	}
	return calculateCPUUtilization(total, idle, sampler), nil
}

// Updates the counter baseline and returns utilization for the interval since the previous sample.
func calculateCPUUtilization(total uint64, idle uint64, sampler *cpuUtilSampler) float64 {
	// The first reading is only a baseline, so it cannot produce a delta yet.
	if !sampler.initialized {
		sampler.previousTotal = total
		sampler.previousIdle = idle
		sampler.initialized = true
		return 0
	}

	// Refresh the baseline on every call so the next call measures a new interval.
	totalDelta := total - sampler.previousTotal
	idleDelta := idle - sampler.previousIdle
	sampler.previousTotal = total
	sampler.previousIdle = idle
	if totalDelta == 0 {
		return 0
	}

	utilization := float64(totalDelta-idleDelta) / float64(totalDelta) * 100
	return math.Round(utilization*100) / 100
}

// Reads and sums the aggregate CPU counters; idle includes iowait.
func readCPUUtilization(path string) (uint64, uint64, error) {
	fields, err := readCPUStatsLine(path, "cpu")
	if err != nil {
		return 0, 0, err
	}
	return parseCPUUtilizationFields(fields)
}

// Reads and sums counters for the requested CPU core; idle includes iowait.
func readCPUCoreUtilization(path string, coreNumber int) (uint64, uint64, error) {
	if coreNumber < 0 {
		return 0, 0, fmt.Errorf("CPU core number must not be negative")
	}

	label := "cpu" + strconv.Itoa(coreNumber)
	fields, err := readCPUStatsLine(path, label)
	if err != nil {
		return 0, 0, err
	}
	return parseCPUUtilizationFields(fields)
}

// Reads the /proc/stat row matching label and returns its whitespace-separated fields.
func readCPUStatsLine(path string, label string) ([]string, error) {
	file, err := lib.GetFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Find the aggregate row (`cpu`) or the selected per-core row (`cpuN`).
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var fields []string = strings.Fields(scanner.Text())
		if len(fields) > 0 && fields[0] == label {
			return fields, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("CPU stats for %s not found in %s", label, path)
}

// Sums the standard CPU counters, treating idle and iowait as idle time.
func parseCPUUtilizationFields(fields []string) (uint64, uint64, error) {
	if len(fields) < 5 {
		return 0, 0, fmt.Errorf("invalid CPU stats")
	}

	var total uint64
	var idle uint64
	for index, field := range fields[1:] {
		if index >= 8 {
			break
		}
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		total += value
		if index == 3 || index == 4 {
			idle += value
		}
	}
	return total, idle, nil
}
