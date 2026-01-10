package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/mtvrsh/dashboard/api"
)

func getAll(mountpoints, fuserCmd, dirsToWatch []string) (api.All, error) {
	stats := api.All{}
	errs := []error{}
	var err error

	stats.DisksUsage, err = getDisksUsage(mountpoints)
	if err != nil {
		errs = append(errs, fmt.Errorf("disk usage: %w", err))
	}

	stats.Hostname, err = os.Hostname()
	if err != nil {
		errs = append(errs, fmt.Errorf("hostname: %w", err))
	}

	uptime, err := getSystemUptime()
	if err != nil {
		errs = append(errs, fmt.Errorf("system uptime: %w", err))
	}
	stats.Uptime = prettyPrintDuration(uptime)

	stats.MountsUsers, err = getMountpointsUsers(fuserCmd, dirsToWatch)
	if err != nil {
		errs = append(errs, fmt.Errorf("mountpoint users: %w", err))
	}

	return stats, errors.Join(errs...)
}

func getDisksUsage(mountpoints []string) (map[string]api.DiskUsage, error) {
	diskUsage := make(map[string]api.DiskUsage, len(mountpoints))

	df, err := exec.Command("df", "-h").CombinedOutput()
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
				diskUsage[mount] = api.DiskUsage{
					Size:       fields[1],
					Used:       fields[2],
					Avail:      fields[3],
					UsePercent: fields[4],
				}
				break
			}
		}
		// exit early when we found everything
		if len(diskUsage) == len(mountpoints) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error during reading df output: %w", err)
	}

	return diskUsage, nil
}

func getSystemUptime() (time.Duration, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("reading /proc/uptime: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("got %d fields instead of 2", len(fields))
	}

	uptime, err := time.ParseDuration(fields[0] + "s")
	if err != nil {
		return 0, fmt.Errorf("parsing time: %w", err)
	}
	return uptime, nil
}

// prettyPrintDuration formats a duration as "Xd Yh Zm" truncated to minutes.
func prettyPrintDuration(d time.Duration) string {
	sign := ""
	if d < 0 {
		sign = "-"
		d = d.Abs()
	}
	totalMinutes := int(d.Minutes())

	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "0m"
	}
	return sign + strings.Join(parts, " ")
}

func getMountpointsUsers(cmd, paths []string) (map[string][]string, error) {
	mountUsers := make(map[string][]string, len(paths))

	if len(paths) == 0 {
		return mountUsers, nil
	}

	args := append(cmd[1:], "-mv")
	args = append(args, paths...)

	output, err := exec.Command(cmd[0], args...).CombinedOutput()
	// fmt.Println(string(output))
	if err != nil {
		return mountUsers, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header
	if !scanner.Scan() {
		return nil, fmt.Errorf("output too short")
	}

	currentMount := ""
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		// len = 1 is when output is wrapped because of too long mount path
		if (len(fields) == 5 || len(fields) == 1) && strings.HasSuffix(fields[0], ":") {
			currentMount = strings.TrimSuffix(fields[0], ":")
			mountUsers[currentMount] = []string{}
			// fmt.Println("current mount set to ", currentMount)
		} else if len(fields) == 4 {
			if !slices.Contains(mountUsers[currentMount], fields[0]) {
				mountUsers[currentMount] = append(mountUsers[currentMount], fields[0])
				// fmt.Println("appending", fields[0])
			}
		}
	}

	// fmt.Printf("%+v\n", mountUsers)
	// fmt.Println("len=", len(mountUsers))
	return mountUsers, nil
}
