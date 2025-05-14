# UDM Pro Deployment Instructions

This package contains the Cloudflare Zero Trust WireGuard Manager for UDM-Pro.

## Quick Installation

1. Extract this tarball on your UDM-Pro device:
   ```bash
   tar xzf cfwg-zt.tar.gz
   cd cfwg-zt
   ```

2. Run the installation script:
   ```bash
   chmod +x install.sh
   ./install.sh
   ```

3. Create and edit the configuration file:
   ```bash
   cp config.yaml.example /etc/cfwg-zt/config.yaml
   nano /etc/cfwg-zt/config.yaml
   ```

4. Start and enable the service:
   ```bash
   systemctl start cfwg-zt
   systemctl enable cfwg-zt
   ```

5. Verify the service is running:
   ```bash
   systemctl status cfwg-zt
   ```

## Manual Installation

If the installation script fails, you can manually install as follows:

1. Create required directories:
   ```bash
   mkdir -p /usr/local/bin /etc/cfwg-zt /etc/wireguard/backup
   ```

2. Copy the files:
   ```bash
   cp cfwg-zt /usr/local/bin/
   chmod +x /usr/local/bin/cfwg-zt
   cp cfwg-zt.service /etc/systemd/system/
   cp config.yaml.example /etc/cfwg-zt/config.yaml
   chmod 600 /etc/cfwg-zt/config.yaml
   ```

3. Edit the configuration file:
   ```bash
   nano /etc/cfwg-zt/config.yaml
   ```

4. Start and enable the service:
   ```bash
   systemctl daemon-reload
   systemctl start cfwg-zt
   systemctl enable cfwg-zt
   ```

See the full README for comprehensive documentation and setup instructions.
