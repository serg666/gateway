package main

import "log"

func main() {
	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatalf("can not parse flags due to: %v", err)
	}

	cfg, err := NewConfig(cfgPath)
	if err != nil {
		log.Fatalf("can not get new config due to: %v", err)
	}

	// Run the server
	cfg.RunServer()
}
