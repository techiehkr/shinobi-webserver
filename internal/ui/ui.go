package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phayes/freeport"

	"shinobi-webserver/internal/config"
	"shinobi-webserver/internal/editor"
	"shinobi-webserver/internal/server"
	"shinobi-webserver/internal/tray"
)

// SiteWidget is a custom widget for displaying site information
type SiteWidget struct {
	widget.BaseWidget
	site      *config.Site
	isRunning bool
	ui        *UI

	statusLabel *widget.Label
	nameLabel   *widget.Label
	portLabel   *widget.Label
	startBtn    *widget.Button
	stopBtn     *widget.Button
	logsBtn     *widget.Button
	editBtn     *widget.Button
	deleteBtn   *widget.Button
}

func NewSiteWidget(site *config.Site, ui *UI, isRunning bool) *SiteWidget {
	s := &SiteWidget{
		site:      site,
		isRunning: isRunning,
		ui:        ui,
	}
	s.ExtendBaseWidget(s)
	s.createWidgets()
	s.updateButtons()
	return s
}

func (s *SiteWidget) createWidgets() {
	// Create labels
	s.statusLabel = widget.NewLabel("🔴")
	s.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	s.nameLabel = widget.NewLabel(s.site.Name)
	s.nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	s.portLabel = widget.NewLabel(fmt.Sprintf("Port: %d", s.site.Port))
	s.portLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Create buttons
	s.startBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		s.ui.startSite(s.site.Name)
	})
	s.stopBtn = widget.NewButtonWithIcon("", theme.MediaStopIcon(), func() {
		s.ui.stopSite(s.site.Name)
	})
	s.logsBtn = widget.NewButtonWithIcon("", theme.DocumentIcon(), func() {
		s.ui.showLogs(s.site.Name)
	})
	s.editBtn = widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		s.ui.editSite(s.site.Name)
	})
	s.deleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		s.ui.deleteSite(s.site.Name)
	})
}

func (s *SiteWidget) updateButtons() {
	if s.isRunning {
		s.statusLabel.SetText("🟢")
		s.startBtn.Disable()
		s.stopBtn.Enable()
		s.logsBtn.Enable()
	} else {
		s.statusLabel.SetText("🔴")
		s.startBtn.Enable()
		s.stopBtn.Disable()
		s.logsBtn.Disable()
	}
}

func (s *SiteWidget) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)

	// Info container
	infoContainer := container.NewVBox(
		s.nameLabel,
		s.portLabel,
	)

	// Left container (status + info)
	leftContainer := container.NewHBox(
		container.NewCenter(s.statusLabel),
		container.NewPadded(infoContainer),
	)

	// Button container
	buttonContainer := container.NewHBox(
		s.startBtn,
		s.stopBtn,
		s.logsBtn,
		s.editBtn,
		s.deleteBtn,
	)

	// Main container
	content := container.NewBorder(
		nil,
		widget.NewSeparator(),
		leftContainer,
		buttonContainer,
	)

	return widget.NewSimpleRenderer(content)
}

func (s *SiteWidget) Update(site *config.Site, isRunning bool) {
	s.site = site
	s.isRunning = isRunning

	s.nameLabel.SetText(site.Name)
	s.portLabel.SetText(fmt.Sprintf("Port: %d", site.Port))
	s.updateButtons()
	s.Refresh()
}

type UI struct {
	app          fyne.App
	window       fyne.Window
	config       *config.Config
	servers      map[string]*server.Server
	siteList     *widget.List
	statusBar    *widget.Label
	tray         *tray.Tray
	refreshTimer *time.Timer
}

func Start(cfg *config.Config) {
	a := app.New()
	StartWithApp(a, cfg)
}

func StartWithApp(a fyne.App, cfg *config.Config) {
	ui := &UI{
		app:     a,
		config:  cfg,
		servers: make(map[string]*server.Server),
	}

	ui.window = ui.app.NewWindow("Shinobi Web Server")
	ui.window.Resize(fyne.NewSize(1000, 700))

	// Set app icon if available
	ui.setAppIcon()

	// Initialize tray
	ui.initTray()

	// Build UI
	ui.buildUI()

	// Start refresh timer
	ui.startAutoRefresh()

	// Handle window close
	ui.window.SetCloseIntercept(func() {
		ui.window.Hide()
	})

	ui.window.ShowAndRun()
}

func (u *UI) setAppIcon() {
	// Try to load icon from various locations
	iconPaths := []string{
		"assets/icons/icon.png",
		"icon.png",
		"./icon.png",
	}

	for _, path := range iconPaths {
		if _, err := os.Stat(path); err == nil {
			// For now, use theme icon
			u.app.SetIcon(theme.FyneLogo())
			break
		}
	}
}

func (u *UI) initTray() {
	u.tray = tray.NewTray(u.app, func() {
		u.window.Show()
	})
}

