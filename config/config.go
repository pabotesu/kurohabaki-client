package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type InterfaceConfig struct {
	PrivateKey string `mapstructure:"private_key"`
	Address    string `mapstructure:"address"`
	DNS        string `mapstructure:"dns"`
}

type ServerPeer struct {
	PublicKey           string `mapstructure:"public_key"`
	Endpoint            string `mapstructure:"endpoint"`
	AllowedIPs          string `mapstructure:"allowed_ips"`
	PersistentKeepalive int    `mapstructure:"persistent_keepalive"`
}

type Config struct {
	Interface    InterfaceConfig `mapstructure:"interface"`
	ServerConfig ServerPeer      `mapstructure:"peer"`
	Etcd         struct {
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"etcd"`
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
