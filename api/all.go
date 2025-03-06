package api

type All struct {
	Hostname   string
	Uptime     string
	Commands   []string
	DisksUsage map[string]DiskUsage
	// DirsInUse map[string]bool
	// Temperatures map[string]int
}

type DiskUsage struct {
	Size       string
	Used       string
	UsePercent string
	Avail      string
}
