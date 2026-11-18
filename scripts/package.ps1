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

# PowerShell 5 reads files without a BOM as ANSI: always say UTF-8.
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$versionText = [System.IO.File]::ReadAllText((Join-Path $root "internal\version\version.go"), $utf8NoBom)
$versionMatch = [regex]::Match($versionText, 'Current\s*=\s*"([^"]+)"')
$manifest = [ordered]@{
    product = "CoreDeck"
    version = if ($versionMatch.Success) { $versionMatch.Groups[1].Value } else { "dev" }
    builtAt = (Get-Date).ToUniversalTime().ToString("o")
    launcherSha256 = (Get-FileHash $exe -Algorithm SHA256).Hash.ToLowerInvariant()
    packageType = "public-empty-template"
    note = "Pusta baza startowa. Dane użytkownika są tworzone przy pierwszym uruchomieniu."
}
# UTF-8 without BOM, so JSON parsers and the updater read it as is.
[System.IO.File]::WriteAllText((Join-Path $packageRoot "release-manifest.json"), ($manifest | ConvertTo-Json), $utf8NoBom)

$zipPath = Join-Path $dist "CoreDeck-Portable.zip"
Compress-Archive -Path "$packageRoot\*" -DestinationPath $zipPath -CompressionLevel Optimal
Remove-Item $packageRoot -Recurse -Force

# Release assets next to the ZIP: the bare exe and SHA256SUMS.txt.
# The in-app updater requires lines for CoreDeck.exe and CoreDeck-Portable.zip.
Copy-Item $exe -Destination (Join-Path $dist "CoreDeck.exe")
$sums = @(
    "{0}  CoreDeck.exe" -f (Get-FileHash $exe -Algorithm SHA256).Hash.ToLowerInvariant()
    "{0}  CoreDeck-Portable.zip" -f (Get-FileHash $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
)
if (Test-Path $helper) {
    $sums += "{0}  disk-helper.exe" -f (Get-FileHash $helper -Algorithm SHA256).Hash.ToLowerInvariant()
}
[System.IO.File]::WriteAllText((Join-Path $dist "SHA256SUMS.txt"), (($sums -join "`n") + "`n"), $utf8NoBom)

Write-Host ""
Write-Host "PACZKA GOTOWA" -ForegroundColor Green
Write-Host "  $zipPath" -ForegroundColor Green
