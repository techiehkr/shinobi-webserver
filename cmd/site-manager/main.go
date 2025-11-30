package main

import (
	"shinobi-webserver/internal/config"
	"shinobi-webserver/internal/ui"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		cfg = config.NewDefault()
	}

	// Start UI
	ui.Start(cfg)
}
