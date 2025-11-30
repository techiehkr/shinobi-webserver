package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/phayes/freeport"

	"shinobi-webserver/internal/config"
	"shinobi-webserver/internal/server"
)

type UI struct {
	app      fyne.App
	window   fyne.Window
	config   *config.Config
	servers  map[string]*server.Server
	siteList *widget.List
}

func Start(cfg *config.Config) {
	ui := &UI{
		app:     app.New(),
		config:  cfg,
		servers: make(map[string]*server.Server),
	}

	ui.window = ui.app.NewWindow("Site Manager")
	ui.window.Resize(fyne.NewSize(900, 600))

	ui.buildUI()
	ui.window.ShowAndRun()
}

func (u *UI) buildUI() {
	// Site list
	u.siteList = widget.NewList(
		func() int {
			return len(u.config.Sites)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Site Name"),
				widget.NewButton("Start", nil),
				widget.NewButton("Stop", nil),
				widget.NewButton("Edit", nil),
				widget.NewButton("Delete", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			site := u.config.Sites[id]
			box := obj.(*fyne.Container)

			label := box.Objects[0].(*widget.Label)
			label.SetText(fmt.Sprintf("%s (Port: %d)", site.Name, site.Port))

			startBtn := box.Objects[1].(*widget.Button)
			startBtn.OnTapped = func() { u.startSite(site.Name) }

			stopBtn := box.Objects[2].(*widget.Button)
			stopBtn.OnTapped = func() { u.stopSite(site.Name) }

			editBtn := box.Objects[3].(*widget.Button)
			editBtn.OnTapped = func() { u.editSite(site.Name) }

			deleteBtn := box.Objects[4].(*widget.Button)
			deleteBtn.OnTapped = func() { u.deleteSite(site.Name) }
		},
	)

	// Add site button
	addBtn := widget.NewButton("Add New Site", func() {
		u.showAddSiteDialog()
	})

	// Main layout
	content := container.NewBorder(
		addBtn,
		nil,
		nil,
		nil,
		u.siteList,
	)

	u.window.SetContent(content)
}

func (u *UI) showAddSiteDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Project Name")

	portEntry := widget.NewEntry()
	port, _ := freeport.GetFreePort()
	portEntry.SetText(strconv.Itoa(port))
	portEntry.SetPlaceHolder("Port")

	entryFileEntry := widget.NewEntry()
	entryFileEntry.SetText("index.html")
	entryFileEntry.SetPlaceHolder("Entry File")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Port", Widget: portEntry},
			{Text: "Entry File", Widget: entryFileEntry},
		},
	}

	dialog.ShowCustomConfirm("Add New Site", "Create", "Cancel", form, func(ok bool) {
		if ok && nameEntry.Text != "" {
			port, _ := strconv.Atoi(portEntry.Text)
			site := config.Site{
				Name:      nameEntry.Text,
				Folder:    filepath.Join("sites", nameEntry.Text),
				Port:      port,
				EntryFile: entryFileEntry.Text,
			}

			if err := u.config.AddSite(site); err != nil {
				dialog.ShowError(err, u.window)
				return
			}

			u.siteList.Refresh()
			dialog.ShowInformation("Success", "Site created successfully!", u.window)
		}
	}, u.window)
}

func (u *UI) startSite(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	if _, exists := u.servers[name]; exists {
		dialog.ShowInformation("Info", "Site is already running", u.window)
		return
	}

	srv := server.New(site.Port, site.Folder)
	if err := srv.Start(); err != nil {
		dialog.ShowError(err, u.window)
		return
	}

	u.servers[name] = srv
	dialog.ShowInformation("Success", fmt.Sprintf("Site started on http://localhost:%d", site.Port), u.window)
}

func (u *UI) stopSite(name string) {
	srv, exists := u.servers[name]
	if !exists {
		dialog.ShowInformation("Info", "Site is not running", u.window)
		return
	}

	if err := srv.Stop(); err != nil {
		dialog.ShowError(err, u.window)
		return
	}

	delete(u.servers, name)
	dialog.ShowInformation("Success", "Site stopped", u.window)
}

func (u *UI) editSite(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	// Open file browser
	dialog.ShowInformation("Edit", fmt.Sprintf("Opening folder: %s", site.Folder), u.window)
}

func (u *UI) deleteSite(name string) {
	dialog.ShowConfirm("Delete Site",
		"Are you sure you want to delete this site? This will remove all files.",
		func(ok bool) {
			if ok {
				site := u.config.GetSite(name)
				if site != nil {
					os.RemoveAll(site.Folder)
					u.config.RemoveSite(name)
					u.siteList.Refresh()
				}
			}
		},
		u.window,
	)
}
