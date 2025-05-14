package udm

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/gumbees/cfwg-zt/src/cloudflare"
	"github.com/gumbees/cfwg-zt/src/config"
)

// Client handles interactions with the UDM-Pro system
type Client struct {
	config *config.Config
}

// NewClient creates a new UDM-Pro client
func NewClient(cfg *config.Config) *Client {
	return &Client{config: cfg}
}

// VerifyWireGuardAvailable checks if WireGuard is properly installed and available
func (c *Client) VerifyWireGuardAvailable() error {
	// Check if wg command exists
	wgCmd := exec.Command("which", "wg")
	if err := wgCmd.Run(); err != nil {
		return fmt.Errorf("WireGuard 'wg' command not found: %w", err)
	}

	// Check if wg-quick is available
	wgQuickCmd := exec.Command("which", "wg-quick")
	if err := wgQuickCmd.Run(); err != nil {
		return fmt.Errorf("WireGuard 'wg-quick' command not found: %w", err)
	}

	// Check if the configured interface name is reasonable
	if c.config.WireGuard.InterfaceName == "" {
		return fmt.Errorf("WireGuard interface name not configured")
	}

	// Verify systemd service name
	if c.config.UDMPro.WireGuardServiceName == "" {
		return fmt.Errorf("WireGuard service name not configured")
	}

	return nil
}

// ApplyWireGuardConfig applies the WireGuard configuration to the UDM-Pro system
// It only restarts the WireGuard service and doesn't modify routing
func (c *Client) ApplyWireGuardConfig(cfg *cloudflare.WireGuardConfig) error {
	// First, check if WireGuard is already running
	isRunning, err := c.isWireGuardRunning()
	if err != nil {
		return fmt.Errorf("failed to check WireGuard service status: %w", err)
	}

	// If running, we need to restart the service
	if isRunning {
		if err := c.restartWireGuardService(); err != nil {
			return fmt.Errorf("failed to restart WireGuard service: %w", err)
		}
	} else {
		// If not running, start the service
		if err := c.startWireGuardService(); err != nil {
			return fmt.Errorf("failed to start WireGuard service: %w", err)
		}
	}

	// After restarting/starting WireGuard, verify it's running
	isRunning, err = c.isWireGuardRunning()
	if err != nil {
		return fmt.Errorf("failed to verify WireGuard service status after restart: %w", err)
	}

	if !isRunning {
		return fmt.Errorf("WireGuard service failed to start")
	}

	return nil
}

// isWireGuardRunning checks if the WireGuard service is running
func (c *Client) isWireGuardRunning() (bool, error) {
	cmd := exec.Command("systemctl", "is-active", c.config.UDMPro.WireGuardServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If error is because the service is not active, return false without error
		if strings.TrimSpace(string(output)) == "inactive" || 
		   strings.TrimSpace(string(output)) == "unknown" {
			return false, nil
		}
		return false, fmt.Errorf("error checking WireGuard service: %w", err)
	}

	return strings.TrimSpace(string(output)) == "active", nil
}

// IsWireGuardRunning is a public method for checking if WireGuard is running
func (c *Client) IsWireGuardRunning() (bool, error) {
	return c.isWireGuardRunning()
}

// startWireGuardService starts the WireGuard service
func (c *Client) startWireGuardService() error {
	log.Printf("Starting WireGuard service: %s", c.config.UDMPro.WireGuardServiceName)
	cmd := exec.Command("systemctl", "start", c.config.UDMPro.WireGuardServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start WireGuard service: %v, output: %s", err, output)
	}
	return nil
}

// stopWireGuardService stops the WireGuard service
func (c *Client) stopWireGuardService() error {
	log.Printf("Stopping WireGuard service: %s", c.config.UDMPro.WireGuardServiceName)
	cmd := exec.Command("systemctl", "stop", c.config.UDMPro.WireGuardServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop WireGuard service: %v, output: %s", err, output)
	}
	return nil
}

// restartWireGuardService restarts the WireGuard service
func (c *Client) restartWireGuardService() error {
	log.Printf("Restarting WireGuard service: %s", c.config.UDMPro.WireGuardServiceName)
	cmd := exec.Command("systemctl", "restart", c.config.UDMPro.WireGuardServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart WireGuard service: %v, output: %s", err, output)
	}
	return nil
}
