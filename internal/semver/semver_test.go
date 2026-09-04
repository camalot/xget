package semver

import "testing"

func TestParseAcceptsCommonTagShapes(t *testing.T) {
	cases := map[string]struct {
		core [3]uint64
		pre  []string
	}{
		"1.2.3":              {core: [3]uint64{1, 2, 3}},
		"v1.2.3":             {core: [3]uint64{1, 2, 3}},
		"V1.2.3":             {core: [3]uint64{1, 2, 3}},
		"  v1.2.3  ":         {core: [3]uint64{1, 2, 3}},
		"2.0":                {core: [3]uint64{2, 0, 0}},
		"3":                  {core: [3]uint64{3, 0, 0}},
		"1.2.3+build.5":      {core: [3]uint64{1, 2, 3}},
		"v2.0.0-beta":        {core: [3]uint64{2, 0, 0}, pre: []string{"beta"}},
		"v2.0.0-beta.1":      {core: [3]uint64{2, 0, 0}, pre: []string{"beta", "1"}},
		"v2.0.0-rc.1+meta.2": {core: [3]uint64{2, 0, 0}, pre: []string{"rc", "1"}},
	}
	for tag, want := range cases {
		got, ok := parse(tag)
		if !ok {
			t.Fatalf("parse(%q) reported failure", tag)
		}
		if got.core != want.core {
			t.Fatalf("parse(%q).core = %v, want %v", tag, got.core, want.core)
		}
		if len(got.prerelease) != len(want.pre) {
			t.Fatalf("parse(%q).prerelease = %v, want %v", tag, got.prerelease, want.pre)
		}
		for i := range want.pre {
			if got.prerelease[i] != want.pre[i] {
				t.Fatalf("parse(%q).prerelease = %v, want %v", tag, got.prerelease, want.pre)
			}
		}
	}
}

func TestParseRejectsNonSemver(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"v",
		"latest",
		"main",
		"1.2.3.4",
		"1.x.3",
		"v1.2.3-",
		"v1.2.3-beta..1",
		"release-2024",
		"-1.2.3",
	}
	for _, tag := range invalid {
		if _, ok := parse(tag); ok {
			t.Fatalf("parse(%q) unexpectedly succeeded", tag)
		}
	}
}

func TestCompareOrdersCoreAndPrerelease(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"v1.2.3", "1.2.3", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"1.3.0", "1.2.9", 1},
		{"1.2.4", "1.2.3", 1},
		{"2.0.0", "2.0.0-beta", 1},
		{"2.0.0-beta", "2.0.0", -1},
		{"2.0.0-beta.2", "2.0.0-beta.1", 1},
		{"2.0.0-beta.11", "2.0.0-beta.2", 1},
		{"2.0.0-alpha", "2.0.0-beta", -1},
		{"2.0.0-beta", "2.0.0-beta.1", -1},
		{"2.0.0-1", "2.0.0-alpha", -1},
		{"2.0.0-alpha", "2.0.0-1", 1},
		{"1.2.3+a", "1.2.3+b", 0},
	}
	for _, tc := range cases {
		got, ok := Compare(tc.a, tc.b)
		if !ok {
			t.Fatalf("Compare(%q, %q) reported non-semver", tc.a, tc.b)
		}
		if got != tc.want {
			t.Fatalf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompareReportsNonSemver(t *testing.T) {
	if _, ok := Compare("latest", "1.2.3"); ok {
		t.Fatal("expected non-semver for left operand")
	}
	if _, ok := Compare("1.2.3", "nightly"); ok {
		t.Fatal("expected non-semver for right operand")
	}
}

func TestIsUpgrade(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v2.2.0", "v2.3.0", true},
		{"v2.3.0", "v2.3.0", false},
		{"v2.3.0", "v2.2.0", false},
		{"v2.0.0-beta", "v2.0.0", true},
		{"v2.0.0", "v2.0.0-beta", false},
		{"v2.0.0-beta", "v2.0.0-beta", false},
		{"v2.0.0-beta.1", "v2.0.0-beta.2", true},
		// Non-semver falls back to difference, since latest is the newest release.
		{"nightly-1", "nightly-2", true},
		{"nightly-2", "nightly-2", false},
		{"v1.0.0", "nightly", true},
		// Missing data.
		{"v1.0.0", "", false},
		{"", "", false},
		{"", "v1.0.0", true},
	}
	for _, tc := range cases {
		if got := IsUpgrade(tc.current, tc.latest); got != tc.want {
			t.Fatalf("IsUpgrade(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}
