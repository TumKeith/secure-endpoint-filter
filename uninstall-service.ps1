$TaskName = "EndpointGuard"

Write-Host "[1/2] Stopping service..." -ForegroundColor Yellow
Stop-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
Stop-Process -Name "endpoint-filter" -Force -ErrorAction SilentlyContinue
Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue

Write-Host "[2/2] Turning off Windows system proxy..." -ForegroundColor Yellow
$ProxyReg = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings"
Set-ItemProperty -Path $ProxyReg -Name ProxyEnable -Value 0 -Type DWord
Remove-ItemProperty -Path $ProxyReg -Name ProxyServer -ErrorAction SilentlyContinue

Write-Host "[SUCCESS] Proxy disabled. All settings are back to normal." -ForegroundColor Green