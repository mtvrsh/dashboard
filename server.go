package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"slices"
	"strings"
)

const defaultPort = 8080

//go:embed index.template
var index string

//go:embed style.css
var css []byte

type server struct {
	config config
	main   *template.Template
}

func newServer() server {
	return server{config: config{Port: defaultPort}}
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
		log.Printf("collection failed %v", err)
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
	cmd := s.execCommand(r.PathValue("command"))
	fmt.Fprintf(w, "<!DOCTYPE html><pre id=command-output>%s</pre>", cmd)
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
