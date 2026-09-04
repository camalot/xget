package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestRootCommandIncludesConfigSubcommand(t *testing.T) {
	cmd := newRootCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "config" {
			names := map[string]bool{}
			for _, child := range sub.Commands() {
				names[child.Name()] = true
			}
			for _, want := range []string{"get", "set", "clear", "pop", "list", "path", "edit"} {
				if !names[want] {
					t.Fatalf("config subcommand %q missing", want)
				}
			}
			return
		}
	}
	t.Fatal("config subcommand not registered")
}

func TestConfigSetGetClearYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")

	if _, err := runCLI(t, "config", "set", "global", "target=~/bin", "--config", path); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "config", "set", "zyedidia/micro", "asset_filters=static", "--config", path); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "config", "set", "zyedidia/micro", "asset_filters=.tar.gz", "--config", path); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "config", "get", "zyedidia/micro", "asset_filters", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if out != "static\n.tar.gz\n" {
		t.Fatalf("get output = %q", out)
	}

	out, err = runCLI(t, "config", "get", "global", "target", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "~/bin" {
		t.Fatalf("get target = %q", out)
	}

	if _, err := runCLI(t, "config", "clear", "global", "target", "--config", path); err != nil {
		t.Fatal(err)
	}
	if _, err = runCLI(t, "config", "get", "global", "target", "--config", path); !IsSilent(err) {
		t.Fatalf("expected silent error, got %v", err)
	}
}

func TestConfigSetWritesTOMLWhenConfigIsTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.toml")
	if err := os.WriteFile(path, []byte("[global]\nquiet = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "config", "set", "global", "show_hash=true", "--config", path); err != nil {
		t.Fatal(err)
	}
	content := readFileString(t, path)
	if !strings.Contains(content, "show_hash = true") || !strings.Contains(content, "[global]") {
		t.Fatalf("expected TOML output, got:\n%s", content)
	}
}

func TestConfigSetCreatesMissingFileAsYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", ".xget.yaml")
	if _, err := runCLI(t, "config", "set", "global", "quiet=true", "--config", path); err != nil {
		t.Fatal(err)
	}
	content := readFileString(t, path)
	if !strings.Contains(content, "quiet: true") {
		t.Fatalf("expected YAML output, got:\n%s", content)
	}
}

func TestConfigGetUnsetIsSilentFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	out, err := runCLI(t, "config", "get", "global", "target", "--config", path)
	if !IsSilent(err) {
		t.Fatalf("expected silent error, got %v", err)
	}
	if out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
	if ExitCodeFor(err) != 1 {
		t.Fatalf("exit code = %d", ExitCodeFor(err))
	}
}

func TestConfigInvalidInputsReportErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	cases := [][]string{
		{"config", "set", "global", "target", "--config", path},
		{"config", "set", "global", "bogus=1", "--config", path},
		{"config", "set", "notarepo", "quiet=true", "--config", path},
		{"config", "pop", "global", "ignore", "--config", path},
		{"config", "get", "global", "bogus", "--config", path},
		{"config", "set", "global", "quiet=true", "--config", filepath.Join(t.TempDir(), "conf.json")},
	}
	for _, args := range cases {
		if _, err := runCLI(t, args...); err == nil || IsSilent(err) {
			t.Fatalf("expected error for %v, got %v", args, err)
		}
	}
}

func TestConfigClearAndPopUnsetAreSilent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	if _, err := runCLI(t, "config", "clear", "global", "target", "--config", path); !IsSilent(err) {
		t.Fatalf("clear: %v", err)
	}
	if _, err := runCLI(t, "config", "pop", "global", "ignore=x", "--config", path); !IsSilent(err) {
		t.Fatalf("pop: %v", err)
	}
}

func TestConfigPopRemovesSingleListEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	for _, value := range []string{"a", "b"} {
		if _, err := runCLI(t, "config", "set", "global", "ignore="+value, "--config", path); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runCLI(t, "config", "pop", "global", "ignore=a", "--config", path); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "config", "get", "global", "ignore", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "b" {
		t.Fatalf("ignore = %q", out)
	}
}

