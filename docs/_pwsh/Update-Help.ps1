[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$man = Get-Content (Join-Path $PSScriptRoot '../_man/xget.md') -Raw
$sections = @{}
foreach ($match in [regex]::Matches($man, '(?ms)^# (?<title>[^\r\n]+)\r?\n(?<body>.*?)(?=^# |\z)')) {
    $sections[$match.Groups['title'].Value] = $match.Groups['body'].Value.Trim([char[]] "`r`n")
}

function ConvertFrom-ManMarkdown([string] $Text) {
    $Text = $Text -replace '(?m)^\s*```[^\r\n]*\r?$', ''
    $Text = $Text -replace '\[([^\]]+)\]\(([^)]+)\)', '$1 ($2)'
    $Text = $Text -replace '\*\*(.+?)\*\*', '$1' -replace '`', '' -replace '(?m)^## ', ''
    $Text = $Text -replace '(?m)^: +', '' -replace '(?m)^  ', ''
    $Text.Trim()
}

function Get-ManCommand([string] $Name, [string] $Section = 'COMMANDS') {
    $pattern = '(?ms)^  `xget ' + [regex]::Escape($Name) + '(?:[ `])[^\r\n]*\r?\n(?<body>.*?)(?=^  `xget |\z)'
    $match = [regex]::Match($sections[$Section], $pattern)
    if (-not $match.Success) { throw "Missing man page command: $Name" }
    ConvertFrom-ManMarkdown $match.Groups['body'].Value
}

$options = [ordered]@{}
foreach ($match in [regex]::Matches($sections['OPTIONS'], '(?ms)^  (?<names>`[^\r\n]+)\r?\n(?<body>.*?)(?=^  `[^\r\n]+\r?$|\z)')) {
    foreach ($flag in [regex]::Matches($match.Groups['names'].Value, '--(?<name>[a-z0-9_.-]+)')) {
        $options[$flag.Groups['name'].Value] = (ConvertFrom-ManMarkdown ($match.Groups['names'].Value + "`n" + $match.Groups['body'].Value))
    }
}
$options['installed'] = '--installed: Show stored installed-package metadata without checking for updates. Includes the last install or upgrade date and one row per install location.'
$options['check'] = '--check: Refresh the latest available versions before listing installed packages. Use with --installed.'
$options['no-color'] = '--no-color: Disable colored output; newer available versions would otherwise be yellow.'
$options['no-descriptions'] = '--no-descriptions: Generate completion candidates without descriptions. Available on completion shell subcommands.'

$installFlags = @('tag', 'pre-release', 'source', 'provider', 'to', 'binary', 'name', 'system', 'file', 'all', 'quiet', 'download-only', 'download-all', 'upgrade-only', 'untracked', 'asset', 'ignore', 'clobber_global.ignore', 'clobber_global.asset_filters', 'sha256', 'verify', 'verify-sha256', 'rate', 'remove', 'from', 'disable-ssl', 'config', 'version')
foreach ($flag in $installFlags + @('help', 'non-interactive')) {
    if (-not $options.Contains($flag)) { throw "Missing man page option: $flag" }
}

$topics = @(
    @{
        Name = 'Invoke-XGet'; Usage = 'xget [OPTIONS] TARGET | xget COMMAND [ARGS] [OPTIONS]'
        Synopsis = 'Install prebuilt binaries from GitHub or GitLab.'
        Description = (ConvertFrom-ManMarkdown $sections['DESCRIPTION']) + "`n`nCOMMANDS`n" + (ConvertFrom-ManMarkdown $sections['COMMANDS'])
        Flags = $installFlags
        Arguments = [ordered]@{ TARGET = 'Repository owner/repo, owner/repo@TAG, PROFILE:repository, a direct URL, or a local archive. GitLab supports nested namespaces. Optional when using --rate, --download-all, or a subcommand.'; COMMAND = 'install, uninstall (remove), upgrade (update), self-update, list, rate, config, version, completion, or help. Use Get-Help xget-COMMAND -Full for a command topic; config subcommands use xget-config-SUBCOMMAND.' }
        Example = 'xget install jgm/pandoc --to ~/.local/bin'
    },
    @{
        Name = 'Invoke-XGetInstall'; Usage = 'xget install TARGET [OPTIONS]'
        Synopsis = 'Download and install a prebuilt binary.'
        Description = (Get-ManCommand 'install') + "`n`n" + (ConvertFrom-ManMarkdown $sections['DESCRIPTION'])
        Flags = @($installFlags | Where-Object { $_ -ne 'version' })
        Arguments = [ordered]@{ TARGET = 'Required repository, repository@TAG, PROFILE:repository, direct URL, or local archive. A conflicting PROFILE: prefix and --provider is an error.' }
        Example = "xget install gerardog/gsudo --file '{{.Arch}}/*.exe' --all"
    },
    @{
        Name = 'Invoke-XGetUninstall'; Usage = 'xget uninstall PACKAGE [--from DIR | --all]'
        Synopsis = 'Remove installed files and tracking records.'
        Description = Get-ManCommand 'uninstall'
        Flags = @('from', 'all'); Overrides = @{ all = '--all (-a): Remove the package from every tracked install location. Otherwise select one of multiple locations with --from.' }
        Arguments = [ordered]@{ PACKAGE = 'Required package name, owner/repo, or tracking key (github:owner/repo). If untracked, its basename is removed from --from, XGET_BIN, or the current directory. remove is an alias for uninstall.' }
        Example = 'xget uninstall jgm/pandoc --from ~/.local/bin'
    },
    @{
        Name = 'Invoke-XGetUpgrade'; Usage = 'xget upgrade [PACKAGE[@TAG]] [--all] [--to DIR] [--tag TAG] [--no-color] [--config FILE]'
        Synopsis = 'List and apply upgrades for tracked packages.'
        Description = Get-ManCommand 'upgrade'
        Flags = @('all', 'to', 'tag', 'no-color', 'config')
        Overrides = @{ all = '--all (-a): Upgrade all out-of-date packages except pinned packages.'; to = '--to DIR: Only upgrade the copy installed to this tracked directory. Fail if no matching location exists.'; tag = '--tag TAG (-t): Install an exact release, allowing a downgrade or reinstall. Requires PACKAGE. PACKAGE@TAG is equivalent. An explicitly upgraded pinned package is re-pinned to the newer tag.' }
        Arguments = [ordered]@{ PACKAGE = 'Optional owner/repo, tracking key (github:owner/repo), or bare repository name. Without PACKAGE or --all, refresh metadata and list upgrades. update is an alias for upgrade.' }
        Example = "xget upgrade jgm/pandoc@3.10 --to '~/.local/bin'"
    },
    @{
        Name = 'Invoke-XGetSelfUpdate'; Usage = 'xget self-update [--to DIR] [--tag TAG] [--config FILE]'
        Synopsis = 'Update xget itself.'
        Description = 'Update tracked camalot/xget installations. If no tracked installation exists and --to is absent, update the running executable instead. The global xget_update_check setting controls update notices, not this command.'
        Flags = @('to', 'tag', 'config')
        Overrides = @{ to = '--to DIR: Update only the xget copy tracked in this location.'; tag = '--tag TAG (-t): Install this exact xget release instead of the latest; allows downgrades.' }
        Example = 'xget self-update'
    },
    @{
        Name = 'Invoke-XGetList'; Usage = 'xget list [TARGET] [--installed] [--check] [--pre-release] [--no-color] [--config FILE]'
        Synopsis = 'List releases, configured repositories, or installed packages.'
        Description = Get-ManCommand 'list'
        Flags = @('installed', 'check', 'pre-release', 'no-color', 'config')
        Arguments = [ordered]@{ TARGET = 'Optional release repository. Without TARGET, list configured repositories, or use --installed to list the tracking store.' }
        Example = 'xget list --installed --check --no-color'
    },
    @{
        Name = 'Invoke-XGetRate'; Usage = 'xget rate [--disable-ssl] [--config FILE]'
        Synopsis = 'Show GitHub API rate limiting information.'
        Description = Get-ManCommand 'rate'; Flags = @('disable-ssl', 'config')
        Example = 'xget rate'
    },
    @{
        Name = 'Invoke-XGetVersion'; Usage = 'xget version'
        Synopsis = 'Print the xget version, commit, and build date.'
        Description = Get-ManCommand 'version'; Flags = @()
        Example = 'xget version'
    },
    @{
        Name = 'Invoke-XGetConfig'; Usage = 'xget config SUBCOMMAND [ARGS] [--config FILE]'
        Synopsis = 'Read, write, and edit xget configuration.'
        Description = ConvertFrom-ManMarkdown $sections['CONFIGURATION']; Flags = @('config')
        Arguments = [ordered]@{ SUBCOMMAND = 'get, set, clear, pop, list, path, or edit. Use Get-Help xget-config-SUBCOMMAND -Full for individual help. Without a subcommand, show CLI help.' }
        Example = 'xget config set global target=~/.local/bin'
    },
    @{
        Name = 'Invoke-XGetCompletion'; Usage = 'xget completion SHELL [OPTIONS]'
        Synopsis = 'Generate a shell completion script.'
        Description = (Get-ManCommand 'completion') + "`n`nEach shell is a subcommand: completion bash, completion zsh, completion fish, and completion powershell. All accept --no-descriptions, --help, and --non-interactive. The generated script is written to stdout. Dot-source a saved PowerShell completion script to register native completion."
        Flags = @()
        Arguments = [ordered]@{ SHELL = 'Required completion subcommand: bash, zsh, fish, or powershell. Run xget completion SHELL --help for shell-specific loading instructions.' }
        Example = 'xget completion powershell | Out-String | Invoke-Expression'
    },
    @{
        Name = 'Invoke-XGetHelp'; Usage = 'xget help [COMMAND [SUBCOMMAND]]'
        Synopsis = 'Show the executable help for a command.'
        Description = 'Display Cobra CLI help, distinct from the PowerShell Get-Help topics. Use xget help config set for a nested command, or add --help to any command.'
        Flags = @(); Arguments = [ordered]@{ COMMAND = 'Optional command and subcommand path. With no arguments, show root CLI help.' }
        Example = 'xget help config set'
    }
)

foreach ($shell in @('bash', 'zsh', 'fish', 'powershell')) {
    $loading = @{
        bash = 'Requires the bash-completion package. In bash, run: source <(xget completion bash)'
        zsh = 'Enable compinit first (autoload -U compinit; compinit). In zsh, run: source <(xget completion zsh)'
        fish = 'In fish, run: xget completion fish | source. For persistent completion, save the output to ~/.config/fish/completions/xget.fish.'
        powershell = 'In PowerShell, run: xget completion powershell | Out-String | Invoke-Expression. Save the generated script and dot-source it from your PowerShell profile for persistent completion.'
    }
    $topics += @{
        Name = 'Invoke-XGetCompletion' + (Get-Culture).TextInfo.ToTitleCase($shell)
        Usage = "xget completion $shell [--no-descriptions]"
        Synopsis = "Generate $shell shell completions."
        Description = "Write the $shell completion script to stdout. " + $loading[$shell]
        Flags = @('no-descriptions'); Example = "xget completion $shell"
    }
}

foreach ($subcommand in @('get', 'set', 'clear', 'pop', 'list', 'path', 'edit')) {
    $arguments = [ordered]@{}
    $usage = "xget config $subcommand"
    if ($subcommand -in @('get', 'set', 'clear', 'pop')) {
        $arguments['SECTION'] = 'Required global, owner/repo, or sources.NAME section. Unknown settings are rejected. Source profiles support type, host, api_url, token_env, token, and disable_token_warning.'
        $usage += ' SECTION'
        if ($subcommand -in @('set', 'pop')) {
            $arguments['ASSIGNMENT'] = 'Required key=value argument. Quote the entire argument when it contains spaces or PowerShell metacharacters. Boolean values accept true, false, t, f, 1, and 0.'
            $usage += ' KEY=VALUE'
        } else {
            $arguments['KEY'] = 'Required configuration key, for example target, ignore, asset_filters, or clobber_global.ignore.'
            $usage += ' KEY'
        }
    }
    $examples = @{ get = 'global target'; set = 'global clobber_global.ignore=false'; clear = 'jgm/pandoc ignore'; pop = 'jgm/pandoc ignore=musl'; list = ''; path = ''; edit = '' }
    $topics += @{
        Name = 'Invoke-XGetConfig' + (Get-Culture).TextInfo.ToTitleCase($subcommand)
        Usage = "$usage [--config FILE]"
        Synopsis = "Configuration $subcommand subcommand."
        Description = (Get-ManCommand "config $subcommand" 'CONFIGURATION') + "`n`nOperate on one resolved TOML or YAML file. Use --config to select it explicitly. See Get-Help xget-config -Full for settings, search paths, inheritance, and source profiles."
        Flags = @('config'); Arguments = $arguments
        Example = ("xget config $subcommand " + $examples[$subcommand]).Trim()
    }
}

$namespaces = [ordered]@{
    maml = 'http://schemas.microsoft.com/maml/2004/10'
    command = 'http://schemas.microsoft.com/maml/dev/command/2004/10'
    dev = 'http://schemas.microsoft.com/maml/dev/2004/10'
}
$document = [System.Xml.XmlDocument]::new()
[void] $document.AppendChild($document.CreateXmlDeclaration('1.0', 'utf-8', $null))
$root = $document.CreateElement('helpItems')
$root.SetAttribute('schema', 'maml')
foreach ($namespace in $namespaces.GetEnumerator()) {
    $root.SetAttribute("xmlns:$($namespace.Key)", $namespace.Value)
}
[void] $document.AppendChild($root)

function Add-HelpElement($Parent, [string] $Name, [string] $Text) {
    $prefix, $localName = $Name.Split(':')
    $element = $document.CreateElement($prefix, $localName, $namespaces[$prefix])
    if ($Text) { $element.InnerText = $Text }
    [void] $Parent.AppendChild($element)
    $element
}

function Add-HelpDescription($Parent, [string] $Text) {
    $description = Add-HelpElement $Parent 'maml:description'
    foreach ($paragraph in [regex]::Split($Text.Trim(), '\r?\n\s*\r?\n')) {
        [void] (Add-HelpElement $description 'maml:para' $paragraph.Trim())
    }
}

foreach ($topic in $topics) {
    $command = Add-HelpElement $root 'command:command'
    $details = Add-HelpElement $command 'command:details'
    [void] (Add-HelpElement $details 'command:name' $topic.Name)
    Add-HelpDescription $details $topic.Synopsis
    $syntax = Add-HelpElement $command 'command:syntax'
    $item = Add-HelpElement $syntax 'command:syntaxItem'
    [void] (Add-HelpElement $item 'maml:name' $topic.Usage)
    Add-HelpDescription $command $topic.Description
    $parameters = Add-HelpElement $command 'command:parameters'
    $entries = [ordered]@{}
    if ($topic.Arguments) {
        foreach ($argument in $topic.Arguments.GetEnumerator()) { $entries[$argument.Key] = $argument.Value }
    }
    foreach ($flag in @($topic.Flags) + @('help', 'non-interactive')) {
        $text = $options[$flag]
        if ($topic.Overrides -and $topic.Overrides.ContainsKey($flag)) { $text = $topic.Overrides[$flag] }
        if (-not $text) { throw "Missing parameter documentation: $($topic.Name) $flag" }
        $entries[$flag] = $text
    }
    foreach ($entry in $entries.GetEnumerator()) {
        $parameter = Add-HelpElement $parameters 'command:parameter'
        $parameter.SetAttribute('required', ([string] ($entry.Value -match '^Required ')).ToLowerInvariant())
        $parameter.SetAttribute('position', 'named')
        if ($topic.Arguments -and $topic.Arguments.Contains($entry.Key)) {
            $parameter.SetAttribute('position', [string] @($topic.Arguments.Keys).IndexOf($entry.Key))
        }
        [void] (Add-HelpElement $parameter 'maml:name' $entry.Key)
        Add-HelpDescription $parameter $entry.Value
    }
    $examples = Add-HelpElement $command 'command:examples'
    $example = Add-HelpElement $examples 'command:example'
    [void] (Add-HelpElement $example 'maml:title' 'EXAMPLE 1')
    [void] (Add-HelpElement $example 'dev:code' $topic.Example)
    $alerts = Add-HelpElement $command 'maml:alertSet'
    $alert = Add-HelpElement $alerts 'maml:alert'
    [void] (Add-HelpElement $alert 'maml:para' 'This is native CLI syntax, not PowerShell parameter binding. Pass --long-options and -short-options exactly as shown. Get-Help parameter names omit the leading dashes. Arguments, stdout, and stderr are forwarded to xget on PATH. LASTEXITCODE is preserved: 0 success, 1 general error, 16 input required in non-interactive mode. Quote regex matchers beginning with ~ using single quotes.')
    if ($topic.Name -eq 'Invoke-XGet') {
        foreach ($section in @('ENVIRONMENT', 'EXIT STATUS', 'CONFIGURATION', 'FOR MAINTAINERS', 'AUTHOR')) {
            [void] (Add-HelpElement $alert 'maml:para' ("$section`n" + (ConvertFrom-ManMarkdown $sections[$section])))
        }
    }
    $links = Add-HelpElement $command 'maml:relatedLinks'
    foreach ($url in @('https://camalot.github.io/xget/', 'https://github.com/camalot/xget')) {
        $link = Add-HelpElement $links 'maml:navigationLink'
        [void] (Add-HelpElement $link 'maml:linkText' $(if ($url -match 'github.io') { 'Online manual' } else { 'Source and issues' }))
        [void] (Add-HelpElement $link 'maml:uri' $url)
    }
}

$settings = [System.Xml.XmlWriterSettings]::new()
$settings.Indent = $true
$settings.Encoding = [System.Text.UTF8Encoding]::new($false)
$settings.NewLineChars = "`n"
$writer = [System.Xml.XmlWriter]::Create((Join-Path $PSScriptRoot 'xget-help.xml'), $settings)
try { $document.Save($writer) } finally { $writer.Dispose() }
Write-Output "Generated $($topics.Count) help topics from docs/_man/xget.md."