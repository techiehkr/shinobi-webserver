package main

import (
	"shinobi-webserver/internal/config"
	"shinobi-webserver/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.NewDefault()
	}

	// Create Fyne app
	a := app.New()

	// Start UI
	ui.StartWithApp(a, cfg)
}
