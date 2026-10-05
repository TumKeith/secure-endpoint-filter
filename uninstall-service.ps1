# Ensure script is running as Administrator
if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "Please run this script from an elevated Administrator PowerShell prompt!"
    exit 1
}

$ProjectDir = $PSScriptRoot
$NssmExe = Join-Path $ProjectDir "nssm.exe"

if (Test-Path $NssmExe) {
    Write-Host "[SERVICE] Stopping and removing EndpointGuard..." -ForegroundColor Yellow
    & $NssmExe stop EndpointGuard
    & $NssmExe remove EndpointGuard confirm
    Write-Host "[SERVICE] Removed successfully." -ForegroundColor Green
} else {
    Write-Warning "nssm.exe not found. Checking standard Windows service registry..."
    Stop-Service -Name EndpointGuard -ErrorAction SilentlyContinue
    sc.exe delete EndpointGuard
}