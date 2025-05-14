#!/bin/bash
#
# One-line installer for Cloudflare Zero Trust WireGuard Manager on UDM Pro
#
# Usage:
# curl -s https://raw.githubusercontent.com/gumbees/cfwg-zt/main/install/install-udm-pro.sh | bash
#

set -e

echo "==== Cloudflare Zero Trust WireGuard Manager Installer for UDM Pro ===="
echo

# Check if running on UDM Pro using the info command (most reliable method)
UDM_MODEL=""
if command -v info &> /dev/null; then
  UDM_MODEL=$(info | grep -i "Model:" | grep -i "Dream Machine" || echo "")
fi

if [ -z "$UDM_MODEL" ]; then
  # Fallback to older detection methods
  if [ -f "/usr/bin/ubnt-systool" ] || [ -f "/etc/unifi-os" ] || [ -d "/mnt/data/unifi-os" ] || grep -qi "udm\|ubnt" /etc/os-release 2>/dev/null; then
    echo "Detected UDM Pro using legacy identifiers."
  else
    echo "Warning: This device might not be a UDM Pro."
    echo "Common UDM Pro identifiers were not detected, but continuing anyway."
    echo "If you're certain this is a UDM Pro, press Enter to continue."
    echo "Otherwise, press Ctrl+C to cancel."
    read -r
  fi
else
  echo "Detected: $UDM_MODEL"
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "Downloading latest release..."
curl -L https://github.com/gumbees/cfwg-zt/releases/latest/download/cfwg-zt-udm-pro-arm64.tar.gz -o cfwg-zt.tar.gz

echo "Extracting package..."
tar xzf cfwg-zt.tar.gz

echo "Creating directories..."
mkdir -p /usr/local/bin /etc/cfwg-zt /etc/wireguard/backup /var/log/cfwg-zt

echo "Installing application..."
cp cfwg-zt /usr/local/bin/
chmod +x /usr/local/bin/cfwg-zt

# Copy the dummy WireGuard configuration for UDM Pro UI import
if [ -f "dummy-wireguard.conf" ]; then
  echo "Installing dummy WireGuard configuration file..."
  cp dummy-wireguard.conf /etc/cfwg-zt/
  echo "The dummy WireGuard configuration has been installed to /etc/cfwg-zt/dummy-wireguard.conf"
  echo "To use it:"
  echo "1. Go to UDM Pro UI: Settings > VPN > WireGuard > Create New > Import"
  echo "2. Select the file at /etc/cfwg-zt/dummy-wireguard.conf"
  echo "3. Click 'Add' to create the interface"
  echo "This configuration contains temporary keys that will be replaced by the application"
fi

echo "Installing systemd service..."
cp cfwg-zt.service /etc/systemd/system/
systemctl daemon-reload

# Check if WireGuard is installed and available
if ! command -v wg &> /dev/null; then
  echo "Warning: WireGuard is not installed or not in PATH."
  echo "Make sure WireGuard is properly set up in your UDM Pro UI."
fi

# Create example configuration if none exists
if [ ! -f "/etc/cfwg-zt/config.yaml" ]; then
  echo "Creating example configuration file..."
  cp config.yaml.example /etc/cfwg-zt/config.yaml.example
  
  echo
  echo "No configuration file found. Would you like to run the configuration wizard? (y/n)"
  read -r answer
  if [[ "$answer" =~ ^[Yy]$ ]]; then
    /usr/local/bin/cfwg-zt config-wizard
  else
    echo "You can create a configuration manually by editing /etc/cfwg-zt/config.yaml.example"
    echo "and then renaming it to config.yaml."
    cp config.yaml.example /etc/cfwg-zt/config.yaml
    echo "A default configuration has been created at /etc/cfwg-zt/config.yaml"
    echo "Please edit this file before starting the service."
  fi
fi

# Clean up
cd /
rm -rf "$TMP_DIR"

echo
echo "Installation complete!"
echo
echo "To start the service, run:"
echo "  systemctl start cfwg-zt"
echo
echo "To enable the service at boot, run:"
echo "  systemctl enable cfwg-zt"
echo
echo "View the logs with:"
echo "  journalctl -u cfwg-zt -f"
echo
