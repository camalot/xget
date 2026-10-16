package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"runtime"
	"strings"
	"testing"
)

func TestArchiveReadersRejectUnsafeEntryNames(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{name: "parent", entry: "../outside"},
		{name: "parent only", entry: ".."},
		{name: "nested parent", entry: "dir/../../outside"},
		{name: "absolute", entry: "/outside"},
		{name: "backslash parent", entry: `..\outside`},
		{name: "nested backslash parent", entry: `dir\..\..\outside`},
		{name: "mixed separators", entry: `dir/..\../outside`},
		{name: "backslash absolute", entry: `\outside`},
		{name: "UNC", entry: `\\server\share\outside`},
		{name: "drive absolute", entry: `C:\outside`},
		{name: "drive relative", entry: `C:outside`},
	}

	readers := []struct {
		name string
		new  func(t *testing.T, entry string) Archive
	}{
		{name: "zip", new: newTestZipArchive},
		{name: "tar", new: newTestTarArchive},
	}

	for _, reader := range readers {
		for _, test := range tests {
			t.Run(reader.name+"/"+test.name, func(t *testing.T) {
				_, err := reader.new(t, test.entry).Next()
				if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
					t.Fatalf("Next() error = %v, want unsafe archive path for %q", err, test.entry)
				}
			})
		}
	}
}

func TestValidateArchiveEntryNameRejectsInvalidPaths(t *testing.T) {
	entries := []string{"", "tool\x00name"}
	if runtime.GOOS == "windows" {
		entries = append(entries, "NUL", "dir/CON", "tool:stream")
	}
	for _, entry := range entries {
		t.Run(entry, func(t *testing.T) {
			name, err := validateArchiveEntryName(entry)
			if err == nil || name != "" {
				t.Fatalf("validateArchiveEntryName(%q) = (%q, %v), want empty name and error", entry, name, err)
			}
		})
	}
}

func TestArchiveReadersAcceptCleanedContainedEntry(t *testing.T) {
	for _, entry := range []string{"tool", "dir/tool", "dir/", "./tool", "dir/../tool", `dir\..\tool`} {
		for _, newArchive := range []func(*testing.T, string) Archive{newTestZipArchive, newTestTarArchive} {
			file, err := newArchive(t, entry).Next()
			if err != nil {
				t.Fatalf("Next() rejected contained archive entry %q: %v", entry, err)
			}
			if file.Name != entry {
				t.Fatalf("Next() name = %q, want %q", file.Name, entry)
			}
		}
	}
}

func newTestZipArchive(t *testing.T, entry string) Archive {
	t.Helper()

	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	file, err := writer.Create(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(entry, "/") {
		if _, err := file.Write([]byte("data")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	archive, err := NewZipArchive(data.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

func newTestTarArchive(t *testing.T, entry string) Archive {
	t.Helper()

	var data bytes.Buffer
	writer := tar.NewWriter(&data)
	header := &tar.Header{Name: entry, Mode: 0600, Size: 4}
	if strings.HasSuffix(entry, "/") {
		header.Typeflag = tar.TypeDir
		header.Size = 0
	}
	if err := writer.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if header.Size > 0 {
		if _, err := writer.Write([]byte("data")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	archive, err := NewTarArchive(data.Bytes(), func(reader io.Reader) (io.Reader, error) {
		return reader, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return archive
}
