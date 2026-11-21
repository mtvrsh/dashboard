package main

import (
	"math"
	"os"
	"os/user"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mtvrsh/dashboard/api"
)

func Test_prettyPrintDuration(t *testing.T) {
	cases := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0m"},
		{-time.Second, "0m"},
		{time.Second * 59, "0m"},
		{-time.Hour, "-1h"},
		{-(time.Hour*1 + time.Minute*30 + time.Second), "-1h 30m"},
		{time.Minute * 30, "30m"},
		{time.Minute*1440 + time.Minute*1, "1d 1m"},
		{time.Hour*1 + time.Minute*30, "1h 30m"},
		{time.Hour * 26, "1d 2h"},
		{time.Hour*1000 + time.Microsecond, "41d 16h"},
		{time.Hour * 72, "3d"},
		{time.Hour*50 + time.Minute*50 + time.Second*50, "2d 2h 50m"},
	}

	for _, c := range cases {
		got := prettyPrintDuration(c.duration)
		if got != c.expected {
			t.Errorf("prettyPrintDuration(%v) expected: %q, got: %q", c.duration, c.expected, got)
		}
	}
}

func Test_parseDiskUsage(t *testing.T) {
	const dfOutput = `Filesystem 1-blocks Used Available Capacity Mounted on
tmpfs 16508391424 21901312 16486490112 1% /
/dev/sdb1 238236467200 183839711232 54396755968 78% /media/downloads
/dev/sdc1 1073741824 1073741824 0 100% /full
tmpfs 1000000 0 1000000 0% /tmp
`

	t.Run("normalizes bytes to GiB and filters mounts", func(t *testing.T) {
		got, err := parseDiskUsage([]byte(dfOutput), []string{"/", "/media/downloads", "/full"})
		if err != nil {
			t.Fatalf("parseDiskUsage() failed: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("parseDiskUsage() returned %d mounts, want 3", len(got))
		}

		wantSize := 16508391424.0 / (1024 * 1024 * 1024)
		if math.Abs(got["/"].Size-wantSize) > 0.01 {
			t.Errorf("size = %v GiB, want %v GiB", got["/"].Size, wantSize)
		}
		if got["/"].UsePercent != 1 {
			t.Errorf("use percent = %v, want 1", got["/"].UsePercent)
		}
		if got["/"].UsageLevel() != "ok" {
			t.Errorf("usage level = %q, want ok", got["/"].UsageLevel())
		}
		if got["/media/downloads"].UsageLevel() != "warn" {
			t.Errorf("usage level = %q, want warn", got["/media/downloads"].UsageLevel())
		}
		if got["/media/downloads"].PctLabel() != "78%" {
			t.Errorf("percent label = %q, want 78%%", got["/media/downloads"].PctLabel())
		}
		if got["/full"].UsageLevel() != "crit" {
			t.Errorf("usage level = %q, want crit", got["/full"].UsageLevel())
		}
		if _, ok := got["/tmp"]; ok {
			t.Error("unrequested mount /tmp was included")
		}
	})

	t.Run("malformed input", func(t *testing.T) {
		if _, err := parseDiskUsage([]byte("garbage\n"), []string{"/"}); err == nil {
			t.Error("parseDiskUsage() succeeded with malformed input")
		}
	})

	t.Run("non-numeric percent", func(t *testing.T) {
		bad := "Filesystem 1-blocks Used Available Capacity Mounted on\ntmpfs 1000 500 500 nope% /\n"
		if _, err := parseDiskUsage([]byte(bad), []string{"/"}); err == nil {
			t.Error("parseDiskUsage() succeeded with invalid percent")
		}
	})
}

func Test_buildMounts(t *testing.T) {
	disk := api.NewDiskUsage(1024, 512, 512, 50)
	homeDisk := api.NewDiskUsage(2048, 1024, 1024, 50)
	disksUsage := map[string]api.DiskUsage{
		"/":     disk,
		"/home": homeDisk,
	}
	mountUsers := map[string][]string{
		"/":    {"alice", "bob"},
		"/tmp": {"alice"},
	}

	got := buildMounts(disksUsage, mountUsers)
	want := []api.Mount{
		{Name: "/", Users: 2, UserNames: []string{"alice", "bob"}, HasUserData: true, Disk: &disk},
		{Name: "/home", Disk: &homeDisk},
		{Name: "/tmp", Users: 1, UserNames: []string{"alice"}, HasUserData: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildMounts() = %v, want %v", got, want)
	}
}

func Test_getMountpointsUsers(t *testing.T) {
	currentUser, _ := user.Current()

	t.Run("mount users", func(t *testing.T) {
		paths := []string{"/tmp", "/dev/shm"}
		// Keep a file open on each mountpoint so fuser reports the test process
		// as a user regardless of other activity on the host.
		for _, path := range paths {
			f, err := os.CreateTemp(path, "dashboard-test-*")
			if err != nil {
				t.Fatalf("create temp file in %q: %v", path, err)
			}
			defer os.Remove(f.Name())
			defer f.Close()
		}

		got, err := getMountpointsUsers(defaultFuserCmd, paths)
		if err != nil {
			t.Fatalf("getMountpointUsers() failed: %v", err)
		}
		for _, path := range paths {
			if !slices.Contains(got[path], currentUser.Username) {
				t.Errorf("getMountpointUsers()[%q] = %v, want it to contain %q", path, got[path], currentUser.Username)
			}
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		got, err := getMountpointsUsers(defaultFuserCmd, []string{"/aaaaaaaaaaaaaaaaaaaaaaa"})
		if err == nil {
			t.Fatal("getMountpointUsers() succeeded unexpectedly")
		}
		if len(got) != 0 {
			t.Errorf("getMountpointUsers() = %v, want empty map", got)
		}
	})

	t.Run("no paths", func(t *testing.T) {
		got, err := getMountpointsUsers(defaultFuserCmd, nil)
		if err != nil {
			t.Fatalf("getMountpointUsers() failed: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("getMountpointUsers() = %v, want empty map", got)
		}
	})
}

func Test_execCommandTimeout(t *testing.T) {
	tests := []struct {
		testName    string
		commandName string
		command     string
		willTimeout time.Duration
		output      string
	}{
		{
			testName:    "command times out",
			commandName: "sleep",
			command:     "sleep 5",
			willTimeout: 50 * time.Millisecond,
			output:      "",
		},
		{
			testName:    "command succeeds",
			commandName: "echo",
			command:     "echo hello",
			willTimeout: time.Second,
			output:      "hello\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			s := server{
				config: config{
					Commands: map[string]string{
						tt.commandName: tt.command,
					},
					CommandTimeout: Duration{Duration: tt.willTimeout},
				},
			}

			output := s.execCommand(tt.commandName)

			lines := strings.SplitN(output, "\n", 2)
			if len(lines) < 2 {
				t.Fatalf("expected output with prefix line, got %q", output)
			}
			actualOutput := lines[1]

			if actualOutput != tt.output {
				t.Errorf("expected output %q, got %q", tt.output, actualOutput)
			}
		})
	}
}
