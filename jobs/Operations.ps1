# 1. Self-Elevation (Request Admin Access)
if (!([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Start-Process powershell.exe -ArgumentList ("-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"") -Verb RunAs
    exit
}

# 2. Stop history recording (Stealth)
Set-PSReadLineOption -HistorySaveStyle SaveNothing

# 3. Create the Firewall Block Rule
# This blocks all outbound traffic for all profiles (Domain, Private, Public)
Write-Host "Activating Total Internet Block..." -ForegroundColor Red
New-NetFirewallRule -DisplayName "BlockEverythingOutbound" `
    -Direction Outbound `
    -Action Block `
    -Enabled True `
    -Description "Total internet kill switch"

# 4. Cleanup history
$historyPath = (Get-PSReadLineOption).HistorySavePath
if (Test-Path $historyPath) { Remove-Item $historyPath -Force }
Clear-History

Write-Host "Internet has been successfully blocked." -ForegroundColor Green
