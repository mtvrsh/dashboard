package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"os/exec"
	"slices"
	"time"

	"github.com/mtvrsh/dashboard/ui"
)

const (
	defaultAddress = "127.0.0.1"
	defaultPort    = 8080
	defaultTimeout = 10 * time.Second
)

var defaultFuserCmd = []string{"fuser"}

type server struct {
	config config
}

func newServer() server {
	return server{config: config{
		Address:        defaultAddress,
		Port:           defaultPort,
		CommandTimeout: Duration{Duration: defaultTimeout},
		FuserCommand:   defaultFuserCmd,
	}}
}

func newServerFromConfig(path string) (server, error) {
	s := newServer()
	return s, s.config.loadConfig(path)
}

func (s *server) serve() error {
	http.HandleFunc("GET /", s.mainHandler)
	http.HandleFunc("POST /command/{command}", s.commandHandler)
	http.HandleFunc("GET /style.css", s.styleHandler)

	addr := net.JoinHostPort(s.config.Address, fmt.Sprint(s.config.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *server) mainHandler(w http.ResponseWriter, r *http.Request) {
	info, err := getSystemInfo(s.config.DiskUsageMountpoints, s.config.FuserCommand, s.config.MountpointUsers)
	if err != nil {
		log.Printf("failed to get system info: %v", err)
	}
	info.Commands = s.getCommands()
	info.CommandTimeoutMs = s.config.CommandTimeout.Milliseconds()

	err = ui.Render(w, info)
	if err != nil {
		log.Print(err)
	}
}

func (s *server) commandHandler(w http.ResponseWriter, r *http.Request) {
	command := r.PathValue("command")
	log.Printf("client %q requested execution of command %q", r.RemoteAddr, command)
	output := s.execCommand(command)

	fmt.Fprintf(w, "<!DOCTYPE html><pre id=command-output>%s</pre>", html.EscapeString(output))
}

func (s *server) styleHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "immutable, max-age=86400")
	w.Header().Set("Content-Type", "text/css")
	_, err := w.Write(ui.CSS)
	if err != nil {
		log.Print(err)
	}
}

func (s *server) getCommands() []string {
	commands := make([]string, 0, len(s.config.Commands))
	for k := range s.config.Commands {
		commands = append(commands, k)
	}
	slices.Sort(commands)
	return commands
}

func (s *server) execCommand(command string) string {
	cmdStr := s.config.Commands[command]
	if len(cmdStr) == 0 {
		log.Printf("command not declared: %q", command)
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.CommandTimeout.Duration)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q=%q failed: %s; output: %q", command, cmd.Args, err, truncate(output, 80, "[truncated]"))

		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("command %q=%q exceeded deadline (%v)", command, cmd.Args, s.config.CommandTimeout)
		}
	}

	return fmt.Sprintf("$ %v\n%s", cmdStr, truncate(output, 2000, "..."))
}

func truncate(b []byte, limit int, ellip string) string {
	runes := []rune(string(b))
	if len(runes) > limit {
		return string(runes[:limit]) + ellip
	}
	return string(b)
}
