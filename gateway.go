package main

import (
	"log"
	"github.com/serg666/gateway/config"
)

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatalf("can not parse flags due to: %v", err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatalf("can not get new config due to: %v", err)
	}

	// Run the server
	cfg.RunServer(MakeHandler)
}
