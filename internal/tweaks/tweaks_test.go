package tweaks

import "testing"

func TestCachedPackageInstalledSupportsWildcard(t *testing.T) {
	stateCacheMu.Lock()
	oldCache, oldReady := appxCache, appxCacheReady
	appxCache = map[string]struct{}{
		"microsoft.windows.widgets_8wekyb3d8bbwe": {},
	}
	appxCacheReady = true
	stateCacheMu.Unlock()
	t.Cleanup(func() {
		stateCacheMu.Lock()
		appxCache, appxCacheReady = oldCache, oldReady
		stateCacheMu.Unlock()
	})

	if installed, known := cachedPackageInstalled("Microsoft.Windows.Widgets*"); !known || !installed {
		t.Fatalf("expected wildcard package to be found, installed=%v known=%v", installed, known)
	}
	if installed, known := cachedPackageInstalled("Microsoft.NotInstalled"); !known || installed {
		t.Fatalf("expected missing package, installed=%v known=%v", installed, known)
	}
}

func TestServiceStartTypeMatches(t *testing.T) {
	tests := []struct {
		actual   string
		expected string
		want     bool
	}{
		{"Disabled", "disabled", true},
		{"Manual", "demand", true},
		{"Automatic", "auto", true},
		{"Auto", "automatic", true},
		{"Manual", "disabled", false},
	}
	for _, test := range tests {
		if got := serviceStartTypeMatches(test.actual, test.expected); got != test.want {
			t.Errorf("serviceStartTypeMatches(%q, %q)=%v, want %v", test.actual, test.expected, got, test.want)
		}
	}
}
