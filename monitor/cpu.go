package monitor

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/damonmaz/shrimp-monitor/lib"
)

// Reads the information about the CPU and CPU cores from the files specified in the cpuFilePaths struct. Returns a CPU struct containing the information read from the files.
func getCPUInfo(cpuFilePaths cpuFilePaths) CPU {

	var cpu CPU = CPU{}

	// Measure utilization from two /proc/stat snapshots.
	utilization, err := getCPUUtilization(cpuFilePaths.cpuInfo)
	if err != nil {
		cpu.error = err
		return cpu
	}

	cpu.util = math.Round(utilization*100) / 100

	return cpu
}

// Reads the CPU utilization from the /proc/stat file specified in cpuInfoPath. Returns the CPU utilization as a float64 value and an error if any occurred during the reading process.
func getCPUUtilization(cpuInfoPath string) (float64, error) {
	firstTotal, firstIdle, err := readCPUUtilization(cpuInfoPath)
	if err != nil {
		return 0, err
	}

	// A delay lets the cumulative counters advance so their difference is meaningful.
	time.Sleep(time.Second)

	secondTotal, secondIdle, err := readCPUUtilization(cpuInfoPath)
	if err != nil {
		return 0, err
	}
	if secondTotal <= firstTotal {
		return 0, nil
	}

	totalDelta := secondTotal - firstTotal
	idleDelta := secondIdle - firstIdle
	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100, nil
}

// Reads the CPU statistics from the file specified in cpuInfoPath (/proc/stat). Returns the total and idle CPU time as uint64 values and an error if any occurred during the reading process.
func readCPUUtilization(cpuInfoPath string) (uint64, uint64, error) {
	file, err := lib.GetFile(cpuInfoPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return 0, 0, err
	}

	// The first line of /proc/stat contains the aggregate CPU statistics.
	// It is split into fields, and the total and idle CPU time are calculated from these fields.
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("invalid aggregate CPU stats in %s", cpuInfoPath)
	}

	var total uint64
	var idle uint64

	// Sum the standard CPU counters; idle and iowait are counted as idle time.
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
