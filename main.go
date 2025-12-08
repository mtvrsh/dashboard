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

	dashboard, err := newServerFromConfig(*configPath)
	if err != nil {
		log.Print(err)
	}

	if *port > 0 {
		dashboard.config.Port = *port
	}

	if *showConfig {
		fmt.Println(dashboard.config)
		os.Exit(0)
	}

	err = dashboard.serve()
	log.Fatal("server: ", err)
}
