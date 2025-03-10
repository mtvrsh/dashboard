package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	Address          string
	Port             uint
	Commands         map[string][]string
	WatchMountpoints []string `toml:"watch-mountpoints"`
}

func (s *server) loadConfig(path string) error {
	meta, err := toml.DecodeFile(path, &s.config)
	if err != nil {
		return fmt.Errorf("decoding: %w", err)
	}

	undecoded := meta.Undecoded()
	if len(undecoded) != 0 {
		pretty := strings.Trim(fmt.Sprintf("%q", undecoded), "[]")
		return fmt.Errorf("unknown fields: %v", pretty)
	}
	return nil
}

func (c config) String() string {
	return fmt.Sprintf("address = %q\nport = %v\ncommands = %v\nwatch-mountpoints = %q\n",
		c.Address,
		c.Port,
		pprintCommands(c.Commands),
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
