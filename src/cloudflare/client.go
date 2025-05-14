package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nathanielsmith/cfwg-zt/src/config"
)

// Client handles interactions with the Cloudflare Zero Trust API
type Client struct {
	config      *config.Config
	httpClient  *http.Client
	baseURL     string
	accessToken string
	tokenExpiry time.Time
}

// WireGuardConfig contains WireGuard configuration details
type WireGuardConfig struct {
	PrivateKey       string
	PublicKey        string
	Endpoint         string
	EndpointPort     int
	AllowedIPs       []string
	PeerPublicKey    string
	PeerPresharedKey string
	DNS              []string
}

// DeviceTokenResponse represents the response from Cloudflare device authentication
type DeviceTokenResponse struct {
	Success bool `json:"success"`
	Result  struct {
		DeviceID    string `json:"device_id"`
		Token       string `json:"token"`
		ExpiresAt   string `json:"expires_at"`
		WarpEnabled bool   `json:"warp_enabled"`
	} `json:"result"`
}

// WireGuardConfigResponse represents the WireGuard configuration from Cloudflare
type WireGuardConfigResponse struct {
	Success bool `json:"success"`
	Result  struct {
		ClientPublicKey   string   `json:"client_public_key"`
		ClientPrivateKey  string   `json:"client_private_key"`
		PeerPublicKey     string   `json:"peer_public_key"`
		Endpoint          string   `json:"endpoint"`
		EndpointPort      int      `json:"endpoint_port"`
		AllowedIPs        []string `json:"allowed_ips"`
		PeerPresharedKey  string   `json:"peer_preshared_key,omitempty"`
		DNSServers        []string `json:"dns_servers"`
		RotationExpiresAt string   `json:"rotation_expires_at"`
	} `json:"result"`
}

// NewClient creates a new Cloudflare API client
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.CloudflareZeroTrust.ClientID == "" || cfg.CloudflareZeroTrust.ClientSecret == "" {
		return nil, fmt.Errorf("missing Cloudflare Zero Trust credentials in configuration")
	}

	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s", cfg.CloudflareZeroTrust.AccountID),
	}, nil
}

// AuthenticateDevice authenticates with Cloudflare Zero Trust and returns a device token
func (c *Client) AuthenticateDevice() (string, error) {
	// Check if we have a valid token already
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	// Construct the request URL
	apiURL := fmt.Sprintf("%s/devices/warp/register", c.baseURL)
	
	// Prepare the request body
	requestBody := map[string]interface{}{
		"client_id":     c.config.CloudflareZeroTrust.ClientID,
		"client_secret": c.config.CloudflareZeroTrust.ClientSecret,
		"device_name":   "UDM-Pro-WARP",
		"device_type":   "router",
		"warp_enabled":  true,
	}
	
	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var deviceResp DeviceTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if !deviceResp.Success {
		return "", fmt.Errorf("device authentication failed")
	}

	// Parse the expiration time
	expiresAt, err := time.Parse(time.RFC3339, deviceResp.Result.ExpiresAt)
	if err != nil {
		// If we can't parse the expiry, set a default of 1 hour
		expiresAt = time.Now().Add(time.Hour)
	}

	// Store the token and its expiry
	c.accessToken = deviceResp.Result.Token
	c.tokenExpiry = expiresAt

	return c.accessToken, nil
}

// GetWireGuardConfig retrieves the WireGuard configuration from Cloudflare
func (c *Client) GetWireGuardConfig(deviceToken string) (*WireGuardConfig, error) {
	// Construct the request URL
	apiURL := fmt.Sprintf("%s/devices/warp/wireguard", c.baseURL)
	
	// Create the HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var wgResp WireGuardConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&wgResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !wgResp.Success {
		return nil, fmt.Errorf("failed to get WireGuard configuration")
	}

	// Transform the response to our internal WireGuardConfig structure
	config := &WireGuardConfig{
		PrivateKey:       wgResp.Result.ClientPrivateKey,
		PublicKey:        wgResp.Result.ClientPublicKey,
		Endpoint:         wgResp.Result.Endpoint,
		EndpointPort:     wgResp.Result.EndpointPort,
		AllowedIPs:       wgResp.Result.AllowedIPs,
		PeerPublicKey:    wgResp.Result.PeerPublicKey,
		PeerPresharedKey: wgResp.Result.PeerPresharedKey,
		DNS:              wgResp.Result.DNSServers,
	}

	return config, nil
}

// RefreshDeviceRegistration refreshes the device registration with Cloudflare
func (c *Client) RefreshDeviceRegistration(deviceToken string) error {
	// Construct the request URL
	apiURL := fmt.Sprintf("%s/devices/warp/refresh", c.baseURL)
	
	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	
	// Add query parameters
	q := url.Values{}
	q.Add("device_token", deviceToken)
	req.URL.RawQuery = q.Encode()
	
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check if refresh was successful
	if resp.StatusCode != http.StatusOK {
		// Try to read response body for more details about the error
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("device refresh failed with status: %s, response: %s", resp.Status, string(respBody))
	}

	return nil
}

// GetDeviceStatus retrieves the current status of the device in Cloudflare Zero Trust
func (c *Client) GetDeviceStatus(deviceToken string) (bool, error) {
	// Construct the request URL
	apiURL := fmt.Sprintf("%s/devices/warp/status", c.baseURL)
	
	// Create the HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode == http.StatusOK {
		var statusResp struct {
			Success bool `json:"success"`
			Result struct {
				Active      bool   `json:"active"`
				WarpEnabled bool   `json:"warp_enabled"`
				LastSeen    string `json:"last_seen"`
			} `json:"result"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
			return false, fmt.Errorf("error decoding response: %w", err)
		}
		
		return statusResp.Result.Active && statusResp.Result.WarpEnabled, nil
	}
	
	return false, fmt.Errorf("device status check failed with status: %s", resp.Status)
}
