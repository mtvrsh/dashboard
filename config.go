package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type config struct {
	Address              string
	Port                 uint
	Commands             commands
	CommandTimeout       Duration `toml:"command-timeout"`
	DiskUsageMountpoints []string `toml:"disk-usage"`
	FuserCommand         []string `toml:"fuser-cmd"`
	MountpointUsers      []string `toml:"mount-users"`
}

func (c *config) loadConfig(path string) error {
	meta, err := toml.DecodeFile(path, c)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	undecoded := meta.Undecoded()
	if len(undecoded) != 0 {
		keys := strings.Trim(fmt.Sprintf("%q", undecoded), "[]")
		return fmt.Errorf("unknown config keys: %v", keys)
	}
	return nil
}

func (c config) String() string {
	return fmt.Sprintf(`address = %q
port = %d
commands = %v
command-timeout = %v
disk-usage = %q
fuser-cmd = %q
mount-users = %q`,
		c.Address,
		c.Port,
		c.Commands,
		c.CommandTimeout,
		c.DiskUsageMountpoints,
		c.FuserCommand,
		c.MountpointUsers,
	)
}

type commands map[string][]string

func (c commands) String() string {
	if len(c) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[\n")
	for k, v := range c {
		fmt.Fprintf(&b, "  %v = %+q\n", k, v)
	}
	b.WriteString("]")
	return b.String()
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
