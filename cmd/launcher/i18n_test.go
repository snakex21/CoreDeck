package main

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Required CLDR cardinal plural categories per language (the forms a
// translation must provide for every *_one/_few/... key family).
var pluralCategories = map[string][]string{
	"pl": {"one", "few", "many", "other"}, "cs": {"one", "few", "many", "other"},
	"sk": {"one", "few", "many", "other"}, "ru": {"one", "few", "many", "other"},
	"uk": {"one", "few", "many", "other"}, "lt": {"one", "few", "many", "other"},
	"hr": {"one", "few", "other"}, "sr-Latn": {"one", "few", "other"}, "ro": {"one", "few", "other"},
	"sl": {"one", "two", "few", "other"}, "lv": {"zero", "one", "other"},
	"zh": {"other"}, "ja": {"other"}, "ko": {"other"},
}

// The 27 languages shared with the author's other projects, plus zh/ja/ko.
var expectedLanguages = []string{
	"bg", "cs", "da", "de", "el", "en", "es", "et", "fi", "fr", "hr", "hu", "it", "ja", "ko",
	"lt", "lv", "nb", "nl", "pl", "pt", "ro", "ru", "sk", "sl", "sr-Latn", "sv", "tr", "uk", "zh",
}

var (
	langBlock   = regexp.MustCompile(`(?m)^  (?:'([a-zA-Z-]+)'|([a-z]+)): \{$`)
	keyLine     = regexp.MustCompile(`^\s+'([^']+)': '((?:[^'\\]|\\.)*)',?\s*$`)
	placeholder = regexp.MustCompile(`\{[a-z_]+\}`)
	pluralKey   = regexp.MustCompile(`^(.+)_(zero|one|two|few|many|other)$`)
)

// parseTranslations reads translations.js into lang -> key -> text.
func parseTranslations(t *testing.T) map[string]map[string]string {
	t.Helper()
	data, err := staticFiles.ReadFile("web/js/i18n/translations.js")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]map[string]string{}
	var current map[string]string
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if m := langBlock.FindStringSubmatch(line); m != nil {
			code := m[1] + m[2]
			current = map[string]string{}
			out[code] = current
			continue
		}
		if current == nil {
			continue
		}
		if m := keyLine.FindStringSubmatch(line); m != nil {
			current[m[1]] = m[2]
		}
	}
	return out
}

func placeholders(s string) string {
	found := placeholder.FindAllString(s, -1)
	sort.Strings(found)
	return strings.Join(found, ",")
}

func TestTranslationsAreComplete(t *testing.T) {
	all := parseTranslations(t)
	var langs []string
	for code := range all {
		langs = append(langs, code)
	}
	sort.Strings(langs)
	if strings.Join(langs, " ") != strings.Join(expectedLanguages, " ") {
		t.Fatalf("languages = %v; want %v", langs, expectedLanguages)
	}

	en := all["en"]
	if len(en) < 600 {
		t.Fatalf("English has only %d keys; parser broken?", len(en))
	}
	families := map[string]bool{}
	for key := range all["pl"] {
		if m := pluralKey.FindStringSubmatch(key); m != nil && strings.HasPrefix(key, "count_") {
			families[m[1]] = true
		}
	}
	for _, lang := range langs {
		table := all[lang]
		for key, text := range en {
			if pluralKey.MatchString(key) && families[pluralKey.FindStringSubmatch(key)[1]] {
				continue
			}
			got, ok := table[key]
			if !ok || strings.TrimSpace(got) == "" {
				t.Errorf("%s: missing key %q", lang, key)
				continue
			}
			if placeholders(got) != placeholders(text) {
				t.Errorf("%s: %q placeholders %q, English has %q", lang, key, placeholders(got), placeholders(text))
			}
		}
		cats := pluralCategories[lang]
		if cats == nil {
			cats = []string{"one", "other"}
		}
		for family := range families {
			for _, cat := range cats {
				if v, ok := table[family+"_"+cat]; !ok || !strings.Contains(v, "{n}") {
					t.Errorf("%s: plural form %s_%s missing or without {n}", lang, family, cat)
				}
			}
		}
	}
}

func TestTweakTranslationsCoverEveryLanguage(t *testing.T) {
	data, err := staticFiles.ReadFile("web/js/i18n/tweak-translations.js")
	if err != nil {
		t.Fatal(err)
	}
	block := regexp.MustCompile(`(?m)^    (?:'([a-zA-Z-]+)'|([a-z]+)): \{$`)
	entry := regexp.MustCompile(`(?m)^        ([a-z0-9_]+): \{ name: '`)
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	locs := block.FindAllStringSubmatchIndex(text, -1)
	ids := map[string]map[string]bool{}
	for i, loc := range locs {
		var code string
		if loc[2] >= 0 {
			code = text[loc[2]:loc[3]]
		} else {
			code = text[loc[4]:loc[5]]
		}
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		ids[code] = map[string]bool{}
		for _, m := range entry.FindAllStringSubmatch(text[loc[1]:end], -1) {
			ids[code][m[1]] = true
		}
	}
	if len(ids["en"]) < 50 {
		t.Fatalf("English tweak texts: %d; parser broken?", len(ids["en"]))
	}
	for _, lang := range expectedLanguages {
		if lang == "pl" {
			continue // tweaks.json itself is Polish
		}
		got, ok := ids[lang]
		if !ok {
			t.Errorf("tweak translations missing for %s", lang)
			continue
		}
		for id := range ids["en"] {
			if !got[id] {
				t.Errorf("%s: tweak %q not translated", lang, id)
			}
		}
	}
}

// Every language in translations.js is offered in the picker (runtime.js).
func TestLanguagePickerListsEveryLanguage(t *testing.T) {
	data, err := staticFiles.ReadFile("web/js/i18n/runtime.js")
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	for _, lang := range expectedLanguages {
		if !strings.Contains(src, "['"+lang+"', '") {
			t.Errorf("runtime.js LANGUAGES has no entry for %s", lang)
		}
	}
}
