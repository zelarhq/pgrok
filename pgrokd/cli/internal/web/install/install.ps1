# Installs the pgrok client and runs it as a Windows service.
# Run from an Administrator PowerShell:
#
#   & ([scriptblock]::Create((irm __SERVER_URL__/install.ps1))) -Token <token> [-ForwardAddr 3000] [-BasicAuth user:password]
param(
  [Parameter(Mandatory = $true)][string]$Token,
  [string]$ForwardAddr = "3000",
  [string]$BasicAuth = "",
  [string]$RemoteAddr = "__REMOTE_ADDR__",
  [string]$InstallDir = "$env:ProgramFiles\pgrok"
)
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue" # Invoke-WebRequest is very slow with the progress bar.

$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
  throw "Installing a Windows service needs an Administrator PowerShell (right-click PowerShell > Run as administrator)."
}

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$exe = Join-Path $InstallDir "pgrok.exe"

# Stop a previous install so the binary isn't locked.
if (Test-Path $exe) { try { & $exe service uninstall 2>&1 | Out-Null } catch {} }

Write-Host "==> Downloading pgrok (windows/$arch)"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -UseBasicParsing -Uri "__DOWNLOAD_URL__/pgrok_windows_$arch.exe" -OutFile $exe
& $exe --version

$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if (($machinePath -split ";") -notcontains $InstallDir) {
  [Environment]::SetEnvironmentVariable("Path", "$machinePath;$InstallDir", "Machine")
  $env:Path += ";$InstallDir"
}

$config = Join-Path $env:USERPROFILE ".pgrok\pgrok.yml"
Write-Host "==> Writing $config"
$initArgs = @("init", "--config", $config, "--remote-addr", $RemoteAddr, "--forward-addr", $ForwardAddr, "--token", $Token)
if ($BasicAuth) { $initArgs += @("--basic-auth", $BasicAuth) }
& $exe @initArgs
if ($LASTEXITCODE -ne 0) { throw "pgrok init failed" }

Write-Host "==> Installing service"
& $exe service install --config $config
if ($LASTEXITCODE -ne 0) { throw "pgrok service install failed" }

Write-Host ""
Write-Host "pgrok is running as a Windows service and starts automatically."
Write-Host "  status:    pgrok service status"
Write-Host "  logs:      $(Join-Path (Split-Path $config) 'pgrok.log')"
Write-Host "  uninstall: pgrok service uninstall   (Administrator)"
