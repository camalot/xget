# PowerShell Help

Dot-source the shim to enable native PowerShell help. Keep the shim and its
generated MAML help file together. The xget executable must already be on `PATH`
to run commands, but viewing help does not require it.

```powershell
. ./docs/_pwsh/xget.ps1

Get-Help xget -Full
Get-Help xget-install -Examples
Get-Help xget-install -Parameter asset
Get-Help xget-upgrade -Full
Get-Help xget-config-set -Full
Get-Help xget-completion-powershell -Full
```

For persistent help, dot-source the shim using its absolute path from `$PROFILE`.
Loading it does not download anything, install xget, or change execution policy.
Run it again to reload the functions and aliases.

## Topics

Use `Get-Help xget-*` to discover the command topics:

- `xget-install`
- `xget-uninstall` (`xget-remove`)
- `xget-upgrade` (`xget-update`)
- `xget-self-update`
- `xget-list`
- `xget-rate`
- `xget-version`
- `xget-config`
- `xget-config-get`, `xget-config-set`, `xget-config-clear`, `xget-config-pop`
- `xget-config-list`, `xget-config-path`, `xget-config-edit`
- `xget-completion`
- `xget-completion-bash`, `xget-completion-zsh`, `xget-completion-fish`,
  `xget-completion-powershell`
- `xget-help`

The root topic includes the man page's configuration, environment, exit status,
and maintainer reference. Parameter help uses names without leading dashes, for
example `Get-Help xget-install -Parameter clobber_global.ignore`.

## Running Commands

The shim preserves native CLI syntax. It does not introduce PowerShell-style
parameter binding. Commands, arguments, stdout, stderr, and `$LASTEXITCODE` are
forwarded to the executable. Help-topic aliases also run their respective
commands:

```powershell
xget install jgm/pandoc --to ~/.local/bin
xget-config-get global target
xget list --installed --check
```

Use `xget help config set` or `xget config set --help` for the executable's own
help. Call `xget.exe` explicitly to bypass the shim on Windows.

## Maintaining Help

The Markdown man page is the source for shared descriptions, options, and
reference material. Command mappings and supplemental help for commands absent
from the man page live in the generator. After changing either, regenerate and
test the checked-in help file:

```powershell
& ./docs/_pwsh/Update-Help.ps1
pwsh -NoProfile -File ./docs/_pwsh/Test-Help.ps1
```

The test requires no external modules or network access. It checks help topics,
parameter lookup, aliases, CLI flag coverage, literal regex/glob preservation,
and forwarding with a mocked application. Use a fresh PowerShell process because
`Get-Help` caches external help during a session.
