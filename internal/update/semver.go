package update

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semantic version (SemVer 2.0.0): MAJOR.MINOR.PATCH with
// an optional pre-release part. Build metadata (+...) is accepted and ignored.
type Version struct {
	Major, Minor, Patch uint64
	Pre                 []string // pre-release identifiers, empty for a release
}

// IsPrerelease reports whether the version carries a pre-release suffix.
func (v Version) IsPrerelease() bool { return len(v.Pre) > 0 }

// String renders the version without the "v" prefix and without metadata.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		s += "-" + strings.Join(v.Pre, ".")
	}
	return s
}

// ParseVersion parses "v2.4.0", "2.4.0-rc.1", "v2.4.0-beta.2+build.5", ...
func ParseVersion(raw string) (Version, error) {
	var v Version
	s := strings.TrimSpace(raw)
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = s[1:]
	}
	if s == "" || len(s) > 128 {
		return v, fmt.Errorf("invalid version %q", raw)
	}
	if plus := strings.IndexByte(s, '+'); plus >= 0 {
		if !validIdents(s[plus+1:], false) {
			return v, fmt.Errorf("invalid version %q", raw)
		}
		s = s[:plus]
	}
	if dash := strings.IndexByte(s, '-'); dash >= 0 {
		pre := s[dash+1:]
		if !validIdents(pre, true) {
			return v, fmt.Errorf("invalid version %q", raw)
		}
		v.Pre = strings.Split(pre, ".")
		s = s[:dash]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("invalid version %q", raw)
	}
	nums := [3]*uint64{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		n, err := parseNumber(p)
		if err != nil {
			return v, fmt.Errorf("invalid version %q", raw)
		}
		*nums[i] = n
	}
	return v, nil
}

func parseNumber(s string) (uint64, error) {
	if s == "" || (len(s) > 1 && s[0] == '0') || !isNumeric(s) {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return strconv.ParseUint(s, 10, 64)
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func validIdents(s string, numericRules bool) bool {
	if s == "" {
		return false
	}
	for _, id := range strings.Split(s, ".") {
		if id == "" {
			return false
		}
		for i := 0; i < len(id); i++ {
			c := id[i]
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return false
			}
		}
		if numericRules && isNumeric(id) && len(id) > 1 && id[0] == '0' {
			return false
		}
	}
	return true
}

// Compare returns -1, 0 or 1. A release is newer than its own pre-releases
// (2.4.0 > 2.4.0-rc.1); pre-release identifiers are compared one by one,
// numeric ones numerically and before alphanumeric ones.
func Compare(a, b Version) int {
	for _, pair := range [][2]uint64{{a.Major, b.Major}, {a.Minor, b.Minor}, {a.Patch, b.Patch}} {
		if pair[0] != pair[1] {
			if pair[0] > pair[1] {
				return 1
			}
			return -1
		}
	}
	switch {
	case len(a.Pre) == 0 && len(b.Pre) == 0:
		return 0
	case len(a.Pre) == 0:
		return 1
	case len(b.Pre) == 0:
		return -1
	}
	for i := 0; i < len(a.Pre) && i < len(b.Pre); i++ {
		if c := compareIdent(a.Pre[i], b.Pre[i]); c != 0 {
			return c
		}
	}
	switch {
	case len(a.Pre) > len(b.Pre):
		return 1
	case len(a.Pre) < len(b.Pre):
		return -1
	}
	return 0
}

func compareIdent(a, b string) int {
	an, bn := isNumeric(a), isNumeric(b)
	switch {
	case an && bn:
		if len(a) != len(b) {
			if len(a) > len(b) {
				return 1
			}
			return -1
		}
		return strings.Compare(a, b)
	case an:
		return -1
	case bn:
		return 1
	}
	return strings.Compare(a, b)
}

// CompareStrings compares two version strings; ok is false when either one
// cannot be parsed.
func CompareStrings(a, b string) (cmp int, ok bool) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, false
	}
	vb, err := ParseVersion(b)
	if err != nil {
		return 0, false
	}
	return Compare(va, vb), true
}
