package tray

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type Tray struct {
	menu *fyne.Menu
}

func NewTray(app fyne.App, onShow func()) *Tray {
	menu := fyne.NewMenu("Site Manager",
		fyne.NewMenuItem("Show", onShow),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", app.Quit),
	)

	if desk, ok := app.(desktop.App); ok {
		desk.SetSystemTrayMenu(menu)

	}

	return &Tray{
		menu: menu,
	}
}
