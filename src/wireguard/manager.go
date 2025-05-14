package wireguard

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gumbees/cfwg-zt/src/cloudflare"
	"github.com/gumbees/cfwg-zt/src/config"
)

// Manager handles WireGuard configuration generation and management
type Manager struct {
	config *config.Config
}

// NewManager creates a new WireGuard manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

// ValidateConfig checks if the WireGuard configuration is properly set up for Cloudflare Zero Trust
// This is especially useful for validating that the dummy configuration was properly imported
func (m *Manager) ValidateConfig() (bool, error) {
	configPath := m.config.WireGuard.ConfigPath
	
	// Check if the configuration file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, fmt.Errorf("WireGuard configuration file not found at %s", configPath)
	}
	
	// Read the configuration file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return false, fmt.Errorf("failed to read WireGuard configuration: %w", err)
	}
	
	configContent := string(configData)
	
	// Check for required sections
	if !strings.Contains(configContent, "[Interface]") {
		return false, fmt.Errorf("WireGuard configuration is missing [Interface] section")
	}
	
	if !strings.Contains(configContent, "[Peer]") {
		return false, fmt.Errorf("WireGuard configuration is missing [Peer] section")
	}
	
	// Check if it contains the dummy keys that need to be replaced
	if strings.Contains(configContent, "mLmL+DB1n8MfA+7Dc+vnEdZD+VffR3Li3QcJhdTLuEU=") ||
	   strings.Contains(configContent, "YOw/RK8gT3PR4ImRfpnfvJ8UTY3GfJlO6PcPbl40Tkw=") {
		log.Println("WireGuard configuration contains dummy keys that need to be replaced")
		log.Println("This is normal if you just imported the dummy configuration. Keys will be updated automatically.")
		// Return true because even with dummy keys, the file structure is valid
		return true, nil
	}
	
	return true, nil
}

// UpdateConfig updates the WireGuard configuration file with the provided Cloudflare configuration
// Only updates authentication-related fields while trying to preserve existing UDM Pro UI settings
func (m *Manager) UpdateConfig(cfg *cloudflare.WireGuardConfig) error {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(m.config.UDMPro.ConfigBackupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Check if an existing configuration is present
	configPath := m.config.WireGuard.ConfigPath
	existingConfig := ""
	hasExistingConfig := false

	if _, err := os.Stat(configPath); err == nil {
		// Create a backup of the existing configuration
		backupPath := filepath.Join(
			m.config.UDMPro.ConfigBackupPath,
			fmt.Sprintf("%s.%s.bak", 
				filepath.Base(configPath),
				time.Now().Format("20060102-150405"),
			),
		)
		
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read existing config for backup: %w", err)
		}
		
		existingConfig = string(configData)
		hasExistingConfig = true
		
		if err := os.WriteFile(backupPath, configData, 0600); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		
		log.Printf("Created backup of WireGuard configuration at %s", backupPath)
	}

	// Generate the new configuration content
	configContent := buildWireGuardConfig(cfg)
	
	// If we have an existing config, try to preserve some settings from it
	if hasExistingConfig && len(existingConfig) > 0 {
		configContent = mergeWithExistingConfig(existingConfig, cfg)
	}
	
	// Write the new configuration
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write WireGuard configuration: %w", err)
	}

	log.Printf("Updated WireGuard configuration at %s", configPath)
	return nil
}

