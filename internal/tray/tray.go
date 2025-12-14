package tray

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type Tray struct {
	menu *fyne.Menu
}

func NewTray(app fyne.App, onShow func()) *Tray {
	// Create tray menu
	menu := fyne.NewMenu("Shinobi Web Server",
		fyne.NewMenuItem("Show", onShow),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", app.Quit),
	)

	// Set system tray if supported
	if desk, ok := app.(desktop.App); ok {
		desk.SetSystemTrayMenu(menu)

		// Use theme icon as fallback
		desk.SetSystemTrayIcon(theme.FyneLogo())
	}

	return &Tray{
		menu: menu,
	}
}
