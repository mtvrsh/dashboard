package api

import (
	"fmt"
	"html/template"
	"math"
)

type SystemInfo struct {
	Hostname string
	Uptime   string
	Commands []string
	Mounts   []Mount
	// CommandTimeoutMs is the configured command timeout in milliseconds. The
	// UI uses it to reset a pending command whose response never arrives.
	CommandTimeoutMs int64
	// Temperatures map[string]int
}

// Mount aggregates a mountpoint's occupancy and disk usage for display.
type Mount struct {
	Name string
	// Users is the number of distinct users currently using the mountpoint.
	Users int
	// UserNames lists the users currently using the mountpoint.
	UserNames []string
	// HasUserData reports whether fuser was run for this mountpoint. When
	// false, Users and UserNames are unknown rather than zero.
	HasUserData bool
	// Disk holds disk usage info, nil when no data exists.
	Disk *DiskUsage
}

// DiskUsage holds disk usage values normalized to GiB and percent.
type DiskUsage struct {
	Size       float64
	Used       float64
	Avail      float64
	UsePercent float64
}

const bytesPerGiB = 1024 * 1024 * 1024

// NewDiskUsage converts raw byte counts from df into GiB.
func NewDiskUsage(size, used, avail int64, usePercent float64) DiskUsage {
	return DiskUsage{
		Size:       float64(size) / bytesPerGiB,
		Used:       float64(used) / bytesPerGiB,
		Avail:      float64(avail) / bytesPerGiB,
		UsePercent: usePercent,
	}
}

func (du DiskUsage) SizeLabel() string  { return fmt.Sprintf("%.1f GiB", du.Size) }
func (du DiskUsage) UsedLabel() string  { return fmt.Sprintf("%.1f GiB", du.Used) }
func (du DiskUsage) AvailLabel() string { return fmt.Sprintf("%.1f GiB", du.Avail) }
func (du DiskUsage) PctLabel() string   { return fmt.Sprintf("%.0f%%", du.UsePercent) }

// BarStyle returns an inline width clamped to 0–100%.
func (du DiskUsage) BarStyle() template.CSS {
	p := math.Min(100, math.Max(0, du.UsePercent))
	return template.CSS(fmt.Sprintf("width: %.1f%%", p))
}

// UsageLevel maps a usage percentage to a severity class.
func (du DiskUsage) UsageLevel() string {
	switch {
	case du.UsePercent >= 85:
		return "crit"
	case du.UsePercent >= 70:
		return "warn"
	default:
		return "ok"
	}
}
