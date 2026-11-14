package main

import (
	"regexp"
	"testing"
)

// The pendrive can be mounted under any letter, so UI defaults must come from
// get_portable_paths (derived from the exe location), never from a literal
// like 'L:\Export'.
func TestImportExportHasNoHardcodedDriveLetters(t *testing.T) {
	const file = "web/js/pages/manage/import-export.js"
	data, err := staticFiles.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	literal := regexp.MustCompile("['\"`][A-Za-z]:\\\\")
	if loc := literal.FindIndex(data); loc != nil {
		t.Fatalf("%s contains a hardcoded drive-letter path near %q", file, data[loc[0]:min(len(data), loc[0]+40)])
	}
}
