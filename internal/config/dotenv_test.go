package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDotenvPathsUsesRequestedPriorityAndAlphabeticalWildcards(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{".env", "z.env", "a.env", ".secrets", "z.secrets", "a.secrets"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("VALUE=test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got := dotenvPaths(directory)
	want := []string{
		filepath.Join(directory, ".secrets"),
		filepath.Join(directory, "a.secrets"),
		filepath.Join(directory, "z.secrets"),
		filepath.Join(directory, ".env"),
		filepath.Join(directory, "a.env"),
		filepath.Join(directory, "z.env"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dotenvPaths() = %#v, want %#v", got, want)
	}
}

func TestLoadDotenvFilesPreservesExistingEnvironmentAndPriority(t *testing.T) {
	directory := t.TempDir()
	t.Chdir(directory)
	if err := os.WriteFile(".secrets", []byte("XGET_DOTENV_TEST=secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".env", []byte("XGET_DOTENV_TEST=env\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XGET_DOTENV_TEST", "")
	if err := os.Unsetenv("XGET_DOTENV_TEST"); err != nil {
		t.Fatal(err)
	}
	if err := LoadDotenvFiles(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("XGET_DOTENV_TEST"); got != "secret" {
		t.Fatalf("value = %q, want secret", got)
	}

	t.Setenv("XGET_DOTENV_TEST", "shell")
	if err := LoadDotenvFiles(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("XGET_DOTENV_TEST"); got != "shell" {
		t.Fatalf("value = %q, want shell", got)
	}
}
