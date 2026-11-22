package api

import "testing"

func TestDiskUsageFromBytes(t *testing.T) {
	du := DiskUsageFromBytes(1000*BytesPerGiB, 750*BytesPerGiB, 250*BytesPerGiB)

	if got := du.UsedLabel(); got != "750.0 GiB" {
		t.Errorf("UsedLabel = %q, want 750.0 GiB", got)
	}
	if got := du.AvailLabel(); got != "250.0 GiB" {
		t.Errorf("AvailLabel = %q, want 250.0 GiB", got)
	}
	if got := du.PctLabel(); got != "75%" {
		t.Errorf("PctLabel = %q, want 75%%", got)
	}
	if got := du.UsageLevel(); got != "warn" {
		t.Errorf("UsageLevel = %q, want warn", got)
	}
}
