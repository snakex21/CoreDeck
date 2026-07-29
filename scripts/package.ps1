param(
    [switch]$SkipBuild,
    [switch]$DebugBuild
)

$ErrorActionPreference = "Stop"

$scriptsDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $scriptsDir
$dist = Join-Path $root "dist"
$exe = Join-Path $root "CoreDeck.exe"

if (-not $SkipBuild) {
    $buildArgs = @()
    if ($DebugBuild) { $buildArgs += "debug" }
    & (Join-Path $root "build.bat") @buildArgs
    if ($LASTEXITCODE -ne 0) { throw "Build failed" }
}

if (-not (Test-Path $exe)) {
    throw "CoreDeck.exe not found"
}

if (Test-Path $dist) {
    Remove-Item $dist -Recurse -Force
}

$packageRoot = Join-Path $dist "CoreDeck-Portable"
New-Item -ItemType Directory -Path $packageRoot -Force | Out-Null
Copy-Item $exe -Destination $packageRoot

$helper = Join-Path $root "data\tools\disk-helper.exe"
if (Test-Path $helper) {
    $toolsDir = Join-Path $packageRoot "data\tools"
    New-Item -ItemType Directory -Path $toolsDir -Force | Out-Null
    Copy-Item $helper -Destination $toolsDir
}

$versionText = Get-Content (Join-Path $root "internal\version\version.go") -Raw
$versionMatch = [regex]::Match($versionText, 'Current\s*=\s*"([^"]+)"')
$manifest = [ordered]@{
    product = "CoreDeck"
    version = if ($versionMatch.Success) { $versionMatch.Groups[1].Value } else { "dev" }
    builtAt = (Get-Date).ToUniversalTime().ToString("o")
    launcherSha256 = (Get-FileHash $exe -Algorithm SHA256).Hash.ToLowerInvariant()
    packageType = "public-empty-template"
    note = "Pusta baza startowa. Dane użytkownika są tworzone przy pierwszym uruchomieniu."
}
$manifest | ConvertTo-Json | Set-Content (Join-Path $packageRoot "release-manifest.json") -Encoding UTF8

$zipPath = Join-Path $dist "CoreDeck-Portable.zip"
Compress-Archive -Path "$packageRoot\*" -DestinationPath $zipPath -CompressionLevel Optimal
Remove-Item $packageRoot -Recurse -Force

Write-Host ""
Write-Host "PACZKA GOTOWA" -ForegroundColor Green
Write-Host "  $zipPath" -ForegroundColor Green
