# Plan: configurable GitHub and GitLab release sources

Status: **implemented**

## Goal

Allow `xget install` to download release assets from GitLab as well as GitHub,
while preserving the existing source-archive behavior and existing GitHub
configuration.

## Decisions

- Keep the existing boolean `--source` flag unchanged for downloading repository
  source archives.
- Add `--provider PROFILE` to select a release-source profile.
- Use named source profiles so users can configure multiple accounts, tokens, and
  hosts for the same provider type.
- Resolve tokens from configured environment variable names first, then from the
  profile's `token` setting.
- Keep `global.github_token` as the final fallback for GitHub profiles.
- A normal GitLab release install considers only release asset links. GitLab's
  generated source archives are used only when `--source` is set.
- Support nested GitLab namespaces such as `group/subgroup/project`.

## Configuration model

`global.source` and repository-level `source` select a named profile. The CLI
`--provider` value has the highest precedence. If none is set, the built-in
`github` profile is used.

```toml
[global]
source = "github"

[sources.github]
type = "github"
token_env = ["XGET_GITHUB_TOKEN", "GITHUB_TOKEN", "EGET_GITHUB_TOKEN"]
disable_token_warning = false

[sources.gitlab]
type = "gitlab"
token_env = ["XGET_GITLAB_TOKEN", "GITLAB_TOKEN"]

[sources.work]
type = "github"
host = "github.example.com"
api_url = "https://github.example.com/api/v3"
token_env = ["WORK_GITHUB_TOKEN"]
```

Each source profile supports:

| Setting | Meaning |
| --- | --- |
| `type` | Provider implementation: `github` or `gitlab`. Defaults to the profile name for the built-in profiles. |
| `host` | Repository host. Defaults to `github.com` or `gitlab.com`. |
| `api_url` | API base URL. Defaults to the public API, or the conventional enterprise API path for a custom host. |
| `token_env` | Environment variables checked in order. Values may use the existing `@/path/to/file` syntax. |
| `token` | Plaintext token fallback after the configured environment variables. |
| `disable_token_warning` | Suppress the warning for a token stored in this profile. |

The built-in profiles supply the defaults shown above even when no `sources`
section exists. A configured profile with the same name overrides its built-in
defaults field by field.

For compatibility, `global.github_token` remains accepted as the final token
fallback for every GitHub profile. Its warning will explain that it can be
disabled with `sources.github.disable_token_warning = true`.

`global.disable_token_warning = true` suppresses all plaintext-token warnings;
the source-level setting suppresses only that profile. A `token` or
`global.github_token` value beginning with `@` is a token-file reference, so its
contents are used as the token and no plaintext-storage warning is emitted.

## Implementation

1. Extend `internal/config` with source-profile parsing, built-in defaults,
   validation, profile resolution, and warning control. Exclude the reserved
   `sources` table from repository parsing.
2. Replace process-wide token environment mutation with a resolved source value
   passed through `options.Flags`. Keep compatibility helpers used by the
   GitHub-only `rate` command until that command is generalized separately.
3. Add `--provider` to the root and `install` commands and apply precedence as
   CLI, repository, global, then built-in `github`.
4. Refactor finder construction around the resolved provider. Preserve direct
   URL/local-file behavior and the existing GitHub behavior. Recognize full
   repository URLs only when their host matches the selected profile.
5. Implement GitLab release lookup through API v4. URL-escape the complete
  namespace/project identifier, select latest or requested tags, filter
  upcoming releases, and return `assets.links` URLs. GitLab does not expose a
  prerelease classification, so `--pre-release` has no separate GitLab effect.
6. Implement GitHub and GitLab source-archive finders using each selected
   profile's API/host settings.
7. Make authentication request-scoped: GitHub uses `Authorization: Bearer`, and
   GitLab uses `PRIVATE-TOKEN`. Never attach a token to a host other than the
   profile's configured API host, and retain the existing SSL-disable safeguard.
8. Persist the selected profile name as installed-package source metadata so
   upgrades reuse it. Continue reading legacy `GitHub` and `URL` values.
9. Update README, `docs/configuration`, `docs/usage`, generated man-page source,
   and command examples for GitLab, `--provider`, profiles, token precedence,
   compatibility, and warning suppression.

## Validation

- Config tests for built-ins, named profile overrides, reserved `sources`,
  source precedence, legacy `github_token`, token environment ordering, token
  file values, warning output, and warning suppression in TOML and YAML.
- CLI tests proving `--source` remains boolean and `--provider` overrides
  repository/global source selection.
- Finder tests with `httptest.Server` for GitHub and GitLab latest/tagged
  releases, nested GitLab namespaces, pagination/fallback behavior, prerelease
  handling, source archives, custom API URLs, authentication headers, errors,
  and no token leakage to unrelated hosts.
- Upgrade/install-store compatibility tests for old and new source values.
- Run focused package tests after each slice, then `go test ./...`, formatting,
  lint/build tasks available in `TaskFile.yml`, and documentation generation or
  consistency checks.

## Risks

- GitHub and GitLab model "latest" and prereleases differently. Tests must lock
  down xget's cross-provider semantics rather than relying on API ordering.
- A profile name is persisted instead of only its provider type; deleting or
  renaming that profile can make a later upgrade fail. The resulting error must
  name the missing profile and explain how to restore or override it.
- Custom hosts make accidental credential disclosure possible. Authentication
  must be bound to the parsed API origin, including scheme and host.

## Critique log

An independent implementation critique was completed after the first full test
pass.

- **Accepted:** `upgradeNamed` still allowed only the literal legacy `GitHub`
  source. It now rejects only direct `URL` installs, matching bulk upgrade
  behavior, and tests cover both built-in GitLab and custom named profiles.
- **Accepted:** add explicit GitLab/custom-profile upgrade coverage so persisted
  profile names are verified through refresh and install option resolution.
- **Additional security correction:** the critique confirmed direct request host
  checks, but Go can copy nonstandard headers such as GitLab's `PRIVATE-TOKEN`
  while following redirects. A redirect policy now strips both provider auth
  headers whenever the destination is not one of the profile's HTTPS hosts.
- **Additional config correction:** the editable nested profile path worked for
  get/set/clear/pop, but `xget config list` initially treated `sources` as an
  opaque map. It now emits sorted `sources.NAME.key=value` entries and has a
  regression test.
- **Additional compatibility correction:** profile-aware GitHub source archives
  initially used an API URL without a filename extension, which broke archive
  type detection. The established `.tar.gz` URL shape is retained and now uses
  the configured GitHub host.
- **Token warning refinement:** global warning suppression now applies to every
  configured source, while profile suppression remains scoped. Token-file
  references in either `global.github_token` or a profile `token` are resolved
  by the existing token loader and do not trigger plaintext warnings.
- **Provider shorthand:** `PROFILE:repository` is parsed before repository
  configuration lookup and behaves like `--provider PROFILE`. Conflicting
  shorthand and flag values are rejected. Tracked installs derive their source
  from the selected finder profile and store the canonical repository name.
- **Not implemented as provider inference:** full GitLab URLs still require the
  GitLab profile to be selected. Automatic inference would contradict the
  agreed default/profile precedence and could bypass credentials selected by
  the user.
- **Not implemented as synthetic prerelease detection:** GitLab has no
  prerelease field. Guessing from tag text would be unreliable, so the docs now
  state that `--pre-release` has no distinct GitLab behavior while upcoming
  releases remain excluded.