package jsonutil

import "testing"

func TestUnmarshalStripsUTF8BOM(t *testing.T) {
	var out struct {
		OK bool `json:"ok"`
	}
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"ok":true}`)...)
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal() returned error for BOM-prefixed JSON: %v", err)
	}
	if !out.OK {
		t.Fatalf("Unmarshal() did not decode JSON payload")
	}
}
