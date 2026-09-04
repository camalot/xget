// Package semver provides the minimal semantic version comparison xget needs to
// decide whether a release tag is newer than an installed one.
package semver

import (
	"strconv"
	"strings"
)

type parsed struct {
	core       [3]uint64
	prerelease []string
}

// parse interprets a release tag such as "v1.2.3-beta.1+build" as a semantic
// version. It reports false when the tag is not semver-shaped.
func parse(tag string) (parsed, bool) {
	var out parsed

	value := strings.TrimSpace(tag)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "v"), "V")
	if value == "" {
		return out, false
	}
	if idx := strings.IndexByte(value, '+'); idx >= 0 {
		value = value[:idx]
	}

	core := value
	if idx := strings.IndexByte(value, '-'); idx >= 0 {
		core = value[:idx]
		pre := value[idx+1:]
		if pre == "" {
			return out, false
		}
		out.prerelease = strings.Split(pre, ".")
		for _, identifier := range out.prerelease {
			if identifier == "" {
				return out, false
			}
		}
	}

	parts := strings.Split(core, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		number, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return out, false
		}
		out.core[i] = number
	}
	return out, true
}

func compareIdentifiers(a, b string) int {
	aNum, aErr := strconv.ParseUint(a, 10, 64)
	bNum, bErr := strconv.ParseUint(b, 10, 64)
	switch {
	case aErr == nil && bErr == nil:
		return compareUint(aNum, bNum)
	case aErr == nil:
		// Numeric identifiers always have lower precedence than alphanumeric ones.
		return -1
	case bErr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func compareUint(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func comparePrerelease(a, b []string) int {
	// A version without a prerelease outranks one with a prerelease.
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return 1
	case len(b) == 0:
		return -1
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if cmp := compareIdentifiers(a[i], b[i]); cmp != 0 {
			return cmp
		}
	}
	return compareUint(uint64(len(a)), uint64(len(b)))
}

// Compare returns -1, 0, or 1 comparing two semver tags. It reports false when
// either tag is not semver-shaped.
func Compare(a, b string) (int, bool) {
	left, ok := parse(a)
	if !ok {
		return 0, false
	}
	right, ok := parse(b)
	if !ok {
		return 0, false
	}
	for i := range left.core {
		if cmp := compareUint(left.core[i], right.core[i]); cmp != 0 {
			return cmp, true
		}
	}
	return comparePrerelease(left.prerelease, right.prerelease), true
}

// IsUpgrade reports whether latest is newer than current. When either tag is not
// semver-shaped it falls back to reporting any difference as an upgrade, since
// the latest tag is by definition the newest release the source knows about.
func IsUpgrade(current, latest string) bool {
	if latest == "" {
		return false
	}
	if current == "" {
		return true
	}
	if cmp, ok := Compare(latest, current); ok {
		return cmp > 0
	}
	return latest != current
}
