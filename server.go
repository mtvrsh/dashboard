package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"time"
)

const (
	defaultPort    = 8080
	defaultTimeout = 10 * time.Second
)

//go:embed index.template
var index string

//go:embed style.css
var css []byte

type server struct {
	config config
	main   *template.Template
}

func newServer() server {
	return server{config: config{
		Port:           defaultPort,
		CommandTimeout: Duration{Duration: defaultTimeout},
	}}
}

func newServerFromConfig(path string) (server, error) {
	s := newServer()
	return s, s.config.loadConfig(path)
}

func (s *server) serve() error {
	s.main = template.Must(template.New("index.template").Parse(index))

	http.HandleFunc("GET /", s.mainHandler)
	http.HandleFunc("POST /command/{command}", s.commandHandler)
	http.HandleFunc("GET /style.css", s.styleHandler)

	addr := net.JoinHostPort(s.config.Address, fmt.Sprint(s.config.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *server) mainHandler(w http.ResponseWriter, r *http.Request) {
	all, err := getSystemInfo(s.config.WatchMountpoints)
	if err != nil {
		log.Printf("failed to get system info: %v", err)
		http.Error(w, "Failed to collect data", http.StatusInternalServerError)
		return
	}
	all.Commands = s.getCommands()

	err = s.main.Execute(w, all)
	if err != nil {
		log.Print(err)
	}
}

func (s *server) commandHandler(w http.ResponseWriter, r *http.Request) {
	command := r.PathValue("command")
	log.Printf("client %q requested execution of command %q", r.RemoteAddr, command)
	output := s.execCommand(command)

	fmt.Fprintf(w, "<!DOCTYPE html><pre id=command-output>%s</pre>", output)
}

func (s *server) styleHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "immutable, max-age=86400")
	w.Header().Set("Content-Type", "text/css")
	_, err := w.Write(css)
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
	cmdFromPath := s.config.Commands[command]
	if len(cmdFromPath) == 0 {
		log.Printf("command not declared: %q", command)
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.CommandTimeout.Duration)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdFromPath[0], cmdFromPath[1:]...)
	cmdStr := strings.Join(cmd.Args, " ")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed: %s; output: %q", cmdStr, err, truncate(output, 80, "[truncated]"))

		if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("%q exceeded deadline (%v)\n", command, s.config.CommandTimeout)
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
