# Secure Endpoint Filter & Walled Garden (Windows & Android)

A cross-platform, automated endpoint filtering solution engineered to enforce a zero-trust "walled-garden" browsing environment across Windows and Android devices without altering client firewall policies.

Non-whitelisted destinations (gambling, adult content, social media) are terminated by default, while educational and enterprise resources are transparently routed.

---

## 1. Windows Architecture (Local Transparent Proxy)

Modern Windows browsers bypass local DNS via built-in async stub resolvers and opportunistic DNS-over-HTTPS (DoH). 

This project solves client leakage using an inline loopback HTTP/HTTPS proxy tunnel:
- **System Proxy Hook:** Directs all browser and OS web requests to `127.0.0.1:8080`.
- **Transparent Interception (`main.go`):** Inspects the TLS `CONNECT` handshake and HTTP host headers.
- **Immediate Denial:** Non-allowlisted domains receive an instant `HTTP 403 Forbidden` (`Access Denied by Endpoint Policy`).
- **Clean Teardown:** `uninstall-service.ps1` instantly disengages the proxy, restoring system defaults with zero registry or firewall residue.

### Windows Deployment
```powershell
# Run in Administrator PowerShell
.\install-service.ps1

# Teardown & Reset
.\uninstall-service.ps1
2. Android Architecture (Native DNS-over-TLS / Private DNS)
Android 9+ includes native operating-system-level Private DNS (DoT) that enforces encryption across all mobile browsers, applications, and cellular/Wi-Fi connections.

The android-server/ service acts as an upstream DoT walled garden:

Transport: DNS-over-TLS on port 853.

Filtering Logic: Inspects inbound queries against the centralized allowlist; returns NXDOMAIN for unauthorized hosts and forwards allowed queries to upstream Cloudflare Family DNS (1.1.1.3:53).

Client Deployment: Zero APK installation required. Simply input the host address in Android's native settings under:
Settings > Network & internet > Private DNS > Private DNS provider hostname.

Android Server Deployment (Docker)
Bash
cd android-server
docker build -t dot-guard .
docker run -d -p 853:853 --name dot-guard dot-guard
Allowed Domains (Default)
wikipedia.org, wikimedia.org

khanacademy.org, kastatic.org

google.com, googleapis.com, gstatic.com

classroom.google.com, canvaslms.com, blackboard.com

pbskids.org, scratch.mit.edu, code.org, duolingo.com

nationalgeographic.com, nasa.gov

Custom domains can be added line-by-line to allowlist.txt.