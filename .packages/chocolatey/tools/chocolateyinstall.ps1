$ErrorActionPreference = 'Stop'

$packageName = 'xget'
$owner       = 'camalot'
$repo        = 'xget'
$version     = $env:ChocolateyPackageVersion

$osArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
$arch   = if ($osArch -eq [System.Runtime.InteropServices.Architecture]::Arm64) { 'arm64' } else { 'amd64' }

$zipName    = "xget_${version}_windows_${arch}.zip"
$releaseUrl = "https://github.com/$owner/$repo/releases/download/v$version"
$url        = "$releaseUrl/$zipName"

# verify against the checksums.txt published alongside the release rather than
# embedding a checksum in this script, since this file isn't regenerated per-release
$checksumsContent = (Invoke-WebRequest -Uri "$releaseUrl/checksums.txt" -UseBasicParsing).Content
$checksumLine = ($checksumsContent -split "`r?`n") | Where-Object { $_ -match [regex]::Escape($zipName) } | Select-Object -First 1
if (-not $checksumLine) {
  throw "Could not find a checksum for $zipName in $releaseUrl/checksums.txt"
}
$checksum = ($checksumLine -split '\s+')[0]

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

Install-ChocolateyZipPackage -PackageName $packageName -Url $url -UnzipLocation $toolsDir -Checksum $checksum -ChecksumType 'sha256'
