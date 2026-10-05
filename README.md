# Secure Endpoint Walled-Garden Filter

An endpoint-level DNS filter built in Go designed to enforce a strict Zero-Trust / Walled-Garden browsing experience for child safety.

## Features
- **Default-Deny Architecture:** Drops all internet traffic except domains explicitly listed in the allowlist.
- **SafeSearch Injection:** Rewrites Google, Bing, and YouTube requests to enforce strict SafeSearch and Restricted Mode.
- **Zero-Bypass for App Protocols:** Stops non-browser network applications (Telegram, Discord, Twitter/X, native games) by dropping unauthorized domain resolution requests.
- **Upstream Protection:** Authorized traffic routes through Cloudflare Family DNS (`1.1.1.3`).

## Usage
Run with administrator privileges:
```bash
go build -o endpoint-filter.exe main.go
./endpoint-filter.exe  
## Windows Background Service (Auto-Start)

To run this filter continuously in the background without keeping a terminal open:

1. Open **PowerShell as Administrator**.
2. Run the automated setup script:
   ```powershell
   .\install-service.ps1
This script automatically pulls the required service supervisor, compiles the Go engine if needed, and configures Windows to launch it silently at startup.

To stop and uninstall the service:

PowerShell
.\uninstall-service.ps1