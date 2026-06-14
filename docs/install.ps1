# Pitara installer for Windows — detects your arch, downloads the right binary, adds it to PATH.
# Usage (in PowerShell):  irm https://pitara.dev/install.ps1 | iex
$ErrorActionPreference = "Stop"
$repo = "sailingsam/pitara"

# --- detect architecture ---
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $arch = "arm64" } else { $arch = "amd64" }
$asset = "pitara_windows_$arch.exe"
$url = "https://github.com/$repo/releases/latest/download/$asset"

# --- install dir ---
$dir = "$env:LOCALAPPDATA\Pitara"
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$dest = "$dir\pitara.exe"

Write-Host "Pitara: downloading $asset ..."
Invoke-WebRequest -Uri $url -OutFile $dest

# --- add to user PATH if missing ---
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$dir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$dir", "User")
  Write-Host "Pitara: added $dir to your PATH."
}

Write-Host "Pitara: installed to $dest"
Write-Host "Done! Restart your terminal, then run:  pitara --help"