func (u *UI) buildUI() {
	// Create toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			u.showAddSiteDialog()
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			u.refreshSiteList()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			u.showSettingsDialog()
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			u.showHelpDialog()
		}),
	)

	// Create site list using custom SiteWidget
	u.siteList = widget.NewList(
		func() int {
			return len(u.config.Sites)
		},
		func() fyne.CanvasObject {
			// Create a dummy site for template
			dummySite := &config.Site{Name: "Template", Port: 0}
			return NewSiteWidget(dummySite, u, false)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.config.Sites) {
				return
			}

			site := &u.config.Sites[id]
			isRunning := u.servers[site.Name] != nil && u.servers[site.Name].Running

			siteWidget := obj.(*SiteWidget)
			siteWidget.Update(site, isRunning)
		},
	)

	// Create status bar
	u.statusBar = widget.NewLabel("Ready")
	u.statusBar.Alignment = fyne.TextAlignCenter

	// Create main layout
	content := container.NewBorder(
		toolbar,
		container.NewVBox(
			widget.NewSeparator(),
			u.statusBar,
		),
		nil,
		nil,
		u.siteList,
	)

	u.window.SetContent(content)
}

func (u *UI) showAddSiteDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("my-awesome-site")

	portEntry := widget.NewEntry()
	port, err := freeport.GetFreePort()
	if err != nil {
		// Fallback to config's auto port
		port, _ = u.config.GetAvailablePort()
	}
	portEntry.SetText(strconv.Itoa(port))

	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("sites/site-name")
	folderEntry.SetText("sites/")

	entryFileEntry := widget.NewEntry()
	entryFileEntry.SetText("index.html")

	// Auto-update folder based on name
	nameEntry.OnChanged = func(text string) {
		if text != "" {
			folderEntry.SetText(filepath.Join("sites", strings.ToLower(strings.ReplaceAll(text, " ", "-"))))
		}
	}

	dialog.ShowForm("Add New Site", "Create", "Cancel",
		[]*widget.FormItem{
			{Text: "Site Name", Widget: nameEntry},
			{Text: "Port", Widget: portEntry},
			{Text: "Folder", Widget: folderEntry},
			{Text: "Entry File", Widget: entryFileEntry},
		},
		func(ok bool) {
			if ok && nameEntry.Text != "" {
				port, err := strconv.Atoi(portEntry.Text)
				if err != nil {
					dialog.ShowError(fmt.Errorf("invalid port number"), u.window)
					return
				}

				// Validate port
				if !u.config.IsPortAvailable(port) {
					dialog.ShowError(fmt.Errorf("port %d is already in use", port), u.window)
					return
				}

				site := config.Site{
					Name:      nameEntry.Text,
					Folder:    folderEntry.Text,
					Port:      port,
					EntryFile: entryFileEntry.Text,
				}

				if err := u.config.AddSite(site); err != nil {
					dialog.ShowError(err, u.window)
					return
				}

				u.refreshSiteList()
				u.updateStatus(fmt.Sprintf("Site '%s' created successfully!", site.Name))
				dialog.ShowInformation("Success",
					fmt.Sprintf("Site created successfully!\n\nFolder: %s\nPort: %d\nEntry File: %s",
						site.Folder, site.Port, site.EntryFile), u.window)
			}
		}, u.window)
}

func (u *UI) startSite(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	srv, exists := u.servers[name]
	if exists && srv.Running {
		u.updateStatus(fmt.Sprintf("Site '%s' is already running", name))
		return
	}

	if !exists {
		srv = server.New(site.Port, site.Folder)
		u.servers[name] = srv
	}

	u.updateStatus(fmt.Sprintf("Starting site '%s' on port %d...", name, site.Port))

	if err := srv.Start(); err != nil {
		dialog.ShowError(err, u.window)
		u.updateStatus(fmt.Sprintf("Failed to start site '%s': %v", name, err))
		delete(u.servers, name)
		return
	}

	// Update site's last started time
	site.LastStarted = time.Now()
	u.config.UpdateSite(name, *site)

	u.refreshSiteList()
	u.updateStatus(fmt.Sprintf("Site '%s' started on http://localhost:%d", name, site.Port))
}

func (u *UI) stopSite(name string) {
	srv, exists := u.servers[name]
	if !exists || !srv.Running {
		u.updateStatus(fmt.Sprintf("Site '%s' is not running", name))
		return
	}

	u.updateStatus(fmt.Sprintf("Stopping site '%s'...", name))

	if err := srv.Stop(); err != nil {
		dialog.ShowError(err, u.window)
		u.updateStatus(fmt.Sprintf("Failed to stop site '%s': %v", name, err))
		return
	}

	u.refreshSiteList()
	u.updateStatus(fmt.Sprintf("Site '%s' stopped", name))
}

