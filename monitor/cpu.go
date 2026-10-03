package monitor

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/damonmaz/shrimp-monitor/lib"
)

/////////////////////////////////
// ** CPU monitor functions ** //
///////////////////////////////

// Create a CPU struct and initialize it with static information about the CPU
// Returns the initialized CPU struct.
func initCPUMonitor(operatingSystem string) CPU {
	var cpu CPU = CPU{}

	// Get file paths for monitoring depending on OS
	switch operatingSystem {
	case "linux":
		cpu.cpuFilePaths.cpuStaticPath = "/proc/cpuinfo"           // static CPU info file path
		cpu.cpuFilePaths.cpuDynamicPath = "/proc/stat"             // dynamic CPU info file path
		cpu.cpuFilePaths.cpuCoresPath = "/sys/devices/system/cpu/" // default Linux core info file paths
	default:
		fmt.Printf("Operating system %s is not supported\n", operatingSystem)
	}

	// Get static CPU info
	cpu.getCPUStaticInfo()

	return cpu
}

// Starts the CPU monitoring process
func startCPUMonitor(cpu *CPU) {
	// Record an initial baseline; subsequent ticker events produce utilization samples.
	cpu.getCPUDynamicInfo()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
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
			cpu.cpuStatic.threads++
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
	cpu.cpuDynamic.utilization, cpu.error = getCPUUtilization(cpu.cpuFilePaths.cpuDynamicPath, &cpu.cpuDynamic.cpuUtilSampler)
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

	// The first reading is only a baseline, so it cannot produce a delta yet.
	if !sampler.initialized {
		sampler.previousTotal = total
		sampler.previousIdle = idle
		sampler.initialized = true
		return 0, nil
	}

	// Refresh the baseline on every call so the next call measures a new interval.
	totalDelta := total - sampler.previousTotal
	idleDelta := idle - sampler.previousIdle
	sampler.previousTotal = total
	sampler.previousIdle = idle
	if totalDelta == 0 {
		return 0, nil
	}

	utilization := float64(totalDelta-idleDelta) / float64(totalDelta) * 100
	return math.Round(utilization*100) / 100, nil
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
		fields := strings.Fields(scanner.Text())
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
