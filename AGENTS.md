# AGENTS.md

Guidance for AI coding agents working on **xget**, a Go CLI that downloads and installs
pre-built binaries from GitHub/GitLab releases (or direct URLs / local archives). It is a
fork/rewrite of [eget](https://github.com/zyedidia/eget) (see `eget.LICENSE`).

## Quick commands

| Task | Command |
| --- | --- |
| Build | `go build ./...` |
| Unit tests | `go test ./...` |
| Single package | `go test ./internal/cli/...` |
| Vet | `go vet ./...` |
| Format check | `gofmt -l .` (must print nothing) |
| Lint (same version as CI) | `task lint` or `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...` |
| Full local build (lint, sha256, goreleaser snapshot, vet, test) | `task build` |
| Release dry run (renders scoop/winget/homebrew manifests into `dist/`) | `task release:dry-run` |

- Go version comes from `go.mod` (`go 1.27.0`). Use `go mod tidy` after changing imports.
- `task build` requires `goreleaser` and `bash`. On Windows, make sure `bash` is Git for
  Windows/MSYS2 bash, not `C:\Windows\System32\bash.exe` (the WSL stub).
- Before finishing, run at least `go build ./...`, `go vet ./...`, `gofmt -l .`, and
  `go test ./...`.

## Layout

```text
cmd/xget/main.go        Entry point: removes stale self-update exe, loads .env files, runs cli.Execute(), maps errors to exit codes
internal/cli/           Cobra commands and flags (root/install, list, config, upgrade, uninstall, self-update, version, rate)
internal/engine/        Core logic: target parsing, finders (GitHub/GitLab/direct), asset detection, download, verify, extract, self-update
internal/config/        Config file discovery/merge (TOML/YAML), comment-preserving config documents, dotenv loading
internal/installed/     Tracking store for installed packages (.xget.installed.yml)
internal/options/       Shared options struct passed from CLI to engine
internal/semver/        Version comparison helpers
internal/home/          Home directory / ~ expansion
internal/lib/constants/ Shared constants
test/                   End-to-end smoke test harness (runs a built binary via TEST_XGET; hits the network)
install/                xget.sh / xget.ps1 installer scripts and their .sha256 files
scripts/                generate-install-sha256.sh (regenerates install/*.sha256)
docs/                   Jekyll documentation site (published via GitHub Pages)
docs/_man/xget.md       Source for the man page (pandoc -> xget.1)
docs/_plans/            Design notes for in-progress/planned features
winres/                 Windows resource metadata (icon/version info, via go-winres tool)
.github/actions/        Composite actions used by release/publish workflows (packaging, verification)
```

## Conventions

- **Style**: `gofmt`/`goimports` formatting; golangci-lint config in `.golangci.yml`
  (staticcheck, errcheck, gosec, revive, gocritic, gocyclo, goconst, ...). Fix lint issues
  rather than suppressing. When a `gosec` finding is a deliberate false positive, use a
  targeted `// #nosec GXXX -- reason` comment, matching the existing code.
- **Output**: user-facing prompts, progress, and diagnostics go to **stderr**; stdout is
  reserved for command output (e.g. `config get`, `list`) so it can be piped.
- **Errors**: return errors up to `cli.Execute`; `main` prints them and exits with
  `cli.ExitCodeFor(err)`. Exit codes: `0` success, `1` general error, `16`
  `engine.ErrNonInteractive` (input required while non-interactive).
- **Non-interactive mode**: `engine.NonInteractive()` is true when `--non-interactive` is
  set **or** stdin is not a TTY (`stdinIsTerminal` in `internal/cli/root.go`). Any new
  prompt must check it and return `engine.ErrNonInteractive` instead of reading stdin.
- **Testability seams**: side effects are exposed as package-level function variables
  (e.g. `runEditor`, `getRateLimit`, `stdinIsTerminal`) that tests swap and restore with
  `defer`. Follow the same pattern instead of adding interfaces for single uses.
- **CLI tests** use `runCLI(t, args...)` in `internal/cli/config_test.go`, which builds a
  fresh root command, captures output, and simulates an interactive TTY.
- **Unit tests must not hit the network.** Use `httptest` servers (see
  `engine/source_http_test.go`, `engine/gitlab_test.go`). Network tests belong in `test/`.
- **Cross-platform**: CI runs tests on both `windows-latest` and `ubuntu-latest`. Use
  `filepath` (not `path`) for filesystem paths, and guard OS-specific behavior with
  `runtime.GOOS` or build tags.
- **Backward compatibility with eget**: `EGET_*` env vars and `.eget.toml` are still
  honored alongside `XGET_*` / `.xget.{toml,yaml,yml}`. Don't remove the eget fallbacks.
- **Config precedence**: CLI flags > repository-specific config > global config. Config
  files are merged from lowest to highest priority; see
  `docs/configuration/loading-and-precedence.md`.
- **GitHub token lookup**: `EGET_GITHUB_TOKEN` > `XGET_GITHUB_TOKEN` > `GITHUB_TOKEN`
  (a value of `@path` reads the token from a file). Never log token values.

## Documentation

When changing flags, behavior, config keys, or exit codes, update the docs that cover them:

- `README.md`
- `docs/usage/index.md` (includes the `--help` output and the exit code table)
- `docs/_man/xget.md` (man page)
- `docs/features.md`
- `docs/configuration/*.md` for config changes

Markdown is linted with `.markdownlint.jsonc`.

## Commits, changelog, releases

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`,
  `docs:`, `chore:`, `refactor:`, `test:`, with an optional scope such as `fix(config):`).
  `CHANGELOG.md` is generated from commits by git-cliff (`cliff.toml`), so don't edit
  released changelog sections by hand.
- Releases are cut manually via the `release.yml` workflow (`workflow_dispatch`) using
  GoReleaser (`.goreleaser.yaml`). Publish workflows push packages to Scoop, WinGet,
  Homebrew, Chocolatey, AUR, Alpine, Debian, and RPM.
- If you edit `install/xget.sh` or `install/xget.ps1`, regenerate the checksums with
  `bash ./scripts/generate-install-sha256.sh`. CI's `verify-sha256` job fails otherwise.

## Related repositories

- **xget-action** (`camalot/xget-action`): GitHub Action that installs xget and runs it with
  `--non-interactive --untracked --verify`. Renaming or removing CLI flags breaks it, so
  keep flags backward compatible or update the action as well.
- **scoop** (`camalot/scoop`): Scoop bucket / Homebrew tap / WinGet and Chocolatey manifests
  that the release pipeline updates.
- **actions** (`camalot/actions`): Shared reusable actions (lint, coverage, versioning) used
  by the CI workflows.
