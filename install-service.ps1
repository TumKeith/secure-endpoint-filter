<#
.SYNOPSIS
    Installs EndpointGuard as an automatic Windows background service using NSSM.
.DESCRIPTION
    Downloads NSSM if not present, registers the Go binary as a service,
    sets the working directory, and starts the service.
#>

# Ensure script is running as Administrator
if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "Please run this script from an elevated Administrator PowerShell prompt!"
    exit 1
}

$ProjectDir = $PSScriptRoot
$NssmExe = Join-Path $ProjectDir "nssm.exe"
$BinaryExe = Join-Path $ProjectDir "endpoint-filter.exe"

# 1. Verify binary is compiled
if (-not (Test-Path $BinaryExe)) {
    Write-Host "[BUILD] Compiling endpoint-filter.exe..." -ForegroundColor Cyan
    Set-Location $ProjectDir
    go build -o endpoint-filter.exe main.go
    if (-not (Test-Path $BinaryExe)) {
        Write-Error "Failed to build endpoint-filter.exe. Make sure Go is installed."
        exit 1
    }
}

# 2. Download precompiled nssm.exe if missing
if (-not (Test-Path $NssmExe)) {
    Write-Host "[DOWNLOAD] Fetching NSSM binary..." -ForegroundColor Cyan
    $ZipPath = Join-Path $ProjectDir "nssm.zip"
    $Url = "https://nssm.cc/release/nssm-2.24.zip"
    
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath
    
    Expand-Archive -Path $ZipPath -DestinationPath (Join-Path $ProjectDir "nssm_temp") -Force
    Copy-Item (Join-Path $ProjectDir "nssm_temp\nssm-2.24\win64\nssm.exe") -Destination $NssmExe
    
    # Cleanup temp zip
    Remove-Item -Recurse -Force (Join-Path $ProjectDir "nssm_temp")
    Remove-Item -Force $ZipPath
    Write-Host "[DOWNLOAD] NSSM downloaded successfully." -ForegroundColor Green
}

# 3. Register and Start Windows Service
Write-Host "[SERVICE] Registering EndpointGuard..." -ForegroundColor Cyan
& $NssmExe install EndpointGuard "$BinaryExe"
& $NssmExe set EndpointGuard AppDirectory "$ProjectDir"
& $NssmExe set EndpointGuard Start SERVICE_AUTO_START
& $NssmExe start EndpointGuard

Write-Host "[COMPLETE] EndpointGuard service is installed and running automatically on boot." -ForegroundColor Green