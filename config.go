package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	Address            string
	Port               uint
	Commands           commands
	WatchMountpoints   []string           `toml:"watch-mountpoints"`
	ExecAlwaysCommands execAlwaysCommands `toml:"exec-always"`
}

type execAlwaysCommand struct {
	Cmd     []string
	Success string
	Failure string
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

	if len(s.config.ExecAlwaysCommands) != 0 {
		for _, cmd := range s.config.ExecAlwaysCommands {
			if len(cmd.Cmd) == 0 {
				panic("cmd field is required")
			}
			if cmd.Success == "" {
				panic("success field is required")
			}
			if cmd.Failure == "" {
				panic("failure field is required")
			}
		}
	}
	return nil
}

func (c config) String() string {
	return fmt.Sprintf("address = %q\nport = %v\ncommands = %v\nwatch-mountpoints = %q\nexec-always = %v\n",
		c.Address,
		c.Port,
		c.Commands,
		c.WatchMountpoints,
		c.ExecAlwaysCommands,
	)
}

type commands map[string][]string

func (cmds commands) String() string {
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

type execAlwaysCommands []execAlwaysCommand

func (cmds execAlwaysCommands) String() string {
	if len(cmds) == 0 {
		return "[]"
	}
	s := "[\n"
	for _, v := range cmds {
		s += fmt.Sprintf("  %q: %q | %q\n", v.Cmd, v.Success, v.Failure)
	}
	s += "]"
	return s
}
