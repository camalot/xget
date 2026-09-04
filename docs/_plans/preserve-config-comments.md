# Plan: preserve comments and formatting in `xget config` writes

Status: **proposed** (not implemented)

## Current behavior

`internal/config/document.go` loads a config file into a `map[string]any`,
mutates it, and re-marshals the whole map with `pelletier/go-toml/v2` or
`go.yaml.in/yaml/v3`. This is simple and correct for values, but the round trip
discards everything that is not data:

- comments (leading, trailing, and inline)
- blank lines and section grouping
- original key ordering (both encoders emit map keys alphabetically)
- quoting style (`'single'` vs `"double"`), literal/folded YAML scalars
- TOML inline tables and array-of-tables layout
- YAML anchors, aliases, and merge keys

## Goal

`xget config set|clear|pop` should produce a minimal diff: only the touched key
changes; everything else in the file — including comments — is byte-identical.

## Approach A: node-tree editing (recommended)

Keep the file as a parsed syntax tree instead of a plain map, mutate the node for
the target key, and re-serialize the tree.

### YAML

`go.yaml.in/yaml/v3` already supports this via `yaml.Node`:

- Decode into a `*yaml.Node` (`yaml.Unmarshal(raw, &root)`).
- `yaml.Node` retains `HeadComment`, `LineComment`, `FootComment`, `Style`, and
  key order (mapping nodes are flat `[k1, v1, k2, v2, ...]` slices).
- Mutating means walking `root.Content[0]` to find the section mapping, then the
  key scalar, then replacing/inserting/removing the value node.
- Re-encode with `yaml.Encoder` and `SetIndent(2)`.

Known gaps:

- Blank-line preservation is imperfect. yaml.v3 attaches blank lines to comments
  via `HeadComment` only when a comment is present; standalone blank lines
  between keys are lost. Mitigation: post-process, or accept it.
- Newly inserted keys need a `Style` chosen deliberately (plain vs quoted) so
  values like `~/bin` and `~\.sbom\.json$` stay valid.
- Anchors/aliases: mutating an aliased node changes every use site. Detect
  `node.Kind == yaml.AliasNode` on the write path and return an actionable error
  rather than silently rewriting.

Estimated effort: moderate. The `yaml.Node` API is stable and documented.

### TOML

`pelletier/go-toml/v2` intentionally dropped the v1 document/AST editing API; it
is marshal/unmarshal only. Options:

1. **`go-toml/v2` + `unstable.Parser`** — the `unstable` package exposes an
   expression-level parser with byte offsets into the source. We can locate the
   exact byte range of a `key = value` expression inside a table and splice the
   new text in, leaving the rest of the file untouched. Insertions require
   choosing an anchor (end of the target table) and removals require deleting the
   line plus its trailing newline. This gives true byte-level preservation but
   the API is explicitly marked unstable and may change between minor versions.
2. **Pin `pelletier/go-toml` v1** — v1 has `toml.Tree` with `SetWithComment` and
   preserves ordering and comments reasonably well. It is archived/unmaintained
   and would add a second TOML dependency alongside the v2 one viper already
   pulls in.
3. **Hand-rolled line editor** — treat the file as lines, track the current
   `[table]` header, and rewrite/insert/delete the matching `key =` line. Handles
   the ~95% case (simple flat key/value files, which is what xget configs are)
   with no new dependencies. Breaks on multi-line arrays, inline tables, and
   multi-line basic strings, which must be detected and rejected.

Recommendation: option 1 (`unstable.Parser` + byte splicing) with option 3 as a
fallback if the unstable API churns. Both need a "cannot safely edit this
construct" escape hatch.

Estimated effort: significant for TOML — this is the bulk of the work.

## Approach B: rewrite + comment reattachment

Marshal as today, then diff the old and new files and re-insert the comment lines
that were anchored to surviving keys. Fragile: anchoring comments to keys after
reordering is heuristic, and it fails as soon as a key is renamed or removed.
Not recommended.

## Required work (approach A)

1. Introduce an internal `docstore` interface behind `config.Document`:

   ```go
   type store interface {
       Get(section, key string) (any, bool)
       Set(section, key string, value any) error
       Delete(section, key string) error
       Sections() []string
       Keys(section string) []string
       Bytes() ([]byte, error)
   }
   ```

   The existing map-based implementation becomes `mapStore` and remains the
   fallback.
2. Implement `yamlNodeStore` over `*yaml.Node`.
3. Implement `tomlSpliceStore` over `unstable.Parser` byte ranges.
4. Add a `--rewrite` escape hatch (or `XGET_CONFIG_REWRITE=1`) that forces the
   map-based path when the preserving store reports an unsupported construct.
5. Golden-file tests: for each of TOML and YAML, a fixture with header comments,
   inline comments, blank lines, non-alphabetical key order, and a multi-line
   array. Assert that `set`/`clear`/`pop` change exactly the expected lines.
6. Fuzz/round-trip test: parse → no-op save → assert byte equality for a corpus
   of config files.
7. Document the guarantees (and the known gaps) in
   `docs/configuration/config.md`, replacing the current "comments are not
   preserved" note.

## Risks

- `go-toml/v2`'s `unstable` package can break on dependency upgrades; the
  fallback path and round-trip tests are what make this tolerable.
- Byte-splicing must be careful about CRLF line endings on Windows.
- Preserving comments means preserving *stale* comments: a comment describing a
  value that `xget config set` just changed will now be wrong. Acceptable, and
  the same as `git config` behavior.

## Decision

Ship the rewrite-based implementation now. Revisit this plan afterward; the
`store` interface in step 1 is the only change that would need to land in the
current code to keep the door open, and it can be introduced later without
changing the CLI surface.
