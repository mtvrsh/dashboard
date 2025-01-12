package api

type SystemStatus struct {
	Hostname string
	// DirsInUse map[string]bool
	DisksUsage map[string]DiskUsage
	// Temperatures map[string]int
	Uptime string
}

type DiskUsage struct {
	Total       string
	Used        string
	UsedPercent string
	Free        string
}
