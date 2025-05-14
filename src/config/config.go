package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Cloudflare Zero Trust configuration
	CloudflareZeroTrust struct {
		ClientID     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		TeamName     string `mapstructure:"team_name"`
		AccountID    string `mapstructure:"account_id"`
	} `mapstructure:"cloudflare_zero_trust"`

	// WireGuard configuration
	WireGuard struct {
		InterfaceName string `mapstructure:"interface_name"`
		ConfigPath    string `mapstructure:"config_path"`
	} `mapstructure:"wireguard"`

	// UDM-Pro configuration
	UDMPro struct {
		WireGuardServiceName string `mapstructure:"wireguard_service_name"`
		ConfigBackupPath     string `mapstructure:"config_backup_path"`
	} `mapstructure:"udm_pro"`

	// General configuration
	RefreshIntervalMinutes int  `mapstructure:"refresh_interval_minutes"`
	Debug                  bool `mapstructure:"debug"`
}

// LoadConfig loads the application configuration from file or environment variables
func LoadConfig() (*Config, error) {
	// Set default configuration
	viper.SetDefault("refresh_interval_minutes", 60) // Default refresh every 60 minutes
	viper.SetDefault("debug", false)
	viper.SetDefault("wireguard.interface_name", "wg0")
	viper.SetDefault("wireguard.config_path", "/etc/wireguard/wg0.conf")
	viper.SetDefault("udm_pro.wireguard_service_name", "wg-quick@wg0")
	viper.SetDefault("udm_pro.config_backup_path", "/etc/wireguard/backup")

	// Set the config file name and paths to look for it
	viper.SetConfigName("config") // Name of config file (without extension)
	viper.SetConfigType("yaml")   // Config file type

	// Look for config in the current directory
	viper.AddConfigPath(".")
	
	// Also look for config in /etc/cfwg-zt/ directory
	viper.AddConfigPath("/etc/cfwg-zt/")
	
	// Also look in home directory
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(home, ".cfwg-zt"))
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found; ignore error if desired
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		} else {
			fmt.Println("Config file not found. Using default values and environment variables.")
		}
	}

	// Override config from environment variables (optional)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CFWG") // Environment variables will be prefixed with CFWG_

	// Read the configuration into our struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// CreateDefaultConfigFile creates a default configuration file at the specified path
func CreateDefaultConfigFile(path string) error {
	defaultConfig := `# Cloudflare Zero Trust WireGuard Manager Configuration
# This application maintains Cloudflare Zero Trust authentication for a UDM Pro UI-created WireGuard configuration

# Cloudflare Zero Trust settings
cloudflare_zero_trust:
  client_id: "your_client_id_here"
  client_secret: "your_client_secret_here"
  team_name: "your_team_name_here"
  account_id: "your_account_id_here"

# WireGuard settings - these should match your UI-created configuration
# You can find the interface name in the UDM Pro UI under Settings > VPN > WireGuard
wireguard:
  interface_name: "wg0"
  config_path: "/etc/wireguard/wg0.conf"

# UDM-Pro specific settings
udm_pro:
  wireguard_service_name: "wg-quick@wg0"  # Must match your interface name
  config_backup_path: "/etc/wireguard/backup"

# General settings
refresh_interval_minutes: 60  # How often to refresh authentication
debug: false
`

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}
