<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Cloudflare Zero Trust WireGuard Manager for UDM-Pro

This project is a Go application designed to run on a UDM-Pro device. It maintains a WireGuard configuration authenticated to a Cloudflare Zero Trust for Business tenant, allowing LAN traffic to be forwarded over a Cloudflare WARP tunnel.

## Technical Context

- The application runs on Ubiquiti Dream Machine Pro (UDM-Pro)
- It uses the UDM-Pro's built-in WireGuard capabilities
- It needs to authenticate with Cloudflare Zero Trust and maintain authentication
- It needs to manage WireGuard keys and handle rotation
- Go is the primary language used

## API Considerations

- Cloudflare Zero Trust API for device authentication and key management
- UDM-Pro API/CLI for managing WireGuard configurations
- System-level interactions for managing services and network configuration

## Code Style Preferences

- Follow Go's official style guide and idiomatic Go practices
- Use consistent error handling patterns
- Implement proper logging for monitoring and debugging
- Organize the code into logical packages that reflect the application's architecture
