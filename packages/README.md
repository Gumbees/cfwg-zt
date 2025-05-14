# Deployment Packages

This directory contains ready-to-deploy packages for different platforms:

- `cfwg-zt-udm-pro-arm64.tar.gz` - Complete package for UDM-Pro (ARM64)
- `cfwg-zt-linux-amd64.tar.gz` - Complete package for Linux (x86-64)
- `cfwg-zt-windows-amd64.zip` - Windows package (x86-64)

## UDM-Pro Installation

1. Transfer the tarball to your UDM-Pro:
   ```bash
   scp cfwg-zt-udm-pro-arm64.tar.gz root@<UDM-Pro-IP>:/tmp/
   ```

2. SSH into the UDM-Pro and extract the package:
   ```bash
   ssh root@<UDM-Pro-IP>
   cd /tmp
   tar xzf cfwg-zt-udm-pro-arm64.tar.gz
   cd cfwg-zt
   ```

3. Follow the installation instructions in the README.md file.

## Linux Installation

Similar to UDM-Pro, but use the `cfwg-zt-linux-amd64.tar.gz` package.

## Windows Usage

1. Extract the zip file
2. Create a configuration file based on the example
3. Run from command prompt or PowerShell
