package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type InterfaceConfig struct {
	PrivateKey  string `yaml:"private_key"`
	Address     string `yaml:"address"`
	DNS         string `yaml:"dns"`
	RouteSubnet string `yaml:"route_subnet"`
}

type ServerPeer struct {
	PublicKey           string `yaml:"public_key"`
	Endpoint            string `yaml:"endpoint"`
	AllowedIPs          string `yaml:"allowed_ips"`
	PersistentKeepalive int    `yaml:"persistent_keepalive"`
}

type Config struct {
	Interface    InterfaceConfig `yaml:"interface"`
	ServerConfig ServerPeer      `yaml:"peer"`
	Etcd         struct {
		Endpoint string `yaml:"endpoint"`
	} `yaml:"etcd"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}
