package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
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
		{name: "contained parent", entry: "dir/../tool"},
		{name: "contained backslash parent", entry: `dir\..\tool`},
		{name: "double dots in filename", entry: "tool..exe"},
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

func TestArchiveReadersAcceptContainedEntry(t *testing.T) {
	for _, entry := range []string{"tool", "dir/tool", "dir/", "./tool", "tool.exe"} {
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

type directoryChooser struct{}

func (directoryChooser) Choose(name string, dir bool, _ os.FileMode) (bool, bool) {
	return name == "dir/" && dir, false
}

func TestArchiveDirectoryExtractionRejectsTraversal(t *testing.T) {
	for _, format := range []string{"tar", "zip"} {
		for _, entry := range []string{"dir/../outside", `dir\..\outside`} {
			t.Run(format+"/"+entry, func(t *testing.T) {
				var data bytes.Buffer
				var extractor Extractor
				switch format {
				case "tar":
					writer := tar.NewWriter(&data)
					for _, header := range []*tar.Header{
						{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755},
						{Name: entry, Mode: 0600, Size: 4},
					} {
						if err := writer.WriteHeader(header); err != nil {
							t.Fatal(err)
						}
						if header.Size > 0 {
							if _, err := writer.Write([]byte("data")); err != nil {
								t.Fatal(err)
							}
						}
					}
					if err := writer.Close(); err != nil {
						t.Fatal(err)
					}
					extractor = NewExtractor("archive.tar", "", directoryChooser{})
				case "zip":
					writer := zip.NewWriter(&data)
					for _, name := range []string{"dir/", entry} {
						file, err := writer.Create(name)
						if err != nil {
							t.Fatal(err)
						}
						if name == entry {
							if _, err := file.Write([]byte("data")); err != nil {
								t.Fatal(err)
							}
						}
					}
					if err := writer.Close(); err != nil {
						t.Fatal(err)
					}
					extractor = NewExtractor("archive.zip", "", directoryChooser{})
				}

				selected, _, err := extractor.Extract(data.Bytes(), false)
				if err != nil {
					t.Fatal(err)
				}
				parent := t.TempDir()
				destination := filepath.Join(parent, "output")
				if err := os.Mkdir(destination, 0755); err != nil {
					t.Fatal(err)
				}
				err = selected.Extract(destination)
				if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
					t.Fatalf("Extract() error = %v, want unsafe archive path", err)
				}
				if _, err := os.Lstat(filepath.Join(parent, "outside")); !os.IsNotExist(err) {
					t.Fatalf("outside path exists or cannot be checked: %v", err)
				}
			})
		}
	}
}

func TestTarDirectoryExtractionLinks(t *testing.T) {
	tests := []struct {
		name    string
		typ     byte
		target  string
		wantErr bool
	}{
		{name: "hard link within directory", typ: tar.TypeLink, target: "dir/original"},
		{name: "hard link outside directory", typ: tar.TypeLink, target: "outside", wantErr: true},
		{name: "hard link sibling prefix", typ: tar.TypeLink, target: "dir-sibling/original", wantErr: true},
		{name: "hard link traversal", typ: tar.TypeLink, target: "dir/../outside", wantErr: true},
		{name: "hard link absolute", typ: tar.TypeLink, target: "/outside", wantErr: true},
		{name: "symlink within directory", typ: tar.TypeSymlink, target: "original"},
		{name: "symlink traversal", typ: tar.TypeSymlink, target: "../outside", wantErr: true},
		{name: "symlink backslash traversal", typ: tar.TypeSymlink, target: `..\outside`, wantErr: true},
		{name: "symlink absolute", typ: tar.TypeSymlink, target: "/outside", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && test.typ == tar.TypeSymlink && !test.wantErr {
				t.Skip("creating symlinks on Windows may require additional privileges")
			}
			var data bytes.Buffer
			writer := tar.NewWriter(&data)
			for _, header := range []*tar.Header{
				{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755},
				{Name: "dir/original", Mode: 0600, Size: 4},
				{Name: "dir/link", Typeflag: test.typ, Linkname: test.target, Mode: 0600},
			} {
				if err := writer.WriteHeader(header); err != nil {
					t.Fatal(err)
				}
				if header.Size > 0 {
					if _, err := writer.Write([]byte("data")); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			selected, _, err := NewExtractor("archive.tar", "", directoryChooser{}).Extract(data.Bytes(), false)
			if err != nil {
				t.Fatal(err)
			}
			destination := t.TempDir()
			err = selected.Extract(destination)
			if test.wantErr {
				if err == nil {
					t.Fatal("Extract() succeeded with unsafe link target")
				}
				if _, err := os.Lstat(filepath.Join(destination, "link")); !os.IsNotExist(err) {
					t.Fatalf("link exists or cannot be checked: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			original, err := os.Stat(filepath.Join(destination, "original"))
			if err != nil {
				t.Fatal(err)
			}
			linked, err := os.Stat(filepath.Join(destination, "link"))
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(original, linked) {
				t.Fatal("link does not refer to extracted original")
			}
		})
	}
}

func TestHardLinkCannotFollowSymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks on Windows may require additional privileges")
	}
	base := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "file"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(base, "escape")); err != nil {
		t.Fatal(err)
	}
	err := (link{base: base, newname: filepath.Join(base, "link"), oldname: "escape/file"}).Write()
	if err == nil {
		t.Fatal("Write() succeeded through a symlink outside the root")
	}
	if _, err := os.Lstat(filepath.Join(base, "link")); !os.IsNotExist(err) {
		t.Fatalf("link exists or cannot be checked: %v", err)
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
