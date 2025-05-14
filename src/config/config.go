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

// RunWizard runs an interactive configuration wizard and returns the resulting config
func RunWizard() (*Config, error) {
	cfg := &Config{}

	fmt.Println("==== Cloudflare Zero Trust WireGuard Manager Configuration Wizard ====")
	fmt.Println()
	fmt.Println("This wizard will help you set up your configuration.")
	fmt.Println()

	// Cloudflare Zero Trust settings
	fmt.Println("==== Cloudflare Zero Trust Settings ====")
	fmt.Println("You'll need to get these values from your Cloudflare Zero Trust dashboard.")
	fmt.Println("Visit: https://dash.cloudflare.com/ and navigate to Zero Trust > Settings > Authentication")
	fmt.Println()

	fmt.Print("Enter your Cloudflare Account ID: ")
	fmt.Scanln(&cfg.CloudflareZeroTrust.AccountID)

	fmt.Print("Enter your Cloudflare Team Name: ")
	fmt.Scanln(&cfg.CloudflareZeroTrust.TeamName)

	fmt.Print("Enter your Cloudflare Client ID: ")
	fmt.Scanln(&cfg.CloudflareZeroTrust.ClientID)

	fmt.Print("Enter your Cloudflare Client Secret: ")
	fmt.Scanln(&cfg.CloudflareZeroTrust.ClientSecret)
	// WireGuard settings
	fmt.Println()
	fmt.Println("==== WireGuard Settings ====")
	fmt.Println("These settings should match your UDM Pro WireGuard configuration.")
	fmt.Println("Have you already created a WireGuard configuration in the UDM Pro UI?")
	fmt.Println("If not, you can import the dummy configuration file at /etc/cfwg-zt/dummy-wireguard.conf")
	fmt.Println("Go to UDM Pro UI: Settings > VPN > WireGuard > Create New > Import")
	fmt.Println("The dummy configuration includes temporary keys that will be replaced automatically")
	fmt.Println("and is pre-configured with the correct settings for Cloudflare Zero Trust.")
	fmt.Println()

	// Set default values
	cfg.WireGuard.InterfaceName = "wg0"
	cfg.WireGuard.ConfigPath = "/etc/wireguard/wg0.conf"
	
	fmt.Printf("Enter WireGuard interface name (default: %s): ", cfg.WireGuard.InterfaceName)
	var input string
	fmt.Scanln(&input)
	if input != "" {
		cfg.WireGuard.InterfaceName = input
	}

	fmt.Printf("Enter WireGuard config path (default: %s): ", cfg.WireGuard.ConfigPath)
	input = ""
	fmt.Scanln(&input)
	if input != "" {
		cfg.WireGuard.ConfigPath = input
	}

	// UDM Pro specific settings
	fmt.Println()
	fmt.Println("==== UDM Pro Settings ====")
	fmt.Println()

	// Set default values
	cfg.UDMPro.WireGuardServiceName = "wg-quick@" + cfg.WireGuard.InterfaceName
	cfg.UDMPro.ConfigBackupPath = "/etc/wireguard/backup"

	fmt.Printf("Enter WireGuard service name (default: %s): ", cfg.UDMPro.WireGuardServiceName)
	input = ""
	fmt.Scanln(&input)
	if input != "" {
		cfg.UDMPro.WireGuardServiceName = input
	}

	fmt.Printf("Enter config backup path (default: %s): ", cfg.UDMPro.ConfigBackupPath)
	input = ""
	fmt.Scanln(&input)
	if input != "" {
		cfg.UDMPro.ConfigBackupPath = input
	}

	// General settings
	fmt.Println()
	fmt.Println("==== General Settings ====")
	fmt.Println()

	cfg.RefreshIntervalMinutes = 60
	fmt.Printf("Enter configuration refresh interval in minutes (default: %d): ", cfg.RefreshIntervalMinutes)
	var refreshInterval int
	fmt.Scanln(&refreshInterval)
	if refreshInterval > 0 {
		cfg.RefreshIntervalMinutes = refreshInterval
	}

	cfg.Debug = false
	fmt.Print("Enable debug mode? (y/n, default: n): ")
	input = ""
	fmt.Scanln(&input)
	if input == "y" || input == "Y" {
		cfg.Debug = true
	}

	fmt.Println()
	fmt.Println("Configuration wizard complete!")

	return cfg, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(cfg *Config, path string) error {
	// Create a new viper instance
	v := viper.New()
	v.SetConfigFile(path)
	
	// Set the values from the config struct
	v.Set("cloudflare_zero_trust.client_id", cfg.CloudflareZeroTrust.ClientID)
	v.Set("cloudflare_zero_trust.client_secret", cfg.CloudflareZeroTrust.ClientSecret)
	v.Set("cloudflare_zero_trust.team_name", cfg.CloudflareZeroTrust.TeamName)
	v.Set("cloudflare_zero_trust.account_id", cfg.CloudflareZeroTrust.AccountID)
	
	v.Set("wireguard.interface_name", cfg.WireGuard.InterfaceName)
	v.Set("wireguard.config_path", cfg.WireGuard.ConfigPath)
	
	v.Set("udm_pro.wireguard_service_name", cfg.UDMPro.WireGuardServiceName)
	v.Set("udm_pro.config_backup_path", cfg.UDMPro.ConfigBackupPath)
	
	v.Set("refresh_interval_minutes", cfg.RefreshIntervalMinutes)
	v.Set("debug", cfg.Debug)
	
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Save the config file
	if err := v.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, write it
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}
	
	return nil
}
