package config

import (
	"encoding/json"
	"fmt"
	"net"
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
	// Validate port
	if !c.IsPortAvailable(site.Port) {
		return fmt.Errorf("port %d is already in use", site.Port)
	}

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
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 40px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            text-align: center;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            padding: 40px;
            border-radius: 15px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
        }
        h1 {
            font-size: 2.5em;
            margin-bottom: 20px;
        }
        p {
            font-size: 1.2em;
            margin-bottom: 10px;
        }
        .status {
            background: rgba(0, 255, 0, 0.2);
            padding: 10px 20px;
            border-radius: 20px;
            margin-top: 20px;
            display: inline-block;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 ` + site.Name + `</h1>
        <p>Your site is running successfully!</p>
        <p>Server Port: <strong>` + fmt.Sprintf("%d", site.Port) + `</strong></p>
        <p>Local URL: <strong>http://localhost:` + fmt.Sprintf("%d", site.Port) + `</strong></p>
        <div class="status">
            ✅ Site is online and ready
        </div>
    </div>
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

func (c *Config) IsPortAvailable(port int) bool {
	// Check if port is already used by other sites
	for _, site := range c.Sites {
		if site.Port == port {
			return false
		}
	}

	// Check if port is available on system
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func (c *Config) GetAvailablePort() (int, error) {
	for port := c.AppSettings.AutoPortMin; port <= c.AppSettings.AutoPortMax; port++ {
		if c.IsPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", c.AppSettings.AutoPortMin, c.AppSettings.AutoPortMax)
}
