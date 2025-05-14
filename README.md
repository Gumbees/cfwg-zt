# Cloudflare Zero Trust WireGuard Manager for UDM-Pro

This application is designed to run on a Ubiquiti UDM-Pro device to maintain a WireGuard configuration authenticated to a Cloudflare Zero Trust for Business tenant. It works with a WireGuard configuration that you create in the UDM Pro UI and keeps it authenticated to Cloudflare Zero Trust.

## Enhancements

- Runs on Ubiquiti UDM-Pro
- Works with UDM Pro's built-in WireGuard UI configuration
- Automatically authenticates with Cloudflare Zero Trust for Business
- Manages WireGuard secrets and handles rotation
- Preserves your existing UDM Pro WireGuard settings
- Compatible with UDM Pro's policy-based routing

## Usage

The application provides several command-line options:

```
Usage:
  cfwg-zt [command]

Available Commands:
  help        Help about any command
  setup       Set up a new configuration file
  start       Start the service
  status      Check the status of the WireGuard connection
  version     Print the version number

Flags:
  -c, --config string   Path to config file (default is /etc/cfwg-zt/config.yaml)
  -d, --debug           Enable debug mode
  -h, --help            help for cfwg-zt
```

### Starting the Service

To start the service directly (useful for testing):

```bash
cfwg-zt start
```

For production use, it's better to use the systemd service:

```bash
systemctl start cfwg-zt
```

### Checking Status

To check if the WireGuard tunnel is connected to Cloudflare Zero Trust:

```bash
cfwg-zt status
```

### Viewing Logs

```bash
# View logs via systemd
journalctl -u cfwg-zt -f

# View the application log file directly
cat /var/log/cfwg-zt/cfwg-zt.log
```

### Troubleshooting

If you encounter issues:

1. Check the application logs:
   ```bash
   journalctl -u cfwg-zt -f
   ```

2. Verify your Cloudflare Zero Trust credentials:
   - Ensure the client_id and client_secret are correct
   - Make sure your account_id and team_name are accurate
   - Check that your credentials have the necessary permissions

3. Ensure WireGuard is properly configured:
   - Check the interface status: `wg show`
   - Verify the configuration file exists: `cat /etc/wireguard/wg0.conf`
   - Make sure the interface is running: `ip a show wg0`

4. Check network connectivity:
   - Verify DNS is working: `nslookup api.cloudflare.com`
   - Check API connectivity: `curl -I https://api.cloudflare.com`

5. Enable debug mode to get more verbose logs:
   ```bash
   # Edit config to enable debug mode
   sed -i 's/debug: false/debug: true/' /etc/cfwg-zt/config.yaml
   
   # Restart the service
   systemctl restart cfwg-zt
   
   # Or run manually with debug flag
   /usr/local/bin/cfwg-zt -d start
   ```

6. Common issues and solutions:

   **Problem**: Error authenticating with Cloudflare
   **Solution**: Double-check your Cloudflare Zero Trust credentials and ensure your UDM Pro can reach api.cloudflare.com

   **Problem**: WireGuard interface isn't showing in UDM Pro UI
   **Solution**: Make sure you've created the interface in the UI first before running the application

   **Problem**: Traffic isn't routing through the tunnel
   **Solution**: Verify your policy-based routing configuration in Settings > Routing & Firewall > Routing

   **Problem**: "Interface already exists" errors
   **Solution**: Make sure you're using the correct interface name in config.yaml that matches the UI-created interface

7. Get status information:
   ```bash
   # Check application status
   /usr/local/bin/cfwg-zt status
   
   # Check systemd service status
   systemctl status cfwg-zt
   ```d to run on a Ubiquiti UDM-Pro device to maintain a WireGuard configuration authenticated to a Cloudflare Zero Trust for Business tenant. This allows forwarding of all LAN traffic over a Cloudflare WARP tunnel.

## Enhancements

- Runs on Ubiquiti UDM-Pro
- Uses built-in WireGuard functionality
- Automatically authenticates with Cloudflare Zero Trust for Business
- Manages WireGuard secrets and handles rotation
- Forwards LAN traffic through Cloudflare WARP tunnel

