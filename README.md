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