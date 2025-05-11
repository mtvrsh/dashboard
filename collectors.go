package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type all struct {
	Hostname   string
	Uptime     string
	Commands   []string
	DisksUsage map[string]diskUsage
	ExecAlways int
}

type diskUsage struct {
	Size       string
	Used       string
	UsePercent string
	Avail      string
}

func (du diskUsage) UsePercentStyle() template.CSS {
	p, err := strconv.ParseFloat(strings.TrimSuffix(du.UsePercent, "%"), 64)
	if err != nil {
		return ""
	}
	red := min(255, int(math.Floor((p/100)*255)))
	green := 255 - red
	return template.CSS(fmt.Sprintf("color: rgb(%d, %d, 0)", red, green))
}
func getSystemInfo(mountpoints []string) (all, error) {
	stats := all{}

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

func getDisksUsage(mountpoints []string) (map[string]diskUsage, error) {
	du := map[string]diskUsage{}
	df, err := exec.Command("df", "-h").Output()
	if err != nil {
		return du, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(df))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			return du, fmt.Errorf("invalid output from df")
		}
		for _, mount := range mountpoints {
			if mount == fields[5] {
				d := diskUsage{
					Size:       fields[1],
					Used:       fields[2],
					Avail:      fields[3],
					UsePercent: fields[4],
				}
				du[mount] = d
				continue
			}
		}
	}
	return du, nil
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

// truncated to 1m, human readable duration string
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
