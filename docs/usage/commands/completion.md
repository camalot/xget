---
layout: default
title: 🖥️ Completions
nav_order: 1
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Shell completions

Generate an autocompletion script for xget and load it into Bash, Zsh, Fish, or
PowerShell.

```text
xget completion [command]
```

The `completion` command accepts `-h, --help` to show its help. Its shell
subcommands are `bash`, `fish`, `powershell`, and `zsh`. Each accepts
`-h, --help` and `--no-descriptions`; `--no-descriptions` omits descriptions
from completion suggestions. The global `--non-interactive` flag is also
available.

## Examples

Load completions into the current shell session:

```bash
# Bash
source <(xget completion bash)

# Zsh
source <(xget completion zsh)

# Fish
xget completion fish | source
```

For persistent completions, write the generated script to the appropriate
completion directory:

```bash
# Bash on Linux
xget completion bash > /etc/bash_completion.d/xget

# Bash on macOS
xget completion bash > "$(brew --prefix)/etc/bash_completion.d/xget"

# Zsh on Linux
xget completion zsh > "${fpath[1]}/_xget"

# Zsh on macOS
xget completion zsh > "$(brew --prefix)/share/zsh/site-functions/_xget"

# Fish
xget completion fish > ~/.config/fish/completions/xget.fish
```

For PowerShell, add the following command to your PowerShell profile:

```powershell
xget completion powershell | Out-String | Invoke-Expression
```

Bash completions require the `bash-completion` package. Zsh also requires
completion support to be enabled; add
`autoload -U compinit; compinit` to `~/.zshrc` if it is not already enabled.
Start a new shell after configuring persistent completions.