package jsonutil

import (
	"bytes"
	"encoding/json"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// StripUTF8BOM removes an optional UTF-8 byte-order mark from JSON data.
// Some editors/PowerShell commands write JSON files with a BOM; encoding/json
// rejects that with "invalid character '\ufeff' looking for beginning of value".
func StripUTF8BOM(data []byte) []byte {
	return bytes.TrimPrefix(data, utf8BOM)
}

// Unmarshal is encoding/json.Unmarshal with tolerance for a leading UTF-8 BOM.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(StripUTF8BOM(data), v)
}