## Requirements

- Ubiquiti Dream Machine Pro (UDM-Pro)
- Cloudflare Zero Trust for Business account
- Go 1.20+ for development

## Installation

### Quick Install
```bash
# Download the latest release
curl -L https://github.com/gumbees/cfwg-zt/releases/latest/download/cfwg-zt.tar.gz -o cfwg-zt.tar.gz

# Extract the archive
tar xzf cfwg-zt.tar.gz
cd cfwg-zt

# Run the installation script
chmod +x install/install.sh
./install/install.sh

# Edit configuration
nano /etc/cfwg-zt/config.yaml

# Start and enable the service
systemctl start cfwg-zt
systemctl enable cfwg-zt
```

### Manual Installation

1. Download the latest release binary from the [releases page](https://github.com/gumbees/cfwg-zt/releases)
2. Transfer it to your UDM-Pro using SCP or another file transfer method
3. Make the binary executable: `chmod +x cfwg-zt`
4. Move it to a suitable location: `mv cfwg-zt /usr/local/bin/`
5. Create a configuration directory: `mkdir -p /etc/cfwg-zt`
6. Create a configuration file (see Configuration section)
7. Install and start the systemd service

## Configuration

Create a configuration file at `/etc/cfwg-zt/config.yaml` with the following structure:

```yaml
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
```

### Getting Cloudflare Zero Trust Credentials

1. Log in to your Cloudflare dashboard at [dash.cloudflare.com](https://dash.cloudflare.com)

2. Get your Account ID:
   - Click on your profile in the top-right corner
   - Select the account you want to use
   - The Account ID is shown in the URL: `https://dash.cloudflare.com/<ACCOUNT_ID>/...`

3. Get your Team name:
   - Navigate to Zero Trust (from the sidebar)
   - Your Team name is shown in the URL: `https://dash.cloudflare.com/<ACCOUNT_ID>/zero-trust/<TEAM_NAME>/...`
   - This is usually your company name or the name you chose during setup

4. Create service credentials:
   - Navigate to Zero Trust > Settings > Authentication
   - Scroll to "Service Auth" or "Service Tokens"
   - Click "Create Service Token" 
   - Give it a name like "UDM-Pro-WARP"
   - Copy the Client ID and Client Secret - you'll need these for your config
   - Make sure the token has the necessary permissions (Device Management and WARP)

5. Add these values to your configuration file:
   ```yaml
   cloudflare_zero_trust:
     client_id: "your_client_id_here"
     client_secret: "your_client_secret_here"
     team_name: "your_team_name_here"
     account_id: "your_account_id_here"
   ```

## Complete Setup Guide

### 1. Create a WireGuard configuration in the UDM Pro UI

First, set up your WireGuard configuration through the UDM Pro UI:

1. Log in to your UDM Pro admin interface (typically at https://192.168.1.1 or https://unifi)
2. Navigate to Settings > VPN > WireGuard
3. Click "Create New WireGuard VPN" button
4. Configure the following settings:
   - **Name**: Choose a name (e.g., "CloudflareZT")
   - **WireGuard Interface IPv4**: Enter a private IP address (e.g., "100.64.0.1/32")
   - **WireGuard Interface IPv6**: Leave blank or as default
   - **Listen Port**: Choose a port (e.g., 51820)
   - **WireGuard Private Key**: Leave blank (will be replaced by the application)
   - **Firewall**: Configure as needed for your network 
   - **Peers**: Leave empty (will be configured by the application)

5. Click "Add" to create the interface
6. Make note of the interface name (e.g., "wg0") and its configuration path (typically "/etc/wireguard/wg0.conf")

### 2. Build and transfer the application to UDM Pro

1. Build the application for ARM64:
   
   **On Linux:**
   ```bash
   GOOS=linux GOARCH=arm64 go build -o cfwg-zt ./cmd/cfwg-zt
   ```
   
   **On Windows:**
   ```powershell
   $env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o cfwg-zt ./cmd/cfwg-zt
   ```

2. Create necessary files for deployment:
   - Create a config.yaml file based on config.yaml.example
   - Add your Cloudflare Zero Trust credentials

3. Transfer files to your UDM Pro:
   
   **On Linux:**
   ```bash
   scp cfwg-zt config.yaml install/install.sh root@192.168.1.1:/tmp/
   ```
   
   **On Windows:**
   ```powershell
   scp cfwg-zt config.yaml install/install.sh root@192.168.1.1:/tmp/
   ```

### 3. Install and configure on UDM Pro

1. SSH into your UDM Pro:
   ```bash
   ssh root@192.168.1.1
   ```

2. Create directories and install:
   ```bash
   # Create required directories
   mkdir -p /usr/local/bin /etc/cfwg-zt /etc/wireguard/backup

   # Move executable
   mv /tmp/cfwg-zt /usr/local/bin/
   chmod +x /usr/local/bin/cfwg-zt

   # Move config
   mv /tmp/config.yaml /etc/cfwg-zt/
   chmod 600 /etc/cfwg-zt/config.yaml
   
   # Move and execute install script
   mv /tmp/install.sh /tmp/
   chmod +x /tmp/install.sh
   /tmp/install.sh
   ```

3. Start the service:
   ```bash
   systemctl start cfwg-zt
   systemctl enable cfwg-zt
   ```

4. Check if the service is running:
   ```bash
   systemctl status cfwg-zt
   ```

5. Check the logs to ensure proper operation:
   ```bash
   journalctl -u cfwg-zt -f
   ```

### 4. Verify WireGuard connection in UDM Pro UI

1. Go back to the UDM Pro UI
2. Navigate to Settings > VPN > WireGuard
3. Check the status of your WireGuard interface - it should show as "Running"
4. You can also verify the connection status using the command line on the UDM Pro:
   ```bash
   /usr/local/bin/cfwg-zt status
   ```

### 5. Configure policy-based routing in UDM Pro UI

Now that your WireGuard interface is authenticated and connected to Cloudflare Zero Trust, you can set up policy-based routing to send specific traffic through this tunnel:

1. In the UDM Pro UI, go to Settings > Routing & Firewall > Routing
2. Click "Create New Route"
3. Configure the route:
   - **Name**: Choose a descriptive name (e.g., "CloudflareZT-Route")
   - **Source Networks**: Select your local networks that should use the tunnel
   - **Destination Networks**: Define which traffic should go through the tunnel (e.g., "0.0.0.0/0" for all traffic)
   - **Outbound Interface**: Select the WireGuard interface you created
   - **Next Hop**: Leave as "Default"
   - **Priority**: Set as needed (lower numbers have higher priority)

4. Click "Add" to create the route policy

5. Test the connection:
   - Connect devices to the specified source networks
   - Visit [https://cloudflare.com/cdn-cgi/trace](https://cloudflare.com/cdn-cgi/trace) to verify your traffic is going through Cloudflare WARP
   - The "warp" field should show "on" if properly configured

### 6. Monitor and maintain

- Check the service logs to monitor for any issues:
  ```bash
  journalctl -u cfwg-zt -f
  ```

- Check WireGuard status:
  ```bash
  wg show
  ```

- The application will automatically refresh the authentication with Cloudflare Zero Trust based on the configured interval (default: 60 minutes)

### 7. Updating the application

When a new version of the application is released:

1. Download or build the new version
2. SSH into your UDM Pro:
   ```bash
   ssh root@192.168.1.1
   ```

3. Stop the current service:
   ```bash
   systemctl stop cfwg-zt
   ```

4. Backup your configuration:
   ```bash
   cp /etc/cfwg-zt/config.yaml /etc/cfwg-zt/config.yaml.bak
   ```

5. Transfer and replace the executable:
   ```bash
   # On your local machine
   scp cfwg-zt root@192.168.1.1:/usr/local/bin/
   
   # On UDM Pro
   chmod +x /usr/local/bin/cfwg-zt
   ```

6. Restart the service:
   ```bash
   systemctl start cfwg-zt
   ```

7. Verify the service is working:
   ```bash
   systemctl status cfwg-zt
   journalctl -u cfwg-zt -f
   ```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for details on developing and building the application.
