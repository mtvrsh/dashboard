package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"slices"
	"strings"
)

const defaultPort = 8080

//go:embed static/*
var content embed.FS

type server struct {
	config config
}

func newServer() server {
	return server{config: config{Port: defaultPort}}
}

func (s *server) serve() error {
	http.Handle("GET /", staticHandler())
	http.HandleFunc("GET /commands", s.commandsHandler)
	http.HandleFunc("PUT /command/{command}", s.commandHandler)
	http.HandleFunc("GET /all", s.allHandler)

	addr := net.JoinHostPort(s.config.Address, fmt.Sprint(s.config.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *server) allHandler(w http.ResponseWriter, r *http.Request) {
	all, err := getSystemInfo(s.config.WatchMountpoints)
	if err != nil {
		log.Printf("collection failed %v", err)
		http.Error(w, "Failed to collect data", http.StatusInternalServerError)
		return
	}
	all.Commands = s.getCommands()

	data, err := json.Marshal(all)
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (s *server) commandsHandler(w http.ResponseWriter, r *http.Request) {
	commands := s.getCommands()
	data, err := json.Marshal(commands)
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("failed to write response: %v", err)
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

func (s *server) commandHandler(w http.ResponseWriter, r *http.Request) {
	const defaultSize = 256
	cmdFromPath := s.config.Commands[r.PathValue("command")]
	if len(cmdFromPath) == 0 {
		http.Error(w, "Command does not exist", http.StatusInternalServerError)
		return
	}

	command := exec.Command(cmdFromPath[0])
	if len(cmdFromPath) > 1 {
		command = exec.Command(cmdFromPath[0], cmdFromPath[1:]...)
	}

	rawOutput, err := command.CombinedOutput()
	output := fmt.Sprintf("$ %v\n%s", strings.Join(command.Args, " "),
		truncate(string(rawOutput), defaultSize, "..."))
	if err != nil {
		log.Printf("command %q failed: %s: %s", command, err, rawOutput)
		http.Error(w, strings.TrimSpace(output), http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprint(w, output)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func staticHandler() http.Handler {
	files, err := fs.Sub(content, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServerFS(files)
}

func truncate(s string, max int, ellip string) string {
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + ellip
	}
	return s
}