// buildWireGuardConfig generates a WireGuard configuration file based on Cloudflare data
// It preserves the existing configuration structure and only updates authentication-related fields
func buildWireGuardConfig(cfg *cloudflare.WireGuardConfig) string {
	// Validate the configuration
	if cfg.PrivateKey == "" || cfg.PublicKey == "" || cfg.PeerPublicKey == "" || cfg.Endpoint == "" {
		log.Printf("Error: Invalid WireGuard configuration, missing required fields")
		log.Printf("PrivateKey present: %v", cfg.PrivateKey != "")
		log.Printf("PublicKey present: %v", cfg.PublicKey != "")
		log.Printf("PeerPublicKey present: %v", cfg.PeerPublicKey != "")
		log.Printf("Endpoint present: %v", cfg.Endpoint != "")
		return ""
	}

	wgConfigTemplate := `[Interface]
PrivateKey = {{ .PrivateKey }}
# Note: Address and DNS settings are now managed via the UDM Pro UI
# The following lines are maintained for compatibility with UI-created config
Address = 100.64.0.1/32
{{- if .DNS }}
DNS = {{ range $index, $dns := .DNS }}{{if $index}}, {{end}}{{ $dns }}{{end}}
{{- end }}
MTU = 1280

[Peer]
PublicKey = {{ .PeerPublicKey }}
{{- if .PeerPresharedKey }}
PresharedKey = {{ .PeerPresharedKey }}
{{- end }}
# AllowedIPs is now managed via the UDM Pro UI's policy-based routing
AllowedIPs = {{ range $index, $ip := .AllowedIPs }}{{if $index}}, {{end}}{{ $ip }}{{end}}
Endpoint = {{ .Endpoint }}:{{ .EndpointPort }}
PersistentKeepalive = 25
`

	// Create a template and parse it
	tmpl, err := template.New("wireguard").Parse(wgConfigTemplate)
	if err != nil {
		log.Printf("Error creating WireGuard config template: %v", err)
		// Return a basic configuration as fallback
		return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 100.64.0.1/32
MTU = 1280
Table = off

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = %s:%d
PersistentKeepalive = 25
`, cfg.PrivateKey, cfg.PeerPublicKey, cfg.Endpoint, cfg.EndpointPort)
	}

	var result strings.Builder
	err = tmpl.Execute(&result, cfg)
	if err != nil {
		log.Printf("Error executing WireGuard config template: %v", err)
		// Return a basic configuration as fallback
		return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 100.64.0.1/32
MTU = 1280
Table = off

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = %s:%d
PersistentKeepalive = 25
`, cfg.PrivateKey, cfg.PeerPublicKey, cfg.Endpoint, cfg.EndpointPort)
	}
	return result.String()
}

// mergeWithExistingConfig tries to preserve settings from the existing WireGuard config
// while updating only the authentication-related fields from Cloudflare
func mergeWithExistingConfig(existingConfig string, cfg *cloudflare.WireGuardConfig) string {
	lines := strings.Split(existingConfig, "\n")
	var result strings.Builder
	inInterface := false
	inPeer := false
	
	// Process each line of the existing config
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		if trimmedLine == "[Interface]" {
			inInterface = true
			inPeer = false
			result.WriteString(line + "\n")
			continue
		} else if trimmedLine == "[Peer]" {
			inInterface = false
			inPeer = true
			result.WriteString(line + "\n")
			continue
		}
		
		// Skip empty lines
		if trimmedLine == "" {
			result.WriteString(line + "\n")
			continue
		}
		
		// Handle Interface section
		if inInterface {
			// Update PrivateKey, keep other settings
			if strings.HasPrefix(trimmedLine, "PrivateKey") {
				result.WriteString("PrivateKey = " + cfg.PrivateKey + "\n")
			} else {
				// Keep original line
				result.WriteString(line + "\n")
			}
		}
		
		// Handle Peer section
		if inPeer {
			if strings.HasPrefix(trimmedLine, "PublicKey") {
				result.WriteString("PublicKey = " + cfg.PeerPublicKey + "\n")
			} else if strings.HasPrefix(trimmedLine, "PresharedKey") && cfg.PeerPresharedKey != "" {
				result.WriteString("PresharedKey = " + cfg.PeerPresharedKey + "\n")
			} else if strings.HasPrefix(trimmedLine, "Endpoint") {
				result.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", cfg.Endpoint, cfg.EndpointPort))
			} else {
				// Keep original line (including AllowedIPs which is now managed via UI)
				result.WriteString(line + "\n")
			}
		}
	}
	
	// If we didn't find certain sections, add them
	if !strings.Contains(existingConfig, "PersistentKeepalive") {
		result.WriteString("PersistentKeepalive = 25\n")
	}
	
	return result.String()
}
