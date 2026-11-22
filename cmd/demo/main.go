// Command demo renders screenshots and standalone demo pages from the
// dashboard template using mocked, reproducible data. It is a separate binary
// so none of this code ends up in the dashboard server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mtvrsh/dashboard/api"
	"github.com/mtvrsh/dashboard/ui"
)

const screenshotSize = "1000,600"

func main() {
	log.SetFlags(0)

	page := flag.String("demo", "", "write a self-contained demo page with mocked data to `path`")
	screenshot := flag.String("screenshot", "", "capture a lossless WebP screenshot of the demo page to `path`")
	dark := flag.Bool("dark", false, "use the dark color scheme for -screenshot")
	flag.Parse()

	if *page == "" && *screenshot == "" {
		flag.Usage()
		return
	}
	if *page != "" {
		if err := writeDemoPage(*page); err != nil {
			log.Fatal("demo: ", err)
		}
		fmt.Printf("wrote %s\n", *page)
	}
	if *screenshot != "" {
		if err := writeDemoScreenshot(*screenshot, *dark); err != nil {
			log.Fatal("screenshot: ", err)
		}
		fmt.Printf("wrote %s\n", *screenshot)
	}
}

// diskUsage builds a fixture from a size and used amount in GiB. The available
// space and usage percentage are derived, so they cannot drift apart.
func diskUsage(sizeGiB, usedGiB int64) *api.DiskUsage {
	avail := sizeGiB - usedGiB
	if avail < 0 {
		avail = 0
	}
	du := api.DiskUsageFromBytes(sizeGiB*api.BytesPerGiB, usedGiB*api.BytesPerGiB, avail*api.BytesPerGiB)
	return &du
}

// demoSystemInfo returns hand-picked, round values so rendered pages and
// screenshots are reproducible and independent of the host machine.
// It also exercises every visual state: ok/warn/crit usage, busy/idle chips,
// and a mount without usage data.
func demoSystemInfo() api.SystemInfo {
	return api.SystemInfo{
		Hostname:         "mediaserver",
		Uptime:           "21h 37m",
		Commands:         []string{"Reboot", "Shutdown"},
		CommandTimeoutMs: 10000,
		Mounts: []api.Mount{
			{Name: "/", Disk: diskUsage(500, 300)},
			{
				Name:        "/media",
				Users:       2,
				UserNames:   []string{"alice", "bob"},
				HasUserData: true,
				Disk:        diskUsage(2000, 1500),
			},
			{
				Name:        "/backup",
				Users:       1,
				UserNames:   []string{"rsync"},
				HasUserData: true,
				Disk:        diskUsage(1000, 900),
			},
			{Name: "/var/log", HasUserData: true},
		},
	}
}

// writeDemoPage renders the mocked dashboard and writes it to path.
func writeDemoPage(path string) error {
	page, err := ui.RenderStandalone(demoSystemInfo())
	if err != nil {
		return err
	}
	return os.WriteFile(path, page, 0o644)
}

// writeDemoScreenshot renders the demo page into a temporary directory,
// captures it with headless Firefox, and encodes it as lossless WebP. A private
// profile keeps Firefox isolated from any instance already running. Dark mode
// is selected via a Firefox preference, so the page's own CSS stays the single
// source of truth for colors.
func writeDemoScreenshot(path string, dark bool) error {
	page, err := ui.RenderStandalone(demoSystemInfo())
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "dashboard-screenshot-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	pagePath := filepath.Join(tmp, "demo.html")
	if err := os.WriteFile(pagePath, page, 0o600); err != nil {
		return fmt.Errorf("write temp page: %w", err)
	}
	profile := filepath.Join(tmp, "profile")
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return fmt.Errorf("create firefox profile dir: %w", err)
	}
	if dark {
		pref := `user_pref("ui.systemUsesDarkTheme", 1);` + "\n"
		if err := os.WriteFile(filepath.Join(profile, "user.js"), []byte(pref), 0o600); err != nil {
			return fmt.Errorf("write firefox prefs: %w", err)
		}
	}

	out, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Firefox can only capture PNG, so keep that intermediate in the temp dir
	// and encode the final image as lossless WebP.
	pngPath := filepath.Join(tmp, "screenshot.png")
	args := []string{
		"--headless",
		"--profile", profile,
		"--window-size", screenshotSize,
		"--screenshot", pngPath,
		"file://" + pagePath,
	}
	output, err := exec.Command("firefox", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run firefox (is it installed and in PATH?): %w: %s", err, strings.TrimSpace(string(output)))
	}

	output, err = exec.Command("cwebp", "-quiet", "-lossless", pngPath, "-o", out).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run cwebp (is it installed and in PATH?): %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
