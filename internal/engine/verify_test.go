package engine

import "testing"

func TestSha256HexTokenAcceptsCommonChecksumFileFormats(t *testing.T) {
	checksum := "7acedd1a84a4cfcb6e7a16003242945e6c269f257815291d0a4aa06e4ad06fd2"
	cases := map[string]string{
		checksum:                        checksum,
		checksum + "\n":                 checksum,
		checksum + "  asset.zip\n":      checksum,
		"sha256:" + checksum:            checksum,
		"sha256:" + checksum + "\r\n":   checksum,
		"sha256:" + checksum + " asset": checksum,
	}

	for input, want := range cases {
		if got := sha256HexToken(input); got != want {
			t.Fatalf("sha256HexToken(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNewSha256VerifierRejectsInvalidHex(t *testing.T) {
	_, err := NewSha256Verifier("not-a-checksum")
	if err == nil {
		t.Fatal("expected invalid checksum to return an error")
	}
}
