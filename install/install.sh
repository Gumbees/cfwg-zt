#!/bin/bash
# Installation script for Cloudflare Zero Trust WireGuard Manager for UDM-Pro
# Version: 1.0.0

# Exit on any error
set -e

echo "=================================================="
echo "Cloudflare Zero Trust WireGuard Manager for UDM-Pro"
echo "Installation Script"
echo "=================================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root"
  exit 1
fi

# Check if we're on a UDM-Pro
if [ ! -f /usr/bin/ubnt-systool ] || ! grep -q "UDM-Pro" /etc/version; then
  echo "Warning: This doesn't appear to be a UDM-Pro device."
  echo "This application is specifically designed for UDM-Pro."
  read -p "Continue anyway? (y/n): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Check if WireGuard is available
if ! command -v wg &> /dev/null; then
  echo "Error: WireGuard is not installed or not available in PATH."
  echo "Please ensure WireGuard is properly installed on your UDM-Pro."
  exit 1
fi

# Check if there are any UI-created WireGuard configurations
WIREGUARD_CONFIG_COUNT=$(find /etc/wireguard -name "*.conf" | wc -l)
if [ "$WIREGUARD_CONFIG_COUNT" -eq 0 ]; then
  echo "Warning: No WireGuard configurations found in /etc/wireguard."
  echo "Please create a WireGuard configuration through the UDM-Pro UI first."
  echo "This application maintains authentication for a UI-created configuration."
  read -p "Continue anyway? (y/n): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
else
  echo "Found $WIREGUARD_CONFIG_COUNT WireGuard configuration(s) in /etc/wireguard."
  echo "This application will maintain Cloudflare Zero Trust authentication for these configurations."
fi

# Create directories
echo "Creating application directories..."
mkdir -p /etc/cfwg-zt
mkdir -p /var/log/cfwg-zt

# Install binary
echo "Installing application binary..."
cp ./cfwg-zt /usr/local/bin/
chmod +x /usr/local/bin/cfwg-zt

# Copy the dummy WireGuard configuration for UDM Pro UI import
if [ -f "./install/dummy-wireguard.conf" ]; then
  echo "Installing dummy WireGuard configuration file..."
  mkdir -p /etc/cfwg-zt
  cp ./install/dummy-wireguard.conf /etc/cfwg-zt/
  echo "The dummy WireGuard configuration has been installed to /etc/cfwg-zt/dummy-wireguard.conf"
  echo "To use it:"
  echo "1. Go to UDM Pro UI: Settings > VPN > WireGuard > Create New > Import"
  echo "2. Select the file at /etc/cfwg-zt/dummy-wireguard.conf"
  echo "3. Click 'Add' to create the interface"
  echo "This configuration contains temporary keys that will be replaced by the application"
fi

# Create default config if it doesn't exist
if [ ! -f /etc/cfwg-zt/config.yaml ]; then
  echo "Creating default configuration file..."
  cat > /etc/cfwg-zt/config.yaml << EOF
# Cloudflare Zero Trust WireGuard Manager Configuration

# Cloudflare Zero Trust settings
cloudflare_zero_trust:
  client_id: "your_client_id_here"
  client_secret: "your_client_secret_here"
  team_name: "your_team_name_here"
  account_id: "your_account_id_here"

# WireGuard settings
wireguard:
  interface_name: "wg0"
  config_path: "/etc/wireguard/wg0.conf"

# UDM-Pro specific settings
udm_pro:
  wireguard_service_name: "wg-quick@wg0"
  config_backup_path: "/etc/wireguard/backup"

# General settings
refresh_interval_minutes: 60
debug: false
EOF

  echo "Please edit /etc/cfwg-zt/config.yaml with your Cloudflare Zero Trust credentials."
fi

# Install systemd service
echo "Installing systemd service..."
cp ./install/cfwg-zt.service /etc/systemd/system/
systemctl daemon-reload

echo
echo "Installation completed."
echo
echo "Next steps:"
echo "1. Create a WireGuard configuration in the UDM Pro UI"
echo "   - Go to Settings > VPN > WireGuard"
echo "   - Click 'Create New WireGuard VPN'"
echo "   - For easy setup, click 'Import' and select the file at /etc/cfwg-zt/dummy-wireguard.conf"
echo "2. Edit your configuration file at /etc/cfwg-zt/config.yaml"
echo "3. Start the service with: systemctl start cfwg-zt"
echo "4. Enable service at boot: systemctl enable cfwg-zt"
echo
echo "To view logs: journalctl -u cfwg-zt -f"
