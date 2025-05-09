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
	"strconv"
	"strings"
	"time"
)

const defaultPort = 8080

//go:embed index.template
var index string

//go:embed style.css
var css []byte

type server struct {
	config         config
	commands       []string
	alwaysCommands int
	main           *template.Template
}

func newServer() server {
	return server{config: config{Port: defaultPort}}
}

func (s *server) serve() error {
	s.commands = s.getCommands()

	s.alwaysCommands = len(s.config.ExecAlwaysCommands)

	s.main = template.Must(template.New("index.template").Parse(index))

	http.HandleFunc("GET /", s.mainHandler)
	http.HandleFunc("POST /command/{command}", s.commandHandler)
	http.HandleFunc("GET /exec-always/{id}", s.execAlwaysCommandHandler)
	http.HandleFunc("GET /style.css", s.styleHandler)

	addr := net.JoinHostPort(s.config.Address, fmt.Sprint(s.config.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *server) mainHandler(w http.ResponseWriter, _ *http.Request) {
	all, err := getSystemInfo(s.config.WatchMountpoints)
	if err != nil {
		log.Printf("collection failed %v", err)
		http.Error(w, "Failed to collect data", http.StatusInternalServerError)
		return
	}
	all.Commands = s.commands
	all.ExecAlways = s.alwaysCommands

	err = s.main.Execute(w, all)
	if err != nil {
		log.Print(err)
	}
}

func (s *server) commandHandler(w http.ResponseWriter, r *http.Request) {
	cmd := s.execCommand(r.PathValue("command"))
	fmt.Fprintf(w, "<!DOCTYPE html><pre id=command-output>%s</pre>", cmd)
}

func (s *server) execAlwaysCommandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	index, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || index >= len(s.config.ExecAlwaysCommands) || index < 0 {
		log.Printf("invalid execAlwaysCommand id requested: %q", r.PathValue("id"))
		http.Error(w, "Invalid command index", http.StatusInternalServerError)
		return
	}
	cmd := s.config.ExecAlwaysCommands[index]
	switch status := execCommandQuiet(cmd.Cmd); status {
	case ok:
		fmt.Fprintf(w, "%s", cmd.Success)
	case exitError:
		fmt.Fprintf(w, "%s", cmd.Failure)
	case timeoutExceeded:
		fmt.Fprintf(w, "Command %q timed out", strings.Join(cmd.Cmd, " "))
	case otherError:
		fmt.Fprintf(w, "Failed to execute %q", strings.Join(cmd.Cmd, " "))
	}
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

	cmd := exec.Command(cmdFromPath[0])
	if len(cmdFromPath) > 1 {
		cmd = exec.Command(cmdFromPath[0], cmdFromPath[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed: %s: %s", command, err, output)
	}
	return fmt.Sprintf("$ %v\n%s", strings.Join(cmd.Args, " "),
		truncate(string(output), 2000, "..."))
}

func truncate(s string, limit int, ellip string) string {
	runes := []rune(s)
	if len(runes) > limit {
		return string(runes[:limit]) + ellip
	}
	return s
}

type exitStatus uint

const (
	undefined exitStatus = iota
	ok
	exitError
	otherError
	timeoutExceeded
)

func execCommandQuiet(command []string) exitStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := &exec.Cmd{}
	if len(command) == 1 {
		cmd = exec.CommandContext(ctx, command[0])
	} else {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	}

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	log.Printf("%q took %v; error: %v", cmd.String(), elapsed, err)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return timeoutExceeded
		}
		var exerr *exec.ExitError
		// if _, isExitError := err.(*exec.ExitError); isExitError {
		if errors.As(err, &exerr) {
			return exitError
		}
		log.Printf("command %q failed: %s", command, err)
		return otherError
	}
	return ok
}