func TestConfigList(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	if err := os.WriteFile(path, []byte("global:\n  target: \"~/bin\"\n  ignore: [a, b]\n\"o/r\":\n  quiet: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "config", "list", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	want := "global.ignore=a\nglobal.ignore=b\nglobal.target=~/bin\no/r.quiet=true\n"
	if out != want {
		t.Fatalf("list output = %q, want %q", out, want)
	}
}

func TestConfigListEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	out, err := runCLI(t, "config", "list", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("expected empty output, got %q", out)
	}
}

func TestConfigPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	out, err := runCLI(t, "config", "path", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != path {
		t.Fatalf("path = %q, want %q", out, path)
	}
}

func TestConfigPathFallsBackToXDGDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Chdir(t.TempDir())

	out, err := runCLI(t, "config", "path")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "xdg", "xget", ".xget.yml")
	if strings.TrimSpace(out) != want {
		t.Fatalf("path = %q, want %q", out, want)
	}
}

func TestConfigCommandWithoutSubcommandPrintsHelp(t *testing.T) {
	out, err := runCLI(t, "config")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Available Commands") {
		t.Fatalf("expected help output, got %q", out)
	}
}

func TestConfigEditCreatesFileAndRunsEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", ".xget.yml")

	original := runEditor
	defer func() { runEditor = original }()

	var gotName string
	var gotArgs []string
	runEditor = func(name string, args []string) error {
		gotName = name
		gotArgs = args
		return nil
	}

	t.Setenv("XGET_EDITOR", "code --wait")
	if _, err := runCLI(t, "config", "edit", "--config", path); err != nil {
		t.Fatal(err)
	}
	if gotName != "code" {
		t.Fatalf("editor = %q", gotName)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "--wait" || gotArgs[1] != path {
		t.Fatalf("args = %v", gotArgs)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}

func TestConfigEditReportsEditorFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	original := runEditor
	defer func() { runEditor = original }()
	runEditor = func(string, []string) error { return errors.New("boom") }

	t.Setenv("XGET_EDITOR", "fake-editor")
	_, err := runCLI(t, "config", "edit", "--config", path)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected editor failure, got %v", err)
	}
}

func TestConfigEditRejectsInvalidConfigAfterEditing(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	original := runEditor
	defer func() { runEditor = original }()
	runEditor = func(string, []string) error {
		return os.WriteFile(path, []byte("global:\n  quiet: true\n bad\n"), 0o600)
	}

	t.Setenv("XGET_EDITOR", "fake-editor")
	if _, err := runCLI(t, "config", "edit", "--config", path); err == nil {
		t.Fatal("expected parse error after edit")
	}
}

func TestResolveEditorPrefersEnvVarsInOrder(t *testing.T) {
	t.Setenv("EDITOR", "vi")
	t.Setenv("VISUAL", "vim")
	t.Setenv("XGET_EDITOR", "nvim -u NONE")

	name, args, err := resolveEditor()
	if err != nil {
		t.Fatal(err)
	}
	if name != "nvim" || len(args) != 2 || args[0] != "-u" || args[1] != "NONE" {
		t.Fatalf("editor = %q %v", name, args)
	}

	t.Setenv("XGET_EDITOR", "  ")
	if name, _, err = resolveEditor(); err != nil || name != "vim" {
		t.Fatalf("VISUAL fallback = %q, %v", name, err)
	}

	t.Setenv("VISUAL", "")
	if name, _, err = resolveEditor(); err != nil || name != "vi" {
		t.Fatalf("EDITOR fallback = %q, %v", name, err)
	}
}

func TestResolveEditorFallsBackToNano(t *testing.T) {
	t.Setenv("XGET_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	original := lookPath
	defer func() { lookPath = original }()
	lookPath = func(file string) (string, error) {
		if file == "nano" {
			return "/usr/bin/nano", nil
		}
		return "", errors.New("not found")
	}

	name, args, err := resolveEditor()
	if err != nil {
		t.Fatal(err)
	}
	if name != "nano" || len(args) != 0 {
		t.Fatalf("editor = %q %v", name, args)
	}
}

func TestResolveEditorWithoutAnyEditor(t *testing.T) {
	t.Setenv("XGET_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	original := lookPath
	defer func() { lookPath = original }()
	lookPath = func(string) (string, error) { return "", errors.New("not found") }

	name, _, err := resolveEditor()
	if runtime.GOOS == "windows" {
		if err != nil || name != "notepad" {
			t.Fatalf("windows fallback = %q, %v", name, err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error, got editor %q", name)
	}
}

func TestRunEditorMissingBinary(t *testing.T) {
	original := lookPath
	defer func() { lookPath = original }()
	lookPath = func(string) (string, error) { return "", errors.New("not found") }

	if err := runEditor("definitely-not-a-real-editor", nil); err == nil {
		t.Fatal("expected error for missing editor binary")
	}
}
