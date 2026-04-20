//go:build darwin

package tools

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	getMemoryInfo = getDarwinMemoryInfo
	getDiskInfo = getDarwinDiskInfo
}

func getDarwinMemoryInfo() string {
	// Use vm_stat to get memory info on macOS
	cmd := exec.Command("vm_stat")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	content := string(out)

	// Parse page size (typically 4096 or 16384 on Apple Silicon)
	pageSizeRe := regexp.MustCompile(`page size of (\d+) bytes`)
	pageSizeMatch := pageSizeRe.FindStringSubmatch(content)
	if len(pageSizeMatch) < 2 {
		return ""
	}
	pageSize, err := strconv.ParseUint(pageSizeMatch[1], 10, 64)
	if err != nil {
		return ""
	}

	// Parse memory statistics
	var pagesFree, pagesActive, pagesInactive, pagesWired, pagesUsed uint64

	freeRe := regexp.MustCompile(`Pages free:\s+(\d+)`)
	if match := freeRe.FindStringSubmatch(content); len(match) >= 2 {
		pagesFree, _ = strconv.ParseUint(match[1], 10, 64)
	}

	activeRe := regexp.MustCompile(`Pages active:\s+(\d+)`)
	if match := activeRe.FindStringSubmatch(content); len(match) >= 2 {
		pagesActive, _ = strconv.ParseUint(match[1], 10, 64)
	}

	inactiveRe := regexp.MustCompile(`Pages inactive:\s+(\d+)`)
	if match := inactiveRe.FindStringSubmatch(content); len(match) >= 2 {
		pagesInactive, _ = strconv.ParseUint(match[1], 10, 64)
	}

	wiredRe := regexp.MustCompile(`Pages wired down:\s+(\d+)`)
	if match := wiredRe.FindStringSubmatch(content); len(match) >= 2 {
		pagesWired, _ = strconv.ParseUint(match[1], 10, 64)
	}

	// Calculate used memory
	pagesUsed = pagesActive + pagesInactive + pagesWired

	// Convert to MB
	freeMB := (pagesFree * pageSize) / 1024 / 1024
	usedMB := (pagesUsed * pageSize) / 1024 / 1024
	totalMB := freeMB + usedMB

	if totalMB == 0 {
		return ""
	}

	return fmt.Sprintf("  Total: %d MB\n  Used: %d MB\n  Free: %d MB\n", totalMB, usedMB, freeMB)
}

func getDarwinDiskInfo() string {
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
