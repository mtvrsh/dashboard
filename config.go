package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	Address          string
	Port             uint
	Commands         map[string][]string
	WatchDirUsage    []string `toml:"watch-dir-usage"`
	WatchMountpoints []string `toml:"watch-mountpoints"`
}

func (s *server) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading: %w", err)
	}
	err = toml.Unmarshal(data, &s.config)
	if err != nil {
		return fmt.Errorf("decoding: %w", err)
	}
	return nil
}

func (c config) String() string {
	return fmt.Sprintf("address = %q\nport = %v\ncommands = %v\nwatch-dir-usage = %q\nwatch-mountpoints = %q\n",
		c.Address,
		c.Port,
		pprintCommands(c.Commands),
		c.WatchDirUsage,
		c.WatchMountpoints,
	)
}

func pprintCommands(cmds map[string][]string) string {
	if len(cmds) == 0 {
		return "[]"
	}
	s := "[\n"
	for k, v := range cmds {
		s += fmt.Sprintf("  %v = %+q\n", k, v)
	}
	s += "]"
	return s
}
