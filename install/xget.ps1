#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Downloads and installs the latest (or a specific) xget release
    binary for this machine's architecture from GitHub Releases.
.EXAMPLE
    irm https://raw.githubusercontent.com/camalot/xget/develop/install/xget.ps1 | iex
.EXAMPLE
    ./install.ps1 -Version v0.1.0 -InstallDir C:\tools\bin
#>
[CmdletBinding()]
param(
	[Parameter(Mandatory=$false, ParameterSetName="Install")]
	[string]$InstallDir = (Join-Path $HOME ".local\bin"),
	[Parameter(Mandatory=$false, ParameterSetName="Install")]
	[string]$Version = "",
	[Parameter(Mandatory=$false, ParameterSetName="Install")]
	[switch]$NoChecksum,
	[Parameter(Mandatory=$false, ParameterSetName="Help")]
	[switch]$Help
)

$ErrorActionPreference = "Stop"
$Repo = "camalot/xget"
$Binary = "xget"

function Get-DetectedArch {
	switch ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture) {
		"X64" { return "amd64" }
		"Arm64" { return "arm64" }
		default { return "" }
	}
}

function Show-Help {
	Write-Host "Usage: install.ps1 [-InstallDir <path>] [-Version <tag>] [-NoChecksum] [-Help]"
	Write-Host ""
	Write-Host "Options:"
	Write-Host "  -InstallDir <path>   Install directory (default: $HOME\.local\bin)"
	Write-Host "  -Version <tag>       Install a specific release tag, e.g. v0.1.0 (default: latest)"
	Write-Host "  -NoChecksum          Skip script checksum verification (not recommended)"
	Write-Host "  -Help                Show this help message"
	exit 0
}

if ($Help) {
	Show-Help
}

function Test-ScriptChecksum {
	$checksumUrl = "https://raw.githubusercontent.com/$Repo/main/install/xget.ps1.sha256"
	try {
		$expected = (Invoke-RestMethod -Uri $checksumUrl -Headers @{ "User-Agent" = "xget-install-script" }).Trim()
		$actual = (Get-FileHash -Path $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant()
		if ($actual -ne $expected) {
			Write-Host "error: script checksum mismatch (expected $expected, got $actual)" -ForegroundColor Red
			exit 1
		}
	} catch {
		Write-Warning $_.Exception.Message
		Write-Warning "could not download script checksum from $checksumUrl; use -NoChecksum to skip this check"
		exit 1
	}
}

if (-not $NoChecksum) {
	Test-ScriptChecksum
}

$arch = Get-DetectedArch
$script:tag = ""
$script:asset = ""

# Prints a block of diagnostics the user can paste directly into the
# GitHub issue linked by Fail-Unsupported.
function Write-IssueBlock {
	@"

### Install script diagnostics

- Repository: $Repo
- Script: install.ps1
- Requested version: $(if ($Version) { $Version } else { "latest" })
- OS: Windows ($([System.Environment]::OSVersion.VersionString))
- Detected .NET arch: $([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture)
- Detected Arch: $(if ($arch) { $arch } else { "<unrecognized>" })
- Resolved tag: $(if ($script:tag) { $script:tag } else { "<none>" })
- Attempted asset: $(if ($script:asset) { $script:asset } else { "<none>" })
"@
}

function Exit-Unsupported {
	param([string]$Message)
	$plat = "windows/$(if ($arch) { $arch } else { [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture })"
	$issueTitle = [System.Uri]::EscapeDataString("Unsupported platform: $plat")
	Write-Host "error: $Message" -ForegroundColor Red
	Write-Host ""
	Write-Host "No supported xget release was found for this machine."
	Write-Host "Please open an issue: https://github.com/$Repo/issues/new?title=$issueTitle"
	Write-Host ""
	Write-Host "Paste the block below into the issue:"
	Write-Host (Write-IssueBlock)
	exit 1
}

if (-not $arch) {
	Exit-Unsupported "unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture)"
}

if ($Version) {
	$script:tag = $Version
} else {
	Write-Host "Looking up the latest release..."
	try {
		$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ "User-Agent" = "xget-install-script" }
		$script:tag = $release.tag_name
	} catch {
		Exit-Unsupported "could not determine the latest release tag from the GitHub API ($($_.Exception.Message))"
	}
}

if (-not $script:tag) {
	Exit-Unsupported "could not determine the latest release tag from the GitHub API"
}

$versionNum = $script:tag.TrimStart("v")
$script:asset = "${Binary}_${versionNum}_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$($script:tag)/$($script:asset)"
$checksumsUrl = "https://github.com/$Repo/releases/download/$($script:tag)/checksums.txt"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("xget-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null
try {
	$assetPath = Join-Path $tmpDir $script:asset
	Write-Host "Downloading $($script:asset) ($($script:tag))..."
	try {
		Invoke-WebRequest -Uri $url -OutFile $assetPath -UseBasicParsing
	} catch {
		Exit-Unsupported "no release asset found at $url"
	}

	$checksumsPath = Join-Path $tmpDir "checksums.txt"
	try {
		Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
		$line = Select-String -Path $checksumsPath -Pattern ([regex]::Escape($script:asset)) | Select-Object -First 1
		if ($line) {
			$expected = ($line.Line -split '\s+')[0].ToLowerInvariant()
			$actual = (Get-FileHash -Path $assetPath -Algorithm SHA256).Hash.ToLowerInvariant()
			if ($actual -ne $expected) {
				Write-Host "error: checksum mismatch for $($script:asset) (expected $expected, got $actual)" -ForegroundColor Red
				exit 1
			}
		}
	} catch {
		Write-Warning "could not download checksums.txt; skipping checksum verification"
	}

	Write-Host "Extracting..."
	Expand-Archive -Path $assetPath -DestinationPath $tmpDir -Force

	if (-not (Test-Path $InstallDir)) {
		New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
	}
	$destPath = Join-Path $InstallDir "$Binary.exe"
	Copy-Item -Path (Join-Path $tmpDir "$Binary.exe") -Destination $destPath -Force

	Write-Host "Installed $Binary $($script:tag) to $destPath"
	$pathDirs = $env:Path -split ";"
	if ($pathDirs -notcontains $InstallDir.TrimEnd("\")) {
		Write-Host ""
		Write-Host "note: $InstallDir is not on your PATH - add it, e.g.:"
		Write-Host "  [Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$InstallDir`", 'User')"
	}
} finally {
	Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
