package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	port := flag.Uint("port", 0, fmt.Sprintf("port `number` (default %d)", defaultPort))
	configPath := flag.String("config", "config.toml", "config `path`")
	showConfig := flag.Bool("v", false, "display current configuration")

	flag.Parse()

	dashboard := newServer()
	if err := dashboard.loadConfig(*configPath); err != nil {
		log.Print("config: ", err)
	}

	if *port > 0 {
		dashboard.config.Port = *port
	}

	if *showConfig {
		fmt.Print(dashboard.config)
		os.Exit(0)
	}

	err := dashboard.serve()
	log.Fatal("server: ", err)
}
