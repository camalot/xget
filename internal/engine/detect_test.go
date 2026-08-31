package engine

import (
	"strings"
	"testing"

	"github.com/camalot/xget/internal/options"
)

func TestGetDetector_IgnoreLiteralWithAssetFilter(t *testing.T) {
	opts := &options.Flags{
		System: "all",
		Asset:  []string{".zip"},
		Ignore: []string{".zip.sbom.json"},
	}

	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/tool_1.0.0_linux_amd64.zip",
		"https://example.com/tool_1.0.0_linux_amd64.zip.sbom.json",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, ".zip") || strings.HasSuffix(choice, ".zip.sbom.json") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_IgnoreRegexWithAssetFilter(t *testing.T) {
	opts := &options.Flags{
		System: "all",
		Asset:  []string{".zip"},
		Ignore: []string{`~\.zip\.sbom\.json$`},
	}

	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/tool_1.0.0_linux_amd64.zip",
		"https://example.com/tool_1.0.0_linux_amd64.zip.sbom.json",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, ".zip") || strings.HasSuffix(choice, ".zip.sbom.json") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_AssetRegexAndNegativeRegex(t *testing.T) {
	opts := &options.Flags{
		System: "all",
		Asset:  []string{`~^tool_.*`, `not:re:.*\.sbom\.json$`},
	}

	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/tool_linux_amd64.zip",
		"https://example.com/tool_linux_amd64.zip.sbom.json",
		"https://example.com/other_linux_amd64.tar.gz",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "tool_linux_amd64.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_InvalidRegex(t *testing.T) {
	opts := &options.Flags{System: "all", Asset: []string{"~("}}
	_, err := getDetector(opts)
	if err == nil {
		t.Fatal("expected error for invalid asset regex")
	}

	opts = &options.Flags{System: "all", Ignore: []string{"~("}}
	_, err = getDetector(opts)
	if err == nil {
		t.Fatal("expected error for invalid ignore regex")
	}
}

func TestGetDetector_RePrefixIsRegexLongForm(t *testing.T) {
	opts := &options.Flags{System: "all", Asset: []string{`re:^tool_.*\.zip$`}}
	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/tool_1.0.0.zip",
		"https://example.com/other.zip",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "tool_1.0.0.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_EscapedLeadingTildeAndCaretAreLiteral(t *testing.T) {
	opts := &options.Flags{System: "all", Asset: []string{`~~tool`}}
	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/~tool.zip",
		"https://example.com/tool.zip",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "~tool.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}

	opts = &options.Flags{System: "all", Asset: []string{`^^tool`}}
	detector, err = getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets = []string{
		"https://example.com/^tool.zip",
		"https://example.com/tool.zip",
	}

	choice, candidates, err = detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "^tool.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_TextPrefixForLiteral(t *testing.T) {
	opts := &options.Flags{System: "all", Asset: []string{`text:~tool`}}
	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/~tool.zip",
		"https://example.com/tool.zip",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "~tool.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}

func TestGetDetector_IgnoreSupportsNegativeForms(t *testing.T) {
	opts := &options.Flags{
		System: "all",
		Ignore: []string{"not:arm64"},
	}

	detector, err := getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	assets := []string{
		"https://example.com/tool_linux_amd64.zip",
		"https://example.com/tool_linux_arm64.zip",
	}

	choice, candidates, err := detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "tool_linux_arm64.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}

	opts = &options.Flags{
		System: "all",
		Ignore: []string{"^~arm64"},
	}

	detector, err = getDetector(opts)
	if err != nil {
		t.Fatalf("getDetector returned error: %v", err)
	}

	choice, candidates, err = detector.Detect(assets)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(candidates))
	}
	if !strings.HasSuffix(choice, "tool_linux_arm64.zip") {
		t.Fatalf("unexpected chosen asset: %s", choice)
	}
}
