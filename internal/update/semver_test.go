package update

import "testing"

func TestParseVersion(t *testing.T) {
	valid := map[string]string{
		"2.4.0":           "2.4.0",
		"v2.4.0":          "2.4.0",
		" V10.0.1 ":       "10.0.1",
		"v2.4.0-rc.1":     "2.4.0-rc.1",
		"2.4.0-beta.2+b7": "2.4.0-beta.2",
		"1.0.0-alpha-1.x": "1.0.0-alpha-1.x",
	}
	for in, want := range valid {
		v, err := ParseVersion(in)
		if err != nil || v.String() != want {
			t.Errorf("ParseVersion(%q) = %v, %v; want %s", in, v, err, want)
		}
	}
	for _, in := range []string{"", "v", "2.4", "2.4.0.1", "02.4.0", "2.4.0-", "2.4.0-rc..1", "2.4.0-01", "2.4.0-rc_1", "latest", "2.4.0+"} {
		if _, err := ParseVersion(in); err == nil {
			t.Errorf("ParseVersion(%q) succeeded", in)
		}
	}
}

func TestCompareOrdering(t *testing.T) {
	ordered := []string{
		"2.3.2",
		"2.4.0-alpha",
		"2.4.0-alpha.1",
		"2.4.0-alpha.beta",
		"2.4.0-beta.2",
		"2.4.0-beta.11",
		"2.4.0-rc.1",
		"2.4.0",
		"2.4.1",
		"2.10.0",
		"10.0.0",
	}
	for i := 0; i < len(ordered); i++ {
		for j := 0; j < len(ordered); j++ {
			got, ok := CompareStrings(ordered[i], ordered[j])
			if !ok {
				t.Fatalf("cannot parse %s / %s", ordered[i], ordered[j])
			}
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", ordered[i], ordered[j], got, want)
			}
		}
	}
	if c, _ := CompareStrings("v2.4.0+build.1", "2.4.0"); c != 0 {
		t.Errorf("build metadata must be ignored")
	}
}
