package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mtvrsh/dashboard/api"
)

func getSystemInfo(mountpoints, fuserCmd, dirsToWatch []string) (api.SystemInfo, error) {
	stats := api.SystemInfo{}
	errs := []error{}
	var err error

	disksUsage, err := getDisksUsage(mountpoints)
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

	mountUsers, err := getMountpointsUsers(fuserCmd, dirsToWatch)
	if err != nil {
		errs = append(errs, fmt.Errorf("mountpoint users: %w", err))
	}
	stats.Mounts = buildMounts(disksUsage, mountUsers)

	return stats, errors.Join(errs...)
}

func getDisksUsage(mountpoints []string) (map[string]api.DiskUsage, error) {
	// -P keeps one filesystem per line, -B1 gives exact byte counts so all
	// values can be normalized to the same unit by parseDiskUsage.
	df, err := exec.Command("df", "-PB1").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseDiskUsage(df, mountpoints)
}

// parseDiskUsage parses `df -PB1` output, keeping only the requested mountpoints.
func parseDiskUsage(df []byte, mountpoints []string) (map[string]api.DiskUsage, error) {
	diskUsage := make(map[string]api.DiskUsage, len(mountpoints))

	scanner := bufio.NewScanner(bytes.NewReader(df))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			return diskUsage, fmt.Errorf("invalid output from df")
		}
		for _, mount := range mountpoints {
			if mount != fields[5] {
				continue
			}
			size, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return diskUsage, fmt.Errorf("parsing size for %q: %w", mount, err)
			}
			used, err := strconv.ParseInt(fields[2], 10, 64)
			if err != nil {
				return diskUsage, fmt.Errorf("parsing used for %q: %w", mount, err)
			}
			avail, err := strconv.ParseInt(fields[3], 10, 64)
			if err != nil {
				return diskUsage, fmt.Errorf("parsing avail for %q: %w", mount, err)
			}
			percent, err := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)
			if err != nil {
				return diskUsage, fmt.Errorf("parsing use percent for %q: %w", mount, err)
			}
			diskUsage[mount] = api.NewDiskUsage(size, used, avail, percent)
			break
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

		if (len(fields) == 5 || len(fields) == 1) && strings.HasSuffix(fields[0], ":") {
			currentMount = strings.TrimSuffix(fields[0], ":")
			mountUsers[currentMount] = []string{}
			// ignore lines before any mount header is seen
		} else if len(fields) == 4 && currentMount != "" {
			if !slices.Contains(mountUsers[currentMount], fields[0]) {
				mountUsers[currentMount] = append(mountUsers[currentMount], fields[0])
			}
		}
	}

	return mountUsers, nil
}

// buildMounts merges mount names from disk usage and mountpoint users into a
// stable, alphabetically sorted list. Mounts without an entry in mountUsers
// are flagged as having no user data so they aren't reported as idle.
func buildMounts(disksUsage map[string]api.DiskUsage, mountUsers map[string][]string) []api.Mount {
	names := make([]string, 0, len(disksUsage)+len(mountUsers))
	for name := range disksUsage {
		names = append(names, name)
	}
	for name := range mountUsers {
		if _, ok := disksUsage[name]; !ok {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	mounts := make([]api.Mount, 0, len(names))
	for _, name := range names {
		var disk *api.DiskUsage
		if usage, ok := disksUsage[name]; ok {
			disk = &usage
		}
		users, hasUserData := mountUsers[name]
		mounts = append(mounts, api.Mount{
			Name:        name,
			Users:       len(users),
			UserNames:   users,
			HasUserData: hasUserData,
			Disk:        disk,
		})
	}
	return mounts
}
