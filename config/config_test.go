package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "kurohabaki-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("ValidConfig", func(t *testing.T) {
		// Create a valid test config file
		configPath := filepath.Join(tempDir, "valid-config.yaml")
		configData := `
interface:
  private_key: ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=
  address: 10.0.0.2/24
  dns: 1.1.1.1
  routes:
    - 0.0.0.0/0
peer:
  public_key: ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=
  endpoint: 192.168.1.1:51820
  allowed_ips: 0.0.0.0/0
  persistent_keepalive: 25
etcd:
  endpoint: 192.168.1.100:2379
`
		if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Load the config
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load valid config: %v", err)
		}

		// Verify config values
		if cfg.Interface.PrivateKey != "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=" {
			t.Errorf("Expected PrivateKey to be ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=, got %s", cfg.Interface.PrivateKey)
		}
		if cfg.Interface.Address != "10.0.0.2/24" {
			t.Errorf("Expected Address to be 10.0.0.2/24, got %s", cfg.Interface.Address)
		}
		if cfg.Interface.DNS != "1.1.1.1" {
			t.Errorf("Expected DNS to be 1.1.1.1, got %s", cfg.Interface.DNS)
		}
		if len(cfg.Interface.Routes) != 1 || cfg.Interface.Routes[0] != "0.0.0.0/0" {
			t.Errorf("Expected Routes to contain [0.0.0.0/0], got %v", cfg.Interface.Routes)
		}
		if cfg.ServerConfig.PublicKey != "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=" {
			t.Errorf("Expected PublicKey to be ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=, got %s", cfg.ServerConfig.PublicKey)
		}
		if cfg.ServerConfig.Endpoint != "192.168.1.1:51820" {
			t.Errorf("Expected Endpoint to be 192.168.1.1:51820, got %s", cfg.ServerConfig.Endpoint)
		}
		if cfg.ServerConfig.AllowedIPs != "0.0.0.0/0" {
			t.Errorf("Expected AllowedIPs to be 0.0.0.0/0, got %s", cfg.ServerConfig.AllowedIPs)
		}
		if cfg.ServerConfig.PersistentKeepalive != 25 {
			t.Errorf("Expected PersistentKeepalive to be 25, got %d", cfg.ServerConfig.PersistentKeepalive)
		}
		if cfg.Etcd.Endpoint != "192.168.1.100:2379" {
			t.Errorf("Expected Etcd endpoint to be 192.168.1.100:2379, got %s", cfg.Etcd.Endpoint)
		}
	})

	t.Run("FileNotExist", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.yaml")
		_, err := Load(nonExistentPath)
		if err == nil {
			t.Error("Expected error when loading non-existent file, got nil")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		// Create an invalid YAML file
		invalidPath := filepath.Join(tempDir, "invalid.yaml")
		invalidData := `
interface:
  private_key: ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghi=
  address: 10.0.0.2/24
  This is not valid YAML!
  dns: 1.1.1.1
`
		if err := os.WriteFile(invalidPath, []byte(invalidData), 0644); err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		_, err := Load(invalidPath)
		if err == nil {
			t.Error("Expected error when loading invalid YAML, got nil")
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		// Create an empty file
		emptyPath := filepath.Join(tempDir, "empty.yaml")
		if err := os.WriteFile(emptyPath, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		cfg, err := Load(emptyPath)
		if err != nil {
			t.Fatalf("Failed to load empty file: %v", err)
		}

		// Empty YAML is valid but should result in zero values
		if cfg.Interface.PrivateKey != "" {
			t.Errorf("Expected empty PrivateKey, got %s", cfg.Interface.PrivateKey)
		}
	})
}
