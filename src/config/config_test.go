package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create default config
	if err := CreateDefaultConfigFile(configPath); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set environment variable to point to our config
	os.Setenv("CFWG_CONFIG_FILE", configPath)
	defer os.Unsetenv("CFWG_CONFIG_FILE")

	// Test loading the config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Basic validation
	if cfg.RefreshIntervalMinutes != 60 {
		t.Errorf("Expected RefreshIntervalMinutes to be 60, got %d", cfg.RefreshIntervalMinutes)
	}

	if cfg.WireGuard.InterfaceName != "wg0" {
		t.Errorf("Expected WireGuard.InterfaceName to be wg0, got %s", cfg.WireGuard.InterfaceName)
	}
}

func TestCreateDefaultConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "new_config.yaml")

	// Create default config
	if err := CreateDefaultConfigFile(configPath); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Read the file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Very basic check that the file contains expected content
	if len(content) == 0 {
		t.Errorf("Config file is empty")
	}
}
