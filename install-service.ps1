# Must run as Administrator
if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "Please run as Administrator!"
    exit 1
}

$ProjectDir = $PSScriptRoot
$BinaryExe = Join-Path $ProjectDir "endpoint-filter.exe"
$TaskName = "EndpointGuard"

# 1. Compile binary
Write-Host "[1/3] Compiling Go proxy filter..." -ForegroundColor Cyan
Set-Location $ProjectDir
go build -o endpoint-filter.exe main.go

# 2. Register persistent background task under SYSTEM
Write-Host "[2/3] Registering background service..." -ForegroundColor Cyan
Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue

$Action = New-ScheduledTaskAction -Execute $BinaryExe -WorkingDirectory $ProjectDir
$Trigger = New-ScheduledTaskTrigger -AtStartup
$Principal = New-ScheduledTaskPrincipal -UserId "NT AUTHORITY\SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit (New-TimeSpan -Days 0)

Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Principal $Principal -Settings $Settings | Out-Null
Start-ScheduledTask -TaskName $TaskName

# 3. Direct Chrome, Edge, and Windows system traffic to 127.0.0.1:8080
Write-Host "[3/3] Setting Windows system proxy to 127.0.0.1:8080..." -ForegroundColor Cyan
$ProxyReg = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings"
Set-ItemProperty -Path $ProxyReg -Name ProxyEnable -Value 1 -Type DWord
Set-ItemProperty -Path $ProxyReg -Name ProxyServer -Value "127.0.0.1:8080" -Type String

# Hard-kill browsers once so they immediately pick up the system proxy
cmd.exe /c "taskkill /F /IM chrome.exe /T 2>nul" | Out-Null
cmd.exe /c "taskkill /F /IM msedge.exe /T 2>nul" | Out-Null

Write-Host "[SUCCESS] Endpoint Guard is active. Chrome will now intercept every site." -ForegroundColor Green