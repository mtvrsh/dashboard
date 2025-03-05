package api

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"
)

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

func (du DiskUsage) UsePercentStyle() template.CSS {
	p, err := strconv.ParseFloat(strings.TrimSuffix(du.UsePercent, "%"), 64)
	if err != nil {
		return ""
	}
	red := min(255, int(math.Floor((p/100)*255)))
	green := 255 - red
	return template.CSS(fmt.Sprintf("color: rgb(%d, %d, 0)", red, green))
}
