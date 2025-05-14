# Development Guide

This document provides instructions for setting up a development environment and working with this project.

## Prerequisites

- Go 1.20 or higher
- Access to a Cloudflare Zero Trust for Business account for testing
- Basic knowledge of UDM-Pro and its WireGuard implementation

## Setup Development Environment

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/cfwg-zt.git
   cd cfwg-zt
   ```

2. Install Go dependencies:
   ```
   go mod download
   ```

## Building the Application

You can build the application on various platforms using the provided Makefile or manual Go commands.

### Building on Linux

You can use the provided Makefile to build the application on Linux:

```bash
# Build for your current platform
make build

# Build specifically for UDM-Pro
make udm-pro

# Build for all supported platforms
make build-all

# Create a release package
make package
```

To build manually for UDM-Pro on Linux (which uses ARM64 architecture):

```bash
GOOS=linux GOARCH=arm64 go build -o cfwg-zt ./cmd/cfwg-zt
```

### Building on Windows

On Windows, you can use the following PowerShell commands to build the application:

```powershell
# Build for your current platform (Windows)
go build -o build/cfwg-zt.exe ./cmd/cfwg-zt

# Build for UDM-Pro (Linux ARM64)
$env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o build/cfwg-zt ./cmd/cfwg-zt

# Build for multiple platforms
# For Linux ARM64 (UDM-Pro)
$env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o build/cfwg-zt_linux_arm64 ./cmd/cfwg-zt
# For Linux AMD64
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o build/cfwg-zt_linux_amd64 ./cmd/cfwg-zt
```

#### Creating a Release Package on Windows

```powershell
# Create the build directory
mkdir -p build/release

# Build for UDM-Pro
$env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o build/cfwg-zt ./cmd/cfwg-zt

# Copy files to the release directory
Copy-Item build/cfwg-zt build/release/
Copy-Item -Recurse install build/release/
Copy-Item README.md build/release/

# If you have 7-Zip installed
# Create archive (requires 7-Zip)
7z a -ttar build/cfwg-zt.tar build/release/*
7z a -tgzip build/cfwg-zt.tar.gz build/cfwg-zt.tar
```

## Docker Development Environment

For development and testing, you can use Docker:

```bash
# Create a config.yaml file in the project root
cp config.yaml.example config.yaml
# Edit the config.yaml file with your credentials

# Run the application in Docker
./dev-run.sh
```

## Project Structure

- `cmd/cfwg-zt`: Main application entry point and CLI
- `src/cloudflare`: Cloudflare Zero Trust API interactions
- `src/wireguard`: WireGuard configuration management
- `src/udm`: UDM-Pro specific functionality
- `src/config`: Configuration handling
- `install`: Installation scripts and service files
- `.github/workflows`: CI/CD configuration

## Architecture

This application is designed to work with the UDM Pro's built-in WireGuard functionality:

1. **User-created configuration**: The WireGuard configuration is first created through the UDM Pro UI
2. **Authentication handling**: This application maintains the authentication with Cloudflare Zero Trust
3. **Key rotation**: When Cloudflare rotates keys, the application updates only the auth-related parts of the config
4. **Policy-based routing**: Network routing is handled via UDM Pro's built-in policy-based routing

### Integration with UDM Pro

The application integrates with the UDM Pro in the following ways:

1. Uses the WireGuard configuration path that the UDM Pro UI creates
2. Only updates authentication-related parts of the configuration (keys, endpoints)
3. Preserves interface settings, routing settings, and other user configurations
4. Uses UDM Pro's built-in systemd service to restart the WireGuard interface when needed
5. Does not modify the routing or NAT - this is handled via UDM Pro's policy-based routing

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests to ensure everything works: `make test` (Linux) or `go test ./...` (Windows)
5. Commit your changes: `git commit -am 'Add new feature'`
6. Push to the branch: `git push origin feature/my-feature`
7. Submit a pull request

### Windows Development Environment

For Windows developers:

1. Install Go 1.20+ from [golang.org](https://golang.org/dl/)
2. Install Git for Windows from [git-scm.com](https://git-scm.com/download/win)
3. Clone the repository:
   ```powershell
   git clone https://github.com/yourusername/cfwg-zt.git
   cd cfwg-zt
   ```
4. Install dependencies:
   ```powershell
   go mod download
   ```
5. Run tests:
   ```powershell
   go test ./...
   ```
6. Build for Windows (for local testing):
   ```powershell
   go build -o build/cfwg-zt.exe ./cmd/cfwg-zt
   ```
7. Build for UDM Pro (Linux ARM64):
   ```powershell
   $env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o build/cfwg-zt ./cmd/cfwg-zt
   ```

### Linux Development Environment

For Linux developers:

1. Install Go 1.20+ through your package manager or from [golang.org](https://golang.org/dl/)
2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/cfwg-zt.git
   cd cfwg-zt
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Use the Makefile:
   ```bash
   # Run tests
   make test
   
   # Build for UDM Pro
   make udm-pro
   ```

### Code Style

Follow standard Go coding conventions:
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for public functions
- Write tests for new functionality

## Testing

Run tests with:

```
go test ./...
```

## Deployment to UDM-Pro

1. Build the binary as described above
2. Transfer the binary to your UDM-Pro:
   ```
   scp cfwg-zt root@<udm-pro-ip>:/root/
   ```

3. SSH into your UDM-Pro:
   ```
   ssh root@<udm-pro-ip>
   ```

4. Make the binary executable:
   ```
   chmod +x /root/cfwg-zt
   ```

5. Create a configuration file and start the application

## Setting Up as a Persistent Service

To ensure the application runs on UDM-Pro startup, you can add it to `/etc/rc.local` or create a proper systemd service.
