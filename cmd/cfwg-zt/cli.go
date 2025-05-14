package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gumbees/cfwg-zt/src/cloudflare"
	"github.com/gumbees/cfwg-zt/src/config"
	"github.com/gumbees/cfwg-zt/src/udm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cfwg-zt",
	Short: "Cloudflare Zero Trust WireGuard Manager for UDM-Pro",
	Long:  `A tool to maintain a WireGuard configuration authenticated to Cloudflare Zero Trust for Business on a UDM-Pro device.`,
}

var (
	configFile string
	debugMode  bool
)

func init() {
	// Root command flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to config file (default is /etc/cfwg-zt/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug mode")

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(configWizardCmd)
	rootCmd.AddCommand(versionCmd)
}

// startCmd represents the start command for running the service
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run: func(cmd *cobra.Command, args []string) {
		// This simply calls the main function which starts the service
		runService()
	},
}

// statusCmd checks the status of the WireGuard connection
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the WireGuard connection",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfigWithFlags()
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}

		// Initialize components
		cfClient, err := cloudflare.NewClient(cfg)
		if err != nil {
			log.Fatalf("Error initializing Cloudflare client: %v", err)
		}

		udmClient := udm.NewClient(cfg)
		
		// First check if the config file exists
		configPath := cfg.WireGuard.ConfigPath
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("WireGuard configuration file not found at %s\n", configPath)
			fmt.Println("If you created a configuration through the UDM Pro UI, make sure this application")
			fmt.Printf("is configured with the correct path to the UI-created WireGuard configuration file.\n")
			os.Exit(1)
		}

		// Check if WireGuard is running
		isRunning, err := udmClient.IsWireGuardRunning()
		if err != nil {
			log.Fatalf("Error checking WireGuard status: %v", err)
		}

		if !isRunning {
			fmt.Println("WireGuard is not running. Please check your UDM Pro UI settings.")
			fmt.Println("You may need to enable the WireGuard interface in the UDM Pro UI.")
			os.Exit(1)
		}

		// Authenticate to check device status
		deviceToken, err := cfClient.AuthenticateDevice()
		if err != nil {
			log.Fatalf("Error authenticating with Cloudflare: %v", err)
		}

		// Check device status
		active, err := cfClient.GetDeviceStatus(deviceToken)
		if err != nil {
			fmt.Println("WireGuard is running but Cloudflare Zero Trust status is unknown")
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if active {
			fmt.Println("WireGuard is running and connected to Cloudflare Zero Trust")
			fmt.Println("The UDM Pro UI-created WireGuard configuration is being maintained successfully.")
			fmt.Println("You can use policy-based routing in the UDM Pro UI to route traffic through this tunnel.")
		} else {
			fmt.Println("WireGuard is running but not active in Cloudflare Zero Trust")
			fmt.Println("The application will attempt to reconnect automatically.")
			os.Exit(1)
		}
	},
}

// setupCmd creates a new configuration file
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up a new configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Determine path for the config file
		configPath := configFile
		if configPath == "" {
			configPath = "/etc/cfwg-zt/config.yaml"
		}

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Configuration file already exists at %s\n", configPath)
			fmt.Println("Do you want to overwrite it? (y/n)")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Setup aborted")
				return
			}
		}

		// Create the config file
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}

		if err := config.CreateDefaultConfigFile(configPath); err != nil {
			log.Fatalf("Failed to create config file: %v", err)
		}

		fmt.Printf("Configuration file created at %s\n", configPath)
		fmt.Println("Please edit this file to add your Cloudflare Zero Trust credentials")
	},
}

// configWizardCmd creates a new configuration file interactively
var configWizardCmd = &cobra.Command{
	Use:   "config-wizard",
	Short: "Interactive configuration wizard",
	Long:  `Guides you through the process of creating a configuration file by asking questions interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine path for the config file
		configPath := configFile
		if configPath == "" {
			configPath = "/etc/cfwg-zt/config.yaml"
		}

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Configuration file already exists at %s\n", configPath)
			fmt.Println("Do you want to overwrite it? (y/n)")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Config wizard aborted")
				return
			}
		}

		// Create the config directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}

		// Start the interactive configuration
		cfg, err := config.RunWizard()
		if err != nil {
			log.Fatalf("Failed to complete configuration wizard: %v", err)
		}

		// Save the configuration
		if err := config.SaveConfig(cfg, configPath); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
		fmt.Printf("Configuration file created at %s\n", configPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Make sure you have a WireGuard configuration in your UDM Pro UI")
		fmt.Println("   - If not, import the dummy configuration at /etc/cfwg-zt/dummy-wireguard.conf")
		fmt.Println("   - Go to UDM Pro UI: Settings > VPN > WireGuard > Create New > Import")
		fmt.Println("   - Select the file '/etc/cfwg-zt/dummy-wireguard.conf' and click 'Add'")
		fmt.Println("   - The dummy configuration contains temporary keys and will be properly configured by the application")
		fmt.Println("2. Start the service with: cfwg-zt start")
	},
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cloudflare Zero Trust WireGuard Manager v1.0.0")
	},
}

// loadConfigWithFlags loads the configuration with command line flags taken into account
func loadConfigWithFlags() (*config.Config, error) {
	// If a config file is provided, set it in the environment
	if configFile != "" {
		os.Setenv("CFWG_CONFIG_FILE", configFile)
	}

	// Load the config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Override debug mode if specified
	if debugMode {
		cfg.Debug = true
	}

	return cfg, nil
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