func (u *UI) openSite(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	url := fmt.Sprintf("http://localhost:%d", site.Port)

	// Try to open in default browser
	if err := editor.OpenURL(url); err != nil {
		dialog.ShowError(fmt.Errorf("failed to open browser: %v", err), u.window)
		return
	}

	u.updateStatus(fmt.Sprintf("Opened %s in browser", url))
}

func (u *UI) showLogs(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	logsDir := filepath.Join(site.Folder, "logs")
	if err := editor.OpenFolder(logsDir); err != nil {
		dialog.ShowError(fmt.Errorf("failed to open logs folder: %v", err), u.window)
		return
	}

	u.updateStatus(fmt.Sprintf("Opened logs folder for '%s'", name))
}

func (u *UI) editSite(name string) {
	site := u.config.GetSite(name)
	if site == nil {
		return
	}

	if err := editor.OpenFolder(site.Folder); err != nil {
		dialog.ShowError(fmt.Errorf("failed to open folder: %v", err), u.window)
		return
	}

	u.updateStatus(fmt.Sprintf("Opened folder for editing: %s", site.Folder))
}

func (u *UI) deleteSite(name string) {
	dialog.ShowConfirm("Delete Site",
		fmt.Sprintf("Are you sure you want to delete '%s'?\n\nThis will:\n• Stop the server if running\n• Delete all site files\n• Remove from configuration", name),
		func(ok bool) {
			if ok {
				// Stop server if running
				if srv, exists := u.servers[name]; exists {
					srv.Stop()
					delete(u.servers, name)
				}

				// Get site info before deletion
				site := u.config.GetSite(name)
				if site != nil {
					// Remove files
					if err := os.RemoveAll(site.Folder); err != nil {
						dialog.ShowError(fmt.Errorf("failed to remove files: %v", err), u.window)
						return
					}

					// Remove from config
					if err := u.config.RemoveSite(name); err != nil {
						dialog.ShowError(err, u.window)
						return
					}

					u.refreshSiteList()
					u.updateStatus(fmt.Sprintf("Site '%s' deleted", name))
				}
			}
		},
		u.window,
	)
}

func (u *UI) showSettingsDialog() {
	minPortEntry := widget.NewEntry()
	minPortEntry.SetText(strconv.Itoa(u.config.AppSettings.AutoPortMin))

	maxPortEntry := widget.NewEntry()
	maxPortEntry.SetText(strconv.Itoa(u.config.AppSettings.AutoPortMax))

	dialog.ShowForm("Settings", "Save", "Cancel",
		[]*widget.FormItem{
			{Text: "Minimum Auto Port", Widget: minPortEntry},
			{Text: "Maximum Auto Port", Widget: maxPortEntry},
		},
		func(ok bool) {
			if ok {
				minPort, err1 := strconv.Atoi(minPortEntry.Text)
				maxPort, err2 := strconv.Atoi(maxPortEntry.Text)

				if err1 != nil || err2 != nil || minPort >= maxPort {
					dialog.ShowError(fmt.Errorf("invalid port range"), u.window)
					return
				}

				u.config.AppSettings.AutoPortMin = minPort
				u.config.AppSettings.AutoPortMax = maxPort

				if err := u.config.Save(); err != nil {
					dialog.ShowError(err, u.window)
					return
				}

				u.updateStatus("Settings saved")
			}
		}, u.window)
}

func (u *UI) showHelpDialog() {
	helpText := `Shinobi Web Server - Help

Features:
• Start/Stop multiple web servers
• Auto-generated HTML templates
• Logging for each site
• System tray integration
• Easy site management

Usage:
1. Click '+' to add a new site
2. Click '▶' to start a server
3. Click '⬛' to stop a server
4. Click '📄' to edit site files
5. Click '📋' to view logs

Keyboard Shortcuts:
• Ctrl+N: New Site
• Ctrl+R: Refresh List
• Ctrl+Q: Quit

Logs are stored in: sites/[site-name]/logs/`

	dialog.ShowInformation("Help", helpText, u.window)
}

func (u *UI) refreshSiteList() {
	if u.siteList != nil {
		u.siteList.Refresh()
	}
}

func (u *UI) updateStatus(message string) {
	if u.statusBar != nil {
		u.statusBar.SetText(message)
	}

	// Clear status after 3 seconds
	if u.refreshTimer != nil {
		u.refreshTimer.Stop()
	}
	u.refreshTimer = time.AfterFunc(3*time.Second, func() {
		if u.statusBar != nil && u.statusBar.Text == message {
			u.statusBar.SetText("Ready")
		}
	})
}

func (u *UI) startAutoRefresh() {
	// Refresh UI every second to update status indicators
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				u.refreshSiteList()
			}
		}
	}()
}

func (u *UI) cleanup() {
	// Stop all servers
	for name, server := range u.servers {
		server.Stop()
		delete(u.servers, name)
	}

	// Stop refresh timer
	if u.refreshTimer != nil {
		u.refreshTimer.Stop()
	}
}
