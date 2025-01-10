package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	log.SetFlags(0)

	port := portFlag{}
	flag.Var(&port, "port", "port `number` (default 8080)")
	configPath := flag.String("config", "config.toml", "config `path`")
	showConfig := flag.Bool("v", false, "display current  configuration")

	flag.Parse()

	dashboard := newServer()
	if err := dashboard.loadConfig(*configPath); err != nil {
		log.Print("config: ", err)
	}

	if port.set {
		dashboard.config.Port = uint(port.port)
	}

	if *showConfig {
		fmt.Print(dashboard.config)
		os.Exit(0)
	}

	err := dashboard.serve()
	log.Fatal("server: ", err)
}

type portFlag struct {
	port uint64
	set  bool
}

func (p portFlag) String() string {
	return fmt.Sprint(p.port)
}

func (p *portFlag) Set(s string) error {
	port, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return err
	}
	p.port, p.set = port, true
	return nil
}
