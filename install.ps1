$ErrorActionPreference = "Stop"

$Repo = "It-Shu/pong"
$BinaryName = "pong-terminal.exe"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\pong-terminal" }
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$version = $Release.tag_name
if (-not $version) {
    throw "Could not resolve latest release version"
}

$asset = "pong-terminal_${version}_windows_$arch.zip"
$downloadUrl = "https://github.com/$Repo/releases/download/$version/$asset"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("pong-terminal-" + [guid]::NewGuid().ToString("N"))
$zipPath = Join-Path $tempDir $asset

New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

try {
    Write-Host "Downloading $asset..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
    Copy-Item -Path (Join-Path $tempDir $BinaryName) -Destination (Join-Path $InstallDir $BinaryName) -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $segments = @()
    if ($userPath) {
        $segments = $userPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    }

    if ($segments -notcontains $InstallDir) {
        $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to user PATH"
    }

    Write-Host ""
    Write-Host "Installed to: $(Join-Path $InstallDir $BinaryName)"
    Write-Host "Run: pong-terminal"
}
finally {
    if (Test-Path $tempDir) {
        Remove-Item -LiteralPath $tempDir -Recurse -Force
    }
}
