package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
)

type server struct {
	config config
}

func newServer() server {
	return server{config: config{Port: defaultPort}}
}

func (s *server) serve() error {
	if s.config.ServerRoot != "" {
		http.Handle("GET /", http.FileServer(http.Dir(s.config.ServerRoot)))
	}
	http.HandleFunc("GET /commands", s.commandsHandler)
	http.HandleFunc("PUT /command/{command}", s.commandHandler)
	http.HandleFunc("GET /system-status", s.systemStatusHandler)

	addr := net.JoinHostPort(fmt.Sprint(s.config.Address), fmt.Sprint(s.config.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *server) systemStatusHandler(w http.ResponseWriter, r *http.Request) {
	all, err := getSystemInfo(s.config.WatchMountpoints)
	if err != nil {
		log.Printf("collection failed %v", err)
		http.Error(w, "failed to collect data", http.StatusInternalServerError)
		return
	}

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
	commands := make([]string, 0, len(s.config.Commands))
	for k := range s.config.Commands {
		commands = append(commands, k)
	}

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

func (s *server) commandHandler(w http.ResponseWriter, r *http.Request) {
	cmdFromPath := s.config.Commands[r.PathValue("command")]
	if len(cmdFromPath) == 0 {
		http.Error(w, "command does not exist", http.StatusInternalServerError)
		return
	}

	command := exec.Command(cmdFromPath[0])
	if len(cmdFromPath) > 1 {
		command = exec.Command(cmdFromPath[0], cmdFromPath[1:]...)
	}

	output, err := command.Output()
	if err != nil {
		log.Printf("command %v failed: %v", command, err)
		http.Error(w, "failed to execute command", http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprintf(w, "command executed successfully\n%s", output)
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
