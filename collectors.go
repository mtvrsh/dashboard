package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mtvrsh/dashboard/api"
)

func getSystemInfo(mountpoints []string) (api.SystemStatus, error) {
	stats := api.SystemStatus{}

	du, err := getDisksUsage(mountpoints)
	if err != nil {
		return stats, fmt.Errorf("disk usage: %w", err)
	}
	stats.DisksUsage = du

	stats.Hostname, err = os.Hostname()
	if err != nil {
		return stats, fmt.Errorf("hostname: %w", err)
	}

	uptime, err := getSystemUptime()
	if err != nil {
		return stats, fmt.Errorf("system uptime: %w", err)
	}
	stats.Uptime = prettyPrintDuration(uptime)

	return stats, nil
}

func getDisksUsage(mountpoints []string) (map[string]api.DiskUsage, error) {
	diskUsage := map[string]api.DiskUsage{}
	df, err := exec.Command("df", "-h").Output()
	if err != nil {
		return diskUsage, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(df))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			return diskUsage, fmt.Errorf("invalid output from df")
		}
		for _, mount := range mountpoints {
			if mount == fields[5] {
				du := api.DiskUsage{
					Total:       fields[1],
					Used:        fields[2],
					UsedPercent: fields[4],
					Free:        fields[3],
				}
				diskUsage[mount] = du
				continue
			}
		}
	}
	return diskUsage, nil
}

func getSystemUptime() (time.Duration, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("reading /proc/uptime: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("got %d fields instead of 2", len(fields))
	}

	uptime, err := time.ParseDuration(fields[0] + "s")
	if err != nil {
		return 0, fmt.Errorf("parsing time: %w", err)
	}
	return uptime, nil
}

// truncated to 1m, human readable duraton string
func prettyPrintDuration(d time.Duration) string {
	result := ""
	if d < 0 {
		d = d.Abs()
		result = "-"
	}
	totalMinutes := int(d.Minutes())

	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := (totalMinutes % 60)

	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	return strings.TrimSuffix(result, " ")
}
