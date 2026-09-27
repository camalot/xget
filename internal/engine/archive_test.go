package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestArchiveReadersRejectUnsafeEntryNames(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{name: "parent", entry: "../outside"},
		{name: "nested parent", entry: "dir/../../outside"},
		{name: "absolute", entry: "/outside"},
		{name: "backslash parent", entry: `..\outside`},
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
				if err == nil {
					t.Fatalf("Next() accepted unsafe archive entry %q", test.entry)
				}
			})
		}
	}
}

func TestArchiveReadersAcceptCleanedContainedEntry(t *testing.T) {
	for _, newArchive := range []func(*testing.T, string) Archive{newTestZipArchive, newTestTarArchive} {
		file, err := newArchive(t, "dir/../tool").Next()
		if err != nil {
			t.Fatalf("Next() rejected contained archive entry: %v", err)
		}
		if file.Name != "dir/../tool" {
			t.Fatalf("Next() name = %q, want %q", file.Name, "dir/../tool")
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
	if _, err := file.Write([]byte("data")); err != nil {
		t.Fatal(err)
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
	if err := writer.WriteHeader(&tar.Header{Name: entry, Mode: 0600, Size: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("data")); err != nil {
		t.Fatal(err)
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
