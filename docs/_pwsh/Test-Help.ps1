[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'xget.ps1')
$xml = [xml] (Get-Content (Join-Path $PSScriptRoot 'xget-help.xml') -Raw)
$topics = $xml.SelectNodes('/*[local-name()="helpItems"]/*[local-name()="command"]')
$aliases = @(Get-Alias xget, xget-*)
foreach ($topic in $topics) {
    $name = $topic.details.name
    $help = Get-Help $name -Full
    if ($help.Synopsis -ne $topic.details.description.para) { throw "Help not loaded: $name" }
    if (-not $help.examples.example.code) { throw "Missing example: $name" }
    foreach ($parameter in $topic.parameters.parameter) {
        $actual = Get-Help $name -Parameter $parameter.name
        if (-not $actual.description.Text) { throw "Missing parameter help: $name $($parameter.name)" }
    }
    if ($name -notin $aliases.Definition) { throw "Missing help alias: $name" }
}
foreach ($alias in $aliases) {
    if ($alias.Definition -notin $topics.details.name) { throw "Undocumented alias: $($alias.Name)" }
}

foreach ($pair in @(
    @('root.go', 'Invoke-XGet'),
    @('list.go', 'Invoke-XGetList'),
    @('uninstall.go', 'Invoke-XGetUninstall'),
    @('upgrade.go', 'Invoke-XGetUpgrade', 'newUpgradeCommand'),
    @('upgrade.go', 'Invoke-XGetSelfUpdate', 'newSelfUpdateCommand'),
    @('config.go', 'Invoke-XGetConfig')
)) {
    $source = Get-Content (Join-Path $PSScriptRoot "../../internal/cli/$($pair[0])") -Raw
    if ($pair.Count -gt 2) {
        $source = [regex]::Match($source, ('(?ms)^func ' + $pair[2] + '\(.*?(?=^func |\z)')).Value
    }
    $names = ($topics | Where-Object { $_.details.name -eq $pair[1] }).parameters.parameter.name
    foreach ($match in [regex]::Matches($source, '(?:Bool|String)(?:Slice)?(?:VarP?|P)?\([^\r\n]*?"(?<name>[a-z][a-z0-9_.-]+)(?<!\.)"')) {
        if ($match.Groups['name'].Value -notin $names) { throw "Missing CLI flag: $($pair[1]) $($match.Groups['name'].Value)" }
    }
}

$asset = (Get-Help xget-install -Parameter asset).description.Text -join "`n"
if (-not $asset.Contains('~^tool_.*\.zip$')) { throw 'Regex wildcard was changed during Markdown conversion' }
$fileHelp = (Get-Help xget-install -Full).examples.example.code
if (-not $fileHelp.Contains('{{.Arch}}/*.exe')) { throw 'Template or glob was changed' }
$rootHelp = (Get-Help xget -Full).alertSet.alert.Text -join "`n"
foreach ($reference in @('ENVIRONMENT', 'CONFIGURATION', 'EXIT STATUS', 'FOR MAINTAINERS', 'XGET_GITHUB_TOKEN', 'clobber_global')) {
    if (-not $rootHelp.Contains($reference)) { throw "Missing manual reference: $reference" }
}
if ((Get-Help xget-config-set -Parameter SECTION).required -ne 'true') { throw 'Required argument metadata is missing' }

$script:receivedArguments = @()
$script:fakeExitCode = 0
function Get-Command {
    param($Name, $CommandType, $ErrorAction)
    if ($Name -ne 'xget' -or $CommandType -ne 'Application') { throw 'Shim must resolve only the native application' }
    {
        $script:receivedArguments = @($args)
        $global:LASTEXITCODE = $script:fakeExitCode
        'native stdout'
    }
}

try {
    foreach ($alias in $aliases) {
        $suffix = $alias.Definition.Substring('Invoke-XGet'.Length)
        $prefix = switch ($suffix) {
            '' { @() }
            'SelfUpdate' { @('self-update') }
            { $_ -like 'Config*' -and $_ -ne 'Config' } { @('config', $_.Substring(6).ToLowerInvariant()) }
            { $_ -like 'Completion*' -and $_ -ne 'Completion' } { @('completion', $_.Substring(10).ToLowerInvariant()) }
            default { @($_.ToLowerInvariant()) }
        }
        $arguments = @('owner/repo', '--asset', '~^tool_.*\.zip$', '--to', 'C:\path with spaces', '--quiet=false', '', '--', '-literal')
        $output = & $alias.Name @arguments
        $expected = @($prefix) + $arguments
        if (($script:receivedArguments | ConvertTo-Json -Compress) -cne ($expected | ConvertTo-Json -Compress)) {
            throw "Arguments changed by $($alias.Name)"
        }
        if ($output -ne 'native stdout') { throw "Stdout changed by $($alias.Name)" }
    }
    foreach ($exitCode in @(0, 1, 16)) {
        $script:fakeExitCode = $exitCode
        $null = xget version
        if ($global:LASTEXITCODE -ne $exitCode) { throw "Exit code not preserved: $exitCode" }
    }
} finally {
    Remove-Item Function:Get-Command
}

Write-Output "Validated $($topics.Count) help topics, $($aliases.Count) aliases, CLI flags, argument forwarding, stdout, and exit codes."