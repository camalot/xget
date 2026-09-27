// Package engine implements asset discovery, download, verification, and extraction.
package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

// FileType classifies an archive entry.
type FileType byte

// Archive entry types.
const (
	TypeNormal FileType = iota
	TypeDir
	TypeLink
	TypeSymlink
	TypeOther
)

func tarft(typ byte) FileType {
	switch typ {
	case tar.TypeReg:
		return TypeNormal
	case tar.TypeDir:
		return TypeDir
	case tar.TypeLink:
		return TypeLink
	case tar.TypeSymlink:
		return TypeSymlink
	}
	return TypeOther
}

// File is an entry within an archive.
type File struct {
	Name     string
	LinkName string
	Mode     fs.FileMode
	Type     FileType
}

// Dir reports whether the entry is a directory.
func (f File) Dir() bool {
	return f.Type == TypeDir
}

// Archive iterates over entries in an archive.
type Archive interface {
	Next() (File, error)
	ReadAll() ([]byte, error)
}

func validateArchiveEntryName(name string) error {
	normalized := strings.ReplaceAll(name, "\\", "/")
	cleaned := path.Clean(normalized)
	if normalized == "" || strings.HasPrefix(normalized, "/") ||
		cleaned == ".." || strings.HasPrefix(cleaned, "../") ||
		isWindowsDrivePath(normalized) {
		return fmt.Errorf("unsafe archive path %q", name)
	}
	return nil
}

func isWindowsDrivePath(name string) bool {
	return len(name) >= 2 && name[1] == ':' &&
		((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z'))
}

// TarArchive is an Archive backed by a tar stream.
type TarArchive struct {
	r *tar.Reader
}

// NewTarArchive creates a TarArchive from data using the given decompressor.
func NewTarArchive(data []byte, decompress DecompFn) (Archive, error) {
	r := bytes.NewReader(data)
	dr, err := decompress(r)
	if err != nil {
		return nil, err
	}
	return &TarArchive{
		r: tar.NewReader(dr),
	}, nil
}

// Next advances to the next supported entry.
func (t *TarArchive) Next() (File, error) {
	for {
		hdr, err := t.r.Next()
		if err != nil {
			return File{}, err
		}
		if err := validateArchiveEntryName(hdr.Name); err != nil {
			return File{}, err
		}
		ft := tarft(hdr.Typeflag)
		if ft != TypeOther {
			return File{
				Name:     hdr.Name,
				LinkName: hdr.Linkname,
				Mode:     hdr.FileInfo().Mode(),
				Type:     ft,
			}, err
		}
	}
}

// ReadAll reads the contents of the current entry.
func (t *TarArchive) ReadAll() ([]byte, error) {
	return io.ReadAll(t.r)
}

// ZipArchive is an Archive backed by a zip file.
type ZipArchive struct {
	r   *zip.Reader
	idx int
}

// NewZipArchive creates a ZipArchive from data. The decompressor is ignored
// because zip has built-in compression.
func NewZipArchive(data []byte, _ DecompFn) (Archive, error) {
	r := bytes.NewReader(data)
	zr, err := zip.NewReader(r, int64(len(data)))
	return &ZipArchive{
		r:   zr,
		idx: -1,
	}, err
}

// Next advances to the next entry.
func (z *ZipArchive) Next() (File, error) {
	z.idx++

	if z.idx < 0 || z.idx >= len(z.r.File) {
		return File{}, io.EOF
	}

	f := z.r.File[z.idx]
	if err := validateArchiveEntryName(f.Name); err != nil {
		return File{}, err
	}

	typ := TypeNormal
	if strings.HasSuffix(f.Name, "/") {
		typ = TypeDir
	}

	return File{
		Name: f.Name,
		Mode: f.Mode(),
		Type: typ,
	}, nil
}

// ReadAll reads the contents of the current entry.
func (z *ZipArchive) ReadAll() ([]byte, error) {
	if z.idx < 0 || z.idx >= len(z.r.File) {
		return nil, io.EOF
	}
	f := z.r.File[z.idx]
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("zip extract: %w", err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			fmt.Println("zip extract: close error:", err)
		}
	}()
	data, err := io.ReadAll(rc)
	return data, err
}
