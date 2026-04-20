//go:build linux

package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func init() {
	getMemoryInfo = getLinuxMemoryInfo
	getDiskInfo = getLinuxDiskInfo
}

func getLinuxMemoryInfo() string {
	// Read /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return ""
	}

	var totalKB, availableKB uint64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			totalKB = parseMeminfoKB(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			availableKB = parseMeminfoKB(line)
		}
	}

	if totalKB == 0 {
		return ""
	}

	usedKB := totalKB - availableKB

	return fmt.Sprintf("  Total: %d MB\n  Used: %d MB\n  Available: %d MB\n",
		totalKB/1024, usedKB/1024, availableKB/1024)
}

func parseMeminfoKB(line string) uint64 {
	// Format: "MemTotal:    16384000 kB"
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			return kb
		}
	}
	return 0
}

func getLinuxDiskInfo() string {
	// Use df -h . to get disk usage for current directory
	cmd := exec.Command("df", "-h", ".")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) >= 2 {
		// First line is header, second is data
		return lines[0] + "\n" + lines[1] + "\n"
	}
	return string(out)
}
