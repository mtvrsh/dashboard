package main

import (
	"strings"
	"testing"

	"github.com/mtvrsh/dashboard/ui"
)

func TestRenderStandalone(t *testing.T) {
	info := demoSystemInfo()
	page, err := ui.RenderStandalone(info)
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)

	// Assert against the data model, not copies of fixture strings.
	for _, want := range []string{info.Hostname, info.Uptime, "<style>"} {
		if !strings.Contains(html, want) {
			t.Errorf("page is missing %q", want)
		}
	}
	for _, command := range info.Commands {
		if !strings.Contains(html, command) {
			t.Errorf("page is missing command %q", command)
		}
	}
	for _, mount := range info.Mounts {
		if !strings.Contains(html, mount.Name) {
			t.Errorf("page is missing mount %q", mount.Name)
		}
		if mount.Disk == nil {
			continue
		}
		for _, want := range []string{
			mount.Disk.UsedLabel() + " of " + mount.Disk.SizeLabel(),
			mount.Disk.AvailLabel() + " free",
			mount.Disk.PctLabel(),
		} {
			if !strings.Contains(html, want) {
				t.Errorf("mount %q is missing %q", mount.Name, want)
			}
		}
	}
	for _, class := range []string{
		"usage--ok", "usage--warn", "usage--crit",
		"is-busy", "is-idle", "mount-empty",
	} {
		if !strings.Contains(html, class) {
			t.Errorf("page is missing visual state %q", class)
		}
	}

	if strings.Contains(html, `<link rel="stylesheet"`) {
		t.Error("standalone page still links to an external stylesheet")
	}
}

func TestDiskUsageIsConsistent(t *testing.T) {
	du := *diskUsage(1000, 750)
	if got := du.Avail; got != 250 {
		t.Errorf("Avail = %v, want 250", got)
	}
	if got := du.PctLabel(); got != "75%" {
		t.Errorf("PctLabel = %q, want 75%%", got)
	}
	if got := du.UsageLevel(); got != "warn" {
		t.Errorf("UsageLevel = %q, want warn", got)
	}
}
