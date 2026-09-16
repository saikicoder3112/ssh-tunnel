package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Connections []Connection `yaml:"connections"`
}

type Connection struct {
	Title        string `yaml:"title"`
	User         string `yaml:"user,omitempty"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port,omitempty"`
	Compression  bool   `yaml:"compression,omitempty"`
	ForwardX     bool   `yaml:"forward_x,omitempty"`
	ForwardAgent bool   `yaml:"forward_agent,omitempty"`
	GatewayPorts bool   `yaml:"gateway_ports,omitempty"`
	Custom       string `yaml:"custom,omitempty"`
	Prepend      string `yaml:"prepend,omitempty"`
	Append       string `yaml:"append,omitempty"`
	Sort         int    `yaml:"sort,omitempty"`
}

func (c Connection) Target() string {
	if c.User != "" {
		return c.User + "@" + c.Host
	}
	return c.Host
}

func (c Connection) SSHPort() int {
	if c.Port > 0 {
		return c.Port
	}
	return 22
}

func Path() string {
	if p := os.Getenv("MYTUNNEL_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil {
		sshPath := filepath.Join(home, ".ssh", "mytunnel.yaml")
		if _, err := os.Stat(sshPath); err == nil {
			return sshPath
		}
		if _, err := os.Stat("config.yaml"); err == nil {
			return "config.yaml"
		}
		return sshPath
	}
	return "config.yaml"
}

func Load() (*Config, string, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, path, nil
		}
		return nil, path, fmt.Errorf("read config: %w", err)
	}

	var raw struct {
		Connections []Connection `yaml:"connections"`
		Servers     []oldServer  `yaml:"servers"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, path, fmt.Errorf("parse config: %w", err)
	}

	cfg := &Config{Connections: raw.Connections}
	if len(cfg.Connections) == 0 && len(raw.Servers) > 0 {
		cfg.Connections = migrateServers(raw.Servers)
	}
	cfg.Sort()
	return cfg, path, nil
}

func (c *Config) Save(path string) error {
	c.Sort()
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	_ = enc.Close()
	if err := os.WriteFile(path, []byte(buf.String()), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) Sort() {
	if c == nil {
		return
	}
	sort.SliceStable(c.Connections, func(i, j int) bool {
		if c.Connections[i].Sort != c.Connections[j].Sort {
			return c.Connections[i].Sort < c.Connections[j].Sort
		}
		return strings.ToLower(c.Connections[i].Title) < strings.ToLower(c.Connections[j].Title)
	})
}

type oldServer struct {
	Name    string      `yaml:"name"`
	Host    string      `yaml:"host"`
	User    string      `yaml:"user"`
	Port    int         `yaml:"port"`
	Key     string      `yaml:"key"`
	Tunnels []oldTunnel `yaml:"tunnels"`
}

type oldTunnel struct {
	Name   string `yaml:"name"`
	Local  int    `yaml:"local"`
	Remote string `yaml:"remote"`
}

func migrateServers(servers []oldServer) []Connection {
	var out []Connection
	for _, s := range servers {
		var flags []string
		if s.Key != "" {
			flags = append(flags, "-i "+s.Key)
		}
		for _, t := range s.Tunnels {
			flags = append(flags, fmt.Sprintf("-L %d:%s", t.Local, t.Remote))
		}
		out = append(out, Connection{
			Title:  s.Name,
			User:   s.User,
			Host:   s.Host,
			Port:   s.Port,
			Custom: strings.Join(flags, " "),
		})
	}
	return out
}
