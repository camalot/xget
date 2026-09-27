package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/camalot/xget/internal/installed"
)

func useVersion(t *testing.T, v string) {
	t.Helper()
	original := version
	version = v
	t.Cleanup(func() { version = original })
}

func stubRunningExecutable(t *testing.T, path string) {
	t.Helper()
	original := runningExecutable
	runningExecutable = func() (string, error) { return path, nil }
	t.Cleanup(func() { runningExecutable = original })
}

func stubNow(t *testing.T, current *time.Time) {
	t.Helper()
	original := now
	now = func() time.Time { return *current }
	t.Cleanup(func() { now = original })
}

func lastSelfCheck(t *testing.T, storePath string) time.Time {
	t.Helper()
	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	return store.SelfUpdate.LastChecked
}

const newVersionNotice = "v1.0.0 → v1.1.0"

func TestPrintUpdateNoticeIsBoxed(t *testing.T) {
	var buf strings.Builder
	printUpdateNotice(&buf, "v1.0.0", "v1.1.0")
	out := buf.String()
	for _, want := range []string{"A new version of xget is available!", newVersionNotice, "Run `xget self-update` to update.", "╭", "╯"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	lines := strings.Split(strings.Trim(out, "\n"), "\n")
	width := utf8.RuneCountInString(lines[0])
	for _, line := range lines {
		if got := utf8.RuneCountInString(line); got != width {
			t.Fatalf("line %q width = %d, want %d", line, got, width)
		}
	}
}

func TestSelfUpdateUpdatesRunningExecutableWhenUntracked(t *testing.T) {
	storePath := useTempInstalledStore(t)
	useVersion(t, "v1.0.0")
	stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})
	calls := stubEngine(t, storePath, nil)
	executable := filepath.Join(t.TempDir(), "xget")
	stubRunningExecutable(t, executable)

	out, err := runCLI(t, "self-update")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("calls = %+v", *calls)
	}
	call := (*calls)[0]
	if call.Target != xgetRepo || call.Opts.Tag != "v1.1.0" || call.Opts.Output != executable || !call.Opts.Untracked {
		t.Fatalf("call = %+v", call)
	}
	if !strings.Contains(out, "Upgrading xget from v1.0.0 to v1.1.0") {
		t.Fatalf("output:\n%s", out)
	}
	if records := storedPackages(t, storePath, "github:"+xgetRepo); len(records) != 0 {
		t.Fatalf("self-update should not start tracking xget: %+v", records)
	}
}

func TestSelfUpdateUntrackedAlreadyUpToDate(t *testing.T) {
	storePath := useTempInstalledStore(t)
	useVersion(t, "v1.1.0")
	stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})
	calls := stubEngine(t, storePath, nil)
	stubRunningExecutable(t, filepath.Join(t.TempDir(), "xget"))

	out, err := runCLI(t, "self-update")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("unexpected calls: %+v", *calls)
	}
	if !strings.Contains(out, "xget is already up to date (v1.1.0).") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestSelfCheckRunsOncePerDay(t *testing.T) {
	storePath := useTempInstalledStore(t)
	useVersion(t, "v1.0.0")
	seen := stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})
	current := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	stubNow(t, &current)

	out, err := runCLI(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, newVersionNotice) {
		t.Fatalf("expected update notice:\n%s", out)
	}
	if len(*seen) != 1 {
		t.Fatalf("lookups = %d, want 1", len(*seen))
	}
	if got := lastSelfCheck(t, storePath); !got.Equal(current) {
		t.Fatalf("last_checked = %v, want %v", got, current)
	}

	current = current.Add(23 * time.Hour)
	out, err = runCLI(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, newVersionNotice) || len(*seen) != 1 {
		t.Fatalf("expected no check within 24h (lookups = %d):\n%s", len(*seen), out)
	}

	current = current.Add(2 * time.Hour)
	if _, err := runCLI(t, "version"); err != nil {
		t.Fatal(err)
	}
	if len(*seen) != 2 {
		t.Fatalf("lookups = %d, want 2 after 24h", len(*seen))
	}
}

func TestUpgradeAndListInstalledAlwaysCheckWithoutRecording(t *testing.T) {
	storePath := useTempInstalledStore(t)
	useVersion(t, "v1.0.0")
	seen := stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})
	current := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	stubNow(t, &current)

	for _, stored := range []time.Time{{}, current.Add(-48 * time.Hour)} {
		store, err := installed.Load(storePath)
		if err != nil {
			t.Fatal(err)
		}
		store.SelfUpdate.LastChecked = stored
		if err := installed.Save(storePath, store); err != nil {
			t.Fatal(err)
		}
		*seen = nil

		commands := [][]string{{"upgrade"}, {"update"}, {"upgrade", "--all"}, {"list", "--installed"}}
		for _, args := range commands {
			out, err := runCLI(t, args...)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, newVersionNotice) {
				t.Fatalf("%v: expected update notice:\n%s", args, out)
			}
		}
		if len(*seen) != len(commands) {
			t.Fatalf("lookups = %d, want %d", len(*seen), len(commands))
		}
		if got := lastSelfCheck(t, storePath); !got.Equal(stored) {
			t.Fatalf("last_checked = %v, want unchanged %v", got, stored)
		}
	}
}

func TestSelfCheckSkippedWhenDisabledInConfig(t *testing.T) {
	storePath := useTempInstalledStore(t)
	t.Setenv("XGET_CONFIG", writeUpgradeConfig(t, "global:\n  xget_update_check: false\n"))
	useVersion(t, "v1.0.0")
	seen := stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})

	for _, args := range [][]string{{"version"}, {"upgrade"}, {"list", "--installed"}} {
		out, err := runCLI(t, args...)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out, newVersionNotice) {
			t.Fatalf("%v: unexpected update notice:\n%s", args, out)
		}
	}
	if len(*seen) != 0 {
		t.Fatalf("lookups = %d, want 0", len(*seen))
	}
	if got := lastSelfCheck(t, storePath); !got.IsZero() {
		t.Fatalf("last_checked = %v, want unset", got)
	}
}

func TestSelfCheckSkippedForDevelopmentBuilds(t *testing.T) {
	useTempInstalledStore(t)
	useVersion(t, developmentVersion)
	seen := stubRefresh(t, map[string]string{xgetRepo: "v1.1.0"})

	if _, err := runCLI(t, "upgrade"); err != nil {
		t.Fatal(err)
	}
	if len(*seen) != 0 {
		t.Fatalf("lookups = %d, want 0", len(*seen))
	}
}
