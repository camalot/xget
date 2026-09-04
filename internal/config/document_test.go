package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFormatForPath(t *testing.T) {
	cases := map[string]struct {
		want    Format
		wantErr bool
	}{
		"a.toml": {want: FormatTOML},
		"a.TOML": {want: FormatTOML},
		"a.yml":  {want: FormatYAML},
		"a.yaml": {want: FormatYAML},
		"a.json": {wantErr: true},
		"a":      {wantErr: true},
	}
	for path, tc := range cases {
		got, err := FormatForPath(path)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("FormatForPath(%q) expected error", path)
			}
			continue
		}
		if err != nil {
			t.Fatalf("FormatForPath(%q): %v", path, err)
		}
		if got != tc.want {
			t.Fatalf("FormatForPath(%q) = %q, want %q", path, got, tc.want)
		}
	}
}

func TestValidateSection(t *testing.T) {
	valid := []string{"global", "zyedidia/micro", "camalot/xget"}
	for _, section := range valid {
		if err := ValidateSection(section); err != nil {
			t.Fatalf("ValidateSection(%q) = %v, want nil", section, err)
		}
	}
	invalid := []string{"", "micro", "a/b/c", "/repo", "owner/"}
	for _, section := range invalid {
		if err := ValidateSection(section); err == nil {
			t.Fatalf("ValidateSection(%q) expected error", section)
		}
	}
}

func TestParseAssignment(t *testing.T) {
	key, value, err := ParseAssignment("target=~/bin")
	if err != nil {
		t.Fatal(err)
	}
	if key != "target" || value != "~/bin" {
		t.Fatalf("got %q=%q", key, value)
	}

	if _, value, err = ParseAssignment("ignore="); err != nil {
		t.Fatal(err)
	} else if value != "" {
		t.Fatalf("expected empty value, got %q", value)
	}

	if _, _, err := ParseAssignment("target"); err == nil {
		t.Fatal("expected error for missing =")
	}
	if _, _, err := ParseAssignment("=value"); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestSectionKeysCoverGlobalAndRepository(t *testing.T) {
	globals := SectionKeys(GlobalSection)
	if globals["quiet"] != KindBool {
		t.Fatalf("global quiet kind = %v", globals["quiet"])
	}
	if globals["target"] != KindString {
		t.Fatalf("global target kind = %v", globals["target"])
	}
	if globals["ignore"] != KindStringSlice {
		t.Fatalf("global ignore kind = %v", globals["ignore"])
	}
	if _, ok := globals["asset_filters"]; ok {
		t.Fatal("asset_filters must not be a global key")
	}

	repos := SectionKeys("owner/repo")
	if repos["asset_filters"] != KindStringSlice {
		t.Fatalf("repo asset_filters kind = %v", repos["asset_filters"])
	}
	if _, ok := repos["name"]; ok {
		t.Fatal("name must not be settable")
	}
	if _, ok := repos["github_token"]; ok {
		t.Fatal("github_token must not be a repository key")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDocumentMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", ".xget.yml")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Existed {
		t.Fatal("expected Existed to be false")
	}
	if len(doc.Data) != 0 {
		t.Fatalf("expected empty data, got %v", doc.Data)
	}
	if doc.Format != FormatYAML {
		t.Fatalf("format = %q", doc.Format)
	}
}

func TestLoadDocumentDirectoryAndBadExtension(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadDocument(dir); err == nil {
		t.Fatal("expected error for unsupported extension on directory path")
	}

	yamlDir := filepath.Join(dir, "conf.yml")
	if err := os.MkdirAll(yamlDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDocument(yamlDir); !errors.Is(err, errNotAFile) {
		t.Fatalf("expected errNotAFile, got %v", err)
	}
}

func TestLoadDocumentEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.toml")
	writeFile(t, path, "   \n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Existed {
		t.Fatal("expected Existed to be true")
	}
	if len(doc.Data) != 0 {
		t.Fatalf("expected empty data, got %v", doc.Data)
	}
}

func TestLoadDocumentInvalidContent(t *testing.T) {
	tomlPath := filepath.Join(t.TempDir(), ".xget.toml")
	writeFile(t, tomlPath, "[global\nquiet = true\n")
	if _, err := LoadDocument(tomlPath); err == nil {
		t.Fatal("expected toml parse error")
	}

	yamlPath := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, yamlPath, "global:\n  quiet: true\n bad\n")
	if _, err := LoadDocument(yamlPath); err == nil {
		t.Fatal("expected yaml parse error")
	}
}

func TestSetGetClearRoundTripYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set(GlobalSection, "quiet", "true"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Set(GlobalSection, "target", "~/bin"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("zyedidia/micro", "asset_filters", "static"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("zyedidia/micro", "asset_filters", ".tar.gz"); err != nil {
		t.Fatal(err)
	}
	// Duplicate append is a no-op.
	if err := doc.Set("zyedidia/micro", "asset_filters", "static"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	values, err := reloaded.Get("zyedidia/micro", "asset_filters")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []string{"static", ".tar.gz"}) {
		t.Fatalf("asset_filters = %v", values)
	}
	if values, err = reloaded.Get(GlobalSection, "quiet"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(values, []string{"true"}) {
		t.Fatalf("quiet = %v", values)
	}

	// The saved file must still parse through the normal loader.
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Global.Quiet || cfg.Global.Target != "~/bin" {
		t.Fatalf("loaded config = %+v", cfg.Global)
	}
	if got := cfg.Repositories["zyedidia/micro"].AssetFilters; !reflect.DeepEqual(got, []string{"static", ".tar.gz"}) {
		t.Fatalf("repo asset_filters = %v", got)
	}

	if err := reloaded.Clear(GlobalSection, "target"); err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.Get(GlobalSection, "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("expected ErrKeyNotSet, got %v", err)
	}
}

func TestSetPreservesTOMLFormatAndUnknownSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.toml")
	writeFile(t, path, "[global]\nquiet = true\n\n[\"owner/repo\"]\ntarget = \"~/bin\"\n")

	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != FormatTOML {
		t.Fatalf("format = %q", doc.Format)
	}
	if err := doc.Set(GlobalSection, "show_hash", "true"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 || raw[0] != '[' {
		t.Fatalf("expected TOML output, got:\n%s", raw)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Global.ShowHash || !cfg.Global.Quiet {
		t.Fatalf("global = %+v", cfg.Global)
	}
	if cfg.Repositories["owner/repo"].Target != "~/bin" {
		t.Fatalf("repo target = %q", cfg.Repositories["owner/repo"].Target)
	}
}

func TestSetRejectsUnknownKeyAndBadBool(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set(GlobalSection, "nope", "1"); err == nil {
		t.Fatal("expected unknown key error")
	}
	if err := doc.Set(GlobalSection, "quiet", "yesplease"); err == nil {
		t.Fatal("expected bool parse error")
	}
	if err := doc.Set("bad-section", "quiet", "true"); err == nil {
		t.Fatal("expected section error")
	}
	if err := doc.Set(GlobalSection, "asset_filters", "x"); err == nil {
		t.Fatal("asset_filters is not a global key")
	}
}

func TestGetErrorsForUnsetSectionKeyAndInvalidInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, path, "global:\n  ignore: []\n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := doc.Get("owner/repo", "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing section: %v", err)
	}
	if _, err := doc.Get(GlobalSection, "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing key: %v", err)
	}
	if _, err := doc.Get(GlobalSection, "ignore"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("empty list should be unset: %v", err)
	}
	if _, err := doc.Get(GlobalSection, "bogus"); err == nil {
		t.Fatal("expected unknown key error")
	}
	if _, err := doc.Get("bad", "quiet"); err == nil {
		t.Fatal("expected section error")
	}
}

func TestClearErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, path, "global:\n  quiet: true\n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Clear("owner/repo", "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing section: %v", err)
	}
	if err := doc.Clear(GlobalSection, "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing key: %v", err)
	}
	if err := doc.Clear(GlobalSection, "bogus"); err == nil {
		t.Fatal("expected unknown key error")
	}
	if err := doc.Clear("bad", "quiet"); err == nil {
		t.Fatal("expected section error")
	}
	if err := doc.Clear(GlobalSection, "quiet"); err != nil {
		t.Fatal(err)
	}
}

func TestPopListScalarAndErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, path, "global:\n  ignore:\n    - a\n    - b\n  target: \"~/bin\"\n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := doc.Pop(GlobalSection, "ignore", "a"); err != nil {
		t.Fatal(err)
	}
	values, err := doc.Get(GlobalSection, "ignore")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []string{"b"}) {
		t.Fatalf("ignore = %v", values)
	}

	// Removing the last list entry removes the key.
	if err := doc.Pop(GlobalSection, "ignore", "b"); err != nil {
		t.Fatal(err)
	}
	if _, err := doc.Get(GlobalSection, "ignore"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("expected key removed, got %v", err)
	}

	// Scalar mismatch is a no-op error; matching value clears the key.
	if err := doc.Pop(GlobalSection, "target", "/other"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("scalar mismatch: %v", err)
	}
	if err := doc.Pop(GlobalSection, "target", "~/bin"); err != nil {
		t.Fatal(err)
	}
	if _, err := doc.Get(GlobalSection, "target"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("expected target cleared, got %v", err)
	}

	if err := doc.Pop("owner/repo", "target", "x"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing section: %v", err)
	}
	if err := doc.Pop(GlobalSection, "quiet", "true"); !errors.Is(err, ErrKeyNotSet) {
		t.Fatalf("missing key: %v", err)
	}
	if err := doc.Pop(GlobalSection, "bogus", "x"); err == nil {
		t.Fatal("expected unknown key error")
	}
	if err := doc.Pop("bad", "quiet", "true"); err == nil {
		t.Fatal("expected section error")
	}
}

func TestEntriesSortsGlobalFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, path, "\"zyedidia/micro\":\n  target: \"~/m\"\n  asset_filters: [static, tar]\n\"aaa/bbb\":\n  quiet: true\nglobal:\n  target: \"~/bin\"\n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	got := doc.Entries()
	want := []Entry{
		{Section: "global", Key: "target", Value: "~/bin"},
		{Section: "aaa/bbb", Key: "quiet", Value: "true"},
		{Section: "zyedidia/micro", Key: "asset_filters", Value: "static"},
		{Section: "zyedidia/micro", Key: "asset_filters", Value: "tar"},
		{Section: "zyedidia/micro", Key: "target", Value: "~/m"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Entries() = %v, want %v", got, want)
	}
}

func TestEntriesHandlesNonSectionValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	writeFile(t, path, "stray: 5\nglobal:\n  quiet: true\n")
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := doc.Entries()
	found := false
	for _, entry := range entries {
		if entry.Section == "stray" && entry.Key == "" && entry.Value == "5" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected stray entry, got %v", entries)
	}
}

func TestFormatValues(t *testing.T) {
	if got := FormatValues(nil); got != nil {
		t.Fatalf("nil = %v", got)
	}
	if got := FormatValues(true); !reflect.DeepEqual(got, []string{"true"}) {
		t.Fatalf("bool = %v", got)
	}
	if got := FormatValues(42); !reflect.DeepEqual(got, []string{"42"}) {
		t.Fatalf("int = %v", got)
	}
	if got := FormatValues([]any{"a", 1}); !reflect.DeepEqual(got, []string{"a", "1"}) {
		t.Fatalf("slice = %v", got)
	}
	if got := FormatValues([]string{"a"}); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("string slice = %v", got)
	}
}

func TestSaveUnsupportedFormat(t *testing.T) {
	doc := &Document{Path: filepath.Join(t.TempDir(), "x.json"), Format: Format("json"), Data: map[string]any{}}
	if err := doc.Save(); err == nil {
		t.Fatal("expected unsupported format error")
	}
}

func TestDefaultWritePathUsesXDGThenHome(t *testing.T) {
	home := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	want := filepath.Join(home, "xdg", "xget", ".xget.yml")
	if got := DefaultWritePath(home); got != want {
		t.Fatalf("DefaultWritePath() = %q, want %q", got, want)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	want = filepath.Join(home, ".config", "xget", ".xget.yml")
	if got := DefaultWritePath(home); got != want {
		t.Fatalf("DefaultWritePath() = %q, want %q", got, want)
	}
}

func TestResolvePathPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Chdir(t.TempDir())

	// Explicit path wins.
	if got, err := ResolvePath("/explicit/.xget.toml"); err != nil || got != "/explicit/.xget.toml" {
		t.Fatalf("explicit = %q, %v", got, err)
	}

	// Nothing exists: fall back to the XDG default.
	want := filepath.Join(home, ".config", "xget", ".xget.yml")
	got, err := ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("fallback = %q, want %q", got, want)
	}

	// An existing candidate wins over the fallback.
	existing := filepath.Join(home, ".xget.toml")
	writeFile(t, existing, "[global]\n")
	if got, err = ResolvePath(""); err != nil || got != existing {
		t.Fatalf("candidate = %q, %v (want %q)", got, err, existing)
	}

	// XGET_CONFIG beats discovery.
	t.Setenv("XGET_CONFIG", "/env/.xget.yml")
	if got, err = ResolvePath(""); err != nil || got != "/env/.xget.yml" {
		t.Fatalf("env = %q, %v", got, err)
	}
}

func TestResolvePathIgnoresCandidateDirectories(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(filepath.Join(home, ".xget.toml"), 0o750); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".config", "xget", ".xget.yml")
	got, err := ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ResolvePath() = %q, want %q", got, want)
	}
}
