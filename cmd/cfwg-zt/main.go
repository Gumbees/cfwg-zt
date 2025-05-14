package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gumbees/cfwg-zt/src/config"
	"github.com/gumbees/cfwg-zt/src/cloudflare"
	"github.com/gumbees/cfwg-zt/src/wireguard"
	"github.com/gumbees/cfwg-zt/src/udm"
	"github.com/spf13/viper"
)

// setupLogging configures the application logging
func setupLogging(debug bool) (*os.File, error) {
	// Create log directory if it doesn't exist
	logDir := "/var/log/cfwg-zt"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Open log file
	logFilePath := filepath.Join(logDir, "cfwg-zt.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Set up multi-writer to log to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// Set log flags to include timestamp and file information
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if debug {
		log.Println("Debug mode enabled")
	}

	return logFile, nil
}

func main() {
	// Parse CLI commands
	Execute()
}

// runService is the main function for running the service
func runService() {
	log.Println("Starting Cloudflare Zero Trust WireGuard Manager for UDM-Pro")
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Setup logging
	logFile, err := setupLogging(cfg.Debug)
	if err != nil {
		log.Fatalf("Error setting up logging: %v", err)
	}
	defer logFile.Close()

	// Log the startup details
	log.Printf("Version: 1.0.0")
	log.Printf("Configuration loaded from: %s", viper.ConfigFileUsed())
	log.Printf("Refresh interval: %d minutes", cfg.RefreshIntervalMinutes)
	
	// Initialize components
	log.Println("Initializing components...")
	cfClient, err := cloudflare.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error initializing Cloudflare client: %v", err)
	}

	wgManager := wireguard.NewManager(cfg)
	udmClient := udm.NewClient(cfg)

	// Validate that we're running on a UDM-Pro (if possible)
	if _, err := os.Stat("/usr/bin/ubnt-systool"); os.IsNotExist(err) {
		log.Println("Warning: This doesn't appear to be a UDM-Pro device. Some functionality may not work as expected.")
	}

	// Verify that WireGuard is available
	if err := udmClient.VerifyWireGuardAvailable(); err != nil {
		log.Fatalf("WireGuard is not properly available on this system: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s, initiating shutdown...", sig)
		// Perform any necessary cleanup here
		done <- true
	}()

	// Validate the WireGuard configuration
	log.Println("Validating WireGuard configuration...")
	valid, err := wgManager.ValidateConfig()
	if err != nil {
		log.Printf("Warning: WireGuard configuration validation error: %v", err)
		log.Println("This might happen if you've just imported the dummy configuration.")
		log.Println("The application will attempt to fix this by updating with proper credentials.")
	} else if valid {
		log.Println("WireGuard configuration validation successful.")
	}
	
	// Start the main service loop
	log.Println("Starting main service loop...")
	go func() {
		consecutiveFailures := 0
		maxConsecutiveFailures := 5

		for {
			// Break the loop if we've had too many consecutive failures
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Printf("Too many consecutive failures (%d), entering exponential backoff", consecutiveFailures)
				backoffTime := time.Duration(math.Min(float64(consecutiveFailures-maxConsecutiveFailures+1)*2, 30)) * time.Minute
				log.Printf("Backing off for %v", backoffTime)
				time.Sleep(backoffTime)
				// Reset counter after backoff, but not completely
				consecutiveFailures = maxConsecutiveFailures - 2
			}

			// Authenticate with Cloudflare Zero Trust
			log.Println("Authenticating with Cloudflare Zero Trust...")
			deviceToken, err := cfClient.AuthenticateDevice()
			if err != nil {
				consecutiveFailures++
				log.Printf("Error authenticating device: %v, retrying in 1 minute (failure %d/%d)", 
					err, consecutiveFailures, maxConsecutiveFailures)
				time.Sleep(time.Minute)
				continue
			}

			// Get WireGuard configuration from Cloudflare
			log.Println("Retrieving WireGuard configuration...")
			wgConfig, err := cfClient.GetWireGuardConfig(deviceToken)
			if err != nil {
				consecutiveFailures++
				log.Printf("Error getting WireGuard config: %v, retrying in 1 minute (failure %d/%d)", 
					err, consecutiveFailures, maxConsecutiveFailures)
				time.Sleep(time.Minute)
				continue
			}			// Check if WireGuard is running before updating config
			isRunning, err := udmClient.IsWireGuardRunning()
			if err != nil {
				log.Printf("Error checking WireGuard status: %v", err)
			}
			
			if !isRunning {
				log.Printf("WireGuard is not running. The UDM-Pro UI-created configuration may have been disabled. " +
				          "Please check your UDM-Pro settings. Will retry in 5 minutes.")
				time.Sleep(5 * time.Minute)
				continue
			}

			// Update WireGuard configuration - preserving UI-created settings
			log.Println("Updating WireGuard configuration file with fresh authentication credentials...")
			log.Println("Note: UI-created settings like interface address and policy-based routing will be preserved")
			err = wgManager.UpdateConfig(wgConfig)
			if err != nil {
				consecutiveFailures++
				log.Printf("Error updating WireGuard config: %v, retrying in 1 minute (failure %d/%d)", 
					err, consecutiveFailures, maxConsecutiveFailures)
				time.Sleep(time.Minute)
				continue
			}

			// Apply the configuration on the UDM-Pro (only restarts the service)
			log.Println("Applying WireGuard configuration to UDM-Pro...")
			err = udmClient.ApplyWireGuardConfig(wgConfig)
			if err != nil {
				consecutiveFailures++
				log.Printf("Error applying WireGuard config: %v, retrying in 1 minute (failure %d/%d)", 
					err, consecutiveFailures, maxConsecutiveFailures)
				time.Sleep(time.Minute)
				continue
			}

			// Reset consecutive failures counter after a successful run
			consecutiveFailures = 0
			log.Println("WireGuard configuration successfully updated and applied")			// Schedule a refresh of the device registration (to keep it active)
			refreshTime := time.Duration(cfg.RefreshIntervalMinutes) * time.Minute / 2
			time.AfterFunc(refreshTime, func() {
				if err := cfClient.RefreshDeviceRegistration(deviceToken); err != nil {
					log.Printf("Warning: Failed to refresh device registration: %v", err)
				} else {
					log.Println("Device registration refreshed successfully")
				}
			})
			
			// Sleep for the refresh interval from config
			log.Printf("Next configuration check in %d minutes", cfg.RefreshIntervalMinutes)
			time.Sleep(time.Duration(cfg.RefreshIntervalMinutes) * time.Minute)
		}
	}()
	<-done
	log.Println("Shutting down...")
}
