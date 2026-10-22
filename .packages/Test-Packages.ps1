[CmdletBinding()]
param(
    [string[]] $Architecture = @(),
    [ValidateSet('debian', 'rpm', 'alpine', 'aur')]
    [string[]] $Format = @('debian', 'rpm', 'alpine', 'aur')
)

$ErrorActionPreference = 'Stop'
$repository = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path.Replace('\', '/')
$work = (Join-Path ([System.IO.Path]::GetTempPath()) "xget-packages-$([guid]::NewGuid())").Replace('\', '/')
$platforms = Get-Content (Join-Path $PSScriptRoot 'platforms.json') -Raw | ConvertFrom-Json
if ($Architecture.Count -eq 0) { $Architecture = @($platforms.release) }
$savedEnvironment = @{}
$bash = if ($IsWindows) { 'C:/Program Files/Git/bin/bash.exe' } else { 'bash' }

function Set-TestEnvironment([string] $Name, [string] $Value) {
    if (-not $savedEnvironment.ContainsKey($Name)) {
        $savedEnvironment[$Name] = [Environment]::GetEnvironmentVariable($Name)
    }
    [Environment]::SetEnvironmentVariable($Name, $Value)
}

function Invoke-Checked([string] $Program, [string[]] $Arguments) {
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed with exit code $LASTEXITCODE" }
}

function Invoke-ActionScript([string] $Action, [string] $Step) {
    Set-TestEnvironment XGET_TEST_STEP $Step
    $script = & yq -r '.runs.steps[] | select(.id == strenv(XGET_TEST_STEP) or .name == strenv(XGET_TEST_STEP)) | .run' ".github/actions/$Action/action.yml"
    if ($LASTEXITCODE -ne 0 -or -not $script) { throw "Missing action script: $Action $Step" }
    $script | & $bash -s
    if ($LASTEXITCODE -ne 0) { throw "Action script failed: $Action $Step" }
}

Push-Location $repository
try {
    foreach ($arch in $Architecture) {
        if ($arch -notin $platforms.release) { throw "Unknown release architecture: $arch" }
    }
    New-Item -ItemType Directory -Force "$work/archives", "$work/tools" | Out-Null
    Set-TestEnvironment MSYS_NO_PATHCONV '1'
    Set-TestEnvironment VERSION '0.0.0'
    Set-TestEnvironment GOOS 'linux'
    Set-TestEnvironment CGO_ENABLED '0'
    foreach ($platform in $platforms) {
        $source = "$work/source/$($platform.release)"
        New-Item -ItemType Directory -Force $source | Out-Null
        Set-TestEnvironment GOARCH $platform.goarch
        Set-TestEnvironment GOARM $platform.goarm
        Invoke-Checked go @('build', '-ldflags', '-X github.com/camalot/xget/internal/cli.version=0.0.0', '-o', "$source/xget", './cmd/xget')
        Copy-Item README.md, LICENSE $source
        Invoke-Checked pandoc @('docs/_man/xget.md', '-s', '-t', 'man', '-o', "$source/xget.1")
        Invoke-Checked tar @('-czf', "$work/archives/xget_0.0.0_linux_$($platform.release).tar.gz", '-C', $source, 'xget', 'README.md', 'LICENSE', 'xget.1')
    }
    $checksums = Get-ChildItem "$work/archives/*.tar.gz" | ForEach-Object {
        "$((Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant())  $($_.Name)"
    }
    [System.IO.File]::WriteAllText("$work/archives/checksums.txt", (($checksums -join "`n") + "`n"))
    Set-TestEnvironment ARTIFACT 'local-fixtures'
    Set-TestEnvironment ARCHIVES 'true'
    Set-TestEnvironment DEST "$work/archives"
    Set-TestEnvironment GITHUB_OUTPUT "$work/archive-outputs.txt"
    [System.IO.File]::WriteAllText($env:GITHUB_OUTPUT, '')
    foreach ($arch in @($platforms.release) + @('')) {
        Set-TestEnvironment ARCHITECTURE $arch
        Invoke-ActionScript release-archives download
    }
    Set-TestEnvironment ARCHITECTURE 'unsupported'
    $rejected = $false
    try { Invoke-ActionScript release-archives download } catch { $rejected = $true }
    if (-not $rejected) { throw 'Downloader accepted an unsupported architecture' }
    Set-TestEnvironment ARCHITECTURE $platforms[0].release
    try {
        [System.IO.File]::AppendAllText("$work/archives/checksums.txt", $checksums[0] + "`n")
        $rejected = $false
        try { Invoke-ActionScript release-archives download } catch { $rejected = $true }
        if (-not $rejected) { throw 'Downloader accepted duplicate checksum entries' }
    } finally {
        [System.IO.File]::WriteAllText("$work/archives/checksums.txt", (($checksums -join "`n") + "`n"))
    }
    if ('armv7' -in $platforms.release) {
        Invoke-Checked $bash @('.packages/test-install.sh', "$work/archives", "$work/installer")
    }
    if ('debian' -in $Format) {
        $container = & docker create --platform linux/amd64 mikefarah/yq:4
        if ($LASTEXITCODE -ne 0) { throw 'Cannot create yq tool container' }
        try { Invoke-Checked docker @('cp', "${container}:/usr/bin/yq", "$work/tools/yq") }
        finally { Invoke-Checked docker @('rm', $container) }
    }
    foreach ($platform in $platforms | Where-Object release -In $Architecture) {
        foreach ($formatName in $Format) {
            $packageArch = $platform.$formatName
            if (-not $packageArch) { continue }
            Write-Output "Testing $formatName/$($platform.release)"
            $root = "$work/$formatName/$($platform.release)"
            New-Item -ItemType Directory -Force $root | Out-Null
            Set-TestEnvironment RELEASE_ARCH $platform.release
            Set-TestEnvironment RUNNER_TEMP $root
            Set-TestEnvironment WORKDIR $root
            Set-TestEnvironment ARCHIVE "$work/archives/xget_0.0.0_linux_$($platform.release).tar.gz"
            Set-TestEnvironment SHA256 (Get-FileHash $env:ARCHIVE).Hash.ToLowerInvariant()
            Set-TestEnvironment CHECKSUMS "$work/archives/checksums.txt"
            Set-TestEnvironment DOWNLOAD "$work/archives"
            Set-TestEnvironment GITHUB_OUTPUT "$root/outputs.txt"
            [System.IO.File]::WriteAllText($env:GITHUB_OUTPUT, '')
            switch ($formatName) {
                debian {
                    Invoke-Checked docker @('run', '--rm', '--platform', 'linux/amd64',
                        '--mount', "type=bind,src=$repository,dst=/repo,readonly",
                        '--mount', "type=bind,src=$work,dst=/work", '-e', "RELEASE_ARCH=$($platform.release)",
                        '-e', "ARCH=$packageArch", 'debian:bookworm-slim', 'bash', '-euc', '
                            chmod +x /work/tools/yq
                            export PATH=/work/tools:$PATH
                            mkdir -p /work/debian/$RELEASE_ARCH/src /work/debian/$RELEASE_ARCH/dist
                            tar -xzf /work/archives/xget_0.0.0_linux_${RELEASE_ARCH}.tar.gz -C /work/debian/$RELEASE_ARCH/src
                            bash /repo/.packages/debian/build-deb.sh --version 0.0.0 --arch "$ARCH" --srcdir /work/debian/$RELEASE_ARCH/src --outdir /work/debian/$RELEASE_ARCH/dist
                            test "$(dpkg-deb -f /work/debian/$RELEASE_ARCH/dist/*.deb Architecture)" = "$ARCH"
                        ')
                    $definition = & yq -o=json '.' .packages/debian/control.yml | ConvertFrom-Json
                    $paths = foreach ($asset in $definition.assets) {
                        if ($asset.type -in @('manpage', 'changelog')) { $asset.path + '.gz' } else { $asset.path }
                    }
                    [System.IO.File]::WriteAllText("$root/expected.txt", (($paths -join "`n") + "`n"))
                    Set-TestEnvironment PACKAGE "$root/dist/xget_0.0.0_${packageArch}.deb"
                    Set-TestEnvironment EXPECTED "$root/expected.txt"
                    Set-TestEnvironment PLATFORM $platform.container
                    Invoke-ActionScript debian 'Verify installation and removal'
                }
                rpm { Invoke-ActionScript rpm build }
                alpine {
                    Invoke-ActionScript alpine render
                    Invoke-Checked docker @('run', '--rm', '--platform', $platform.container,
                        '--mount', "type=bind,src=$repository,dst=/repo,readonly", '--mount', "type=bind,src=$root,dst=/work",
                        '-e', 'APKBUILD_DIR=/work/alpine/main/xget', '-e', 'REPODEST=/work/alpine/packages',
                        '-e', 'SRCDEST=/work/alpine/srcdest', '-e', 'PACKAGER=Package Test <test@example.com>',
                        '-e', 'SIGNING_KEY_NAME=test.rsa', '-e', 'VERSION=0.0.0', '-e', "ARCH=$packageArch",
                        'alpine:latest', 'sh', '-euc', '
                            apk add --no-cache alpine-sdk
                            sh /repo/.packages/alpine/build-apk.sh
                            set -- /work/alpine/packages/main/$ARCH/xget-*.apk
                            test "$#" -eq 1
                            APK_FILE="$1" sh /repo/.packages/alpine/verify-apk.sh
                        ')
                    Set-TestEnvironment WORKDIR "$root/alpine"
                    Set-TestEnvironment ARCH $packageArch
                    Invoke-ActionScript alpine collect
                }
                aur {
                    Invoke-ActionScript aur render
                    Set-TestEnvironment PKGBUILD_DIR "$root/aur"
                    Set-TestEnvironment ARCH $packageArch
                    Invoke-ActionScript aur build
                }
            }
        }
    }
    Write-Output "Package tests passed. Artifacts: $work"
} finally {
    foreach ($entry in $savedEnvironment.GetEnumerator()) {
        [Environment]::SetEnvironmentVariable($entry.Key, $entry.Value)
    }
    Pop-Location
}