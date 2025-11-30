package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Site struct {
	Name        string    `json:"name"`
	Folder      string    `json:"folder"`
	Port        int       `json:"port"`
	EntryFile   string    `json:"entryFile"`
	LastStarted time.Time `json:"lastStarted"`
	Running     bool      `json:"-"`
}

type AppSettings struct {
	AutoPortMin int `json:"autoPortMin"`
	AutoPortMax int `json:"autoPortMax"`
}

type Config struct {
	Sites       []Site      `json:"sites"`
	AppSettings AppSettings `json:"appSettings"`
}

const configFile = "config.json"

func NewDefault() *Config {
	return &Config{
		Sites: []Site{},
		AppSettings: AppSettings{
			AutoPortMin: 8000,
			AutoPortMax: 9000,
		},
	}
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func (c *Config) AddSite(site Site) error {
	// Create site folder
	if err := os.MkdirAll(site.Folder, 0755); err != nil {
		return err
	}

	// Create logs folder
	logsDir := filepath.Join(site.Folder, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Create default index file
	indexPath := filepath.Join(site.Folder, site.EntryFile)
	defaultContent := `<!DOCTYPE html>
<html>
<head>
    <title>` + site.Name + `</title>
</head>
<body>
    <h1>Welcome to ` + site.Name + `</h1>
    <p>Your site is running on port ` + fmt.Sprintf("%d", site.Port) + `</p>
</body>
</html>`

	if err := os.WriteFile(indexPath, []byte(defaultContent), 0644); err != nil {
		return err
	}

	c.Sites = append(c.Sites, site)
	return c.Save()
}

func (c *Config) RemoveSite(name string) error {
	for i, site := range c.Sites {
		if site.Name == name {
			c.Sites = append(c.Sites[:i], c.Sites[i+1:]...)
			return c.Save()
		}
	}
	return nil
}

func (c *Config) UpdateSite(name string, updated Site) error {
	for i, site := range c.Sites {
		if site.Name == name {
			c.Sites[i] = updated
			return c.Save()
		}
	}
	return nil
}

func (c *Config) GetSite(name string) *Site {
	for i, site := range c.Sites {
		if site.Name == name {
			return &c.Sites[i]
		}
	}
	return nil
}
