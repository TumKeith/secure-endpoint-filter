# Secure Endpoint Filter & Walled Garden

A lightweight, automated endpoint filtering agent designed to enforce a zero-trust "walled-garden" browsing environment across Windows endpoints.

The filter blocks all non-whitelisted destinations (e.g., social networks, gambling, unauthorized services) while permitting approved resources (e.g., educational platforms, reference libraries) without modifying local firewall rules or requiring manual browser tweaks.

## Architecture

Traditional DNS interception on modern Windows endpoints frequently leaks due to Chromium's internal stub resolvers, DNS-over-HTTPS (DoH) auto-upgrades, and multi-homed adapter fallbacks.

This project uses an **inline loopback HTTP/HTTPS proxy tunnel**:
1. **Windows System Proxy Binding:** Directs all system web requests to `127.0.0.1:8080`.
2. **Transparent Interception (`main.go`):** Reads the TLS `CONNECT` handshake and HTTP host headers.
3. **Allowlist Policy Enforcement:** 
   - Non-allowlisted domains receive an immediate `HTTP 403 Forbidden` (`Access Denied by Endpoint Policy`).
   - Allowlisted domains establish a bidirectional TCP tunnel to upstream destinations.
4. **Clean Teardown:** Completely decouples from system networking upon uninstallation, restoring DHCP and standard proxy configurations.

## Allowed Domains (Default)

- `wikipedia.org`, `wikimedia.org`
- `khanacademy.org`, `kastatic.org`
- `google.com`, `googleapis.com`, `gstatic.com`
- `classroom.google.com`, `canvaslms.com`, `blackboard.com`
- `pbskids.org`, `scratch.mit.edu`, `code.org`, `duolingo.com`
- `nationalgeographic.com`, `nasa.gov`

Custom domains can be added line-by-line in `allowlist.txt`.

## Getting Started (Windows)

### Prerequisites
- Windows 10/11
- Go 1.20+ installed
- PowerShell running as Administrator

### Installation
Run the automated deployment script from an elevated PowerShell terminal:
```powershell
cd C:\secure-endpoint-filter
.\install-service.ps1
Uninstallation / Reset
To immediately disable the filter and restore all factory network settings:

PowerShell
.\uninstall-service.ps1
Roadmap
[x] Windows loopback HTTP/HTTPS proxy filtering engine.

[x] Zero-touch installation & automated clean teardown scripts.

[ ] Implement TCP connection pooling to eliminate tunnel handshake latency on allowed sites.

[ ] Port the walled-garden filtering engine to Android using native Private DNS (DoT) / VpnService architecture.