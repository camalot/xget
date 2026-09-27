package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/camalot/xget/internal/config"
)

// Verifier checks downloaded bytes.
type Verifier interface {
	Verify(b []byte) error
}

// NoVerifier accepts any input.
type NoVerifier struct{}

// Verify always succeeds.
func (n *NoVerifier) Verify(_ []byte) error {
	return nil
}

// Sha256Error reports a checksum mismatch.
type Sha256Error struct {
	Expected []byte
	Got      []byte
}

func (e *Sha256Error) Error() string {
	return fmt.Sprintf("sha256 checksum mismatch:\nexpected: %x\ngot:      %x", e.Expected, e.Got)
}

// Sha256Verifier checks data against an expected SHA-256 sum.
type Sha256Verifier struct {
	Expected []byte
}

// NewSha256Verifier parses expectedHex into a Sha256Verifier.
func NewSha256Verifier(expectedHex string) (*Sha256Verifier, error) {
	expectedHex = sha256HexToken(expectedHex)
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return nil, err
	}
	if len(expected) != sha256.Size {
		return nil, fmt.Errorf("sha256sum (%s) too small: %d bytes decoded", expectedHex, len(expectedHex))
	}
	return &Sha256Verifier{
		Expected: expected,
	}, nil
}

func sha256HexToken(checksum string) string {
	fields := strings.Fields(checksum)
	for _, field := range fields {
		candidate := strings.TrimPrefix(field, "sha256:")
		decoded, err := hex.DecodeString(candidate)
		if err == nil && len(decoded) == sha256.Size {
			return candidate
		}
	}
	return ""
}

// Verify compares the SHA-256 of b to the expected sum.
func (s256 *Sha256Verifier) Verify(b []byte) error {
	sum := sha256.Sum256(b)
	if bytes.Equal(sum[:], s256.Expected) {
		return nil
	}
	return &Sha256Error{
		Expected: s256.Expected,
		Got:      sum[:],
	}
}

// Sha256Printer prints the SHA-256 of the data instead of verifying it.
type Sha256Printer struct{}

// Verify prints the SHA-256 of b and always succeeds.
func (s256 *Sha256Printer) Verify(b []byte) error {
	sum := sha256.Sum256(b)
	fmt.Printf("%x\n", sum)
	return nil
}

// Sha256AssetVerifier verifies data against a checksum downloaded from AssetURL.
type Sha256AssetVerifier struct {
	AssetURL string
	Source   config.Source
}

// Verify downloads the checksum asset and compares it to the SHA-256 of b.
func (s256 *Sha256AssetVerifier) Verify(b []byte) error {
	resp, err := GetWithSource(s256.AssetURL, s256.Source)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	expectedHex := sha256HexToken(string(data))
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return err
	}
	if len(expected) != sha256.Size {
		return fmt.Errorf("sha256sum (%s) too small: %d bytes decoded", expectedHex, len(expected))
	}
	sum := sha256.Sum256(b)
	if bytes.Equal(sum[:], expected) {
		return nil
	}
	return &Sha256Error{
		Expected: expected,
		Got:      sum[:],
	}
}
