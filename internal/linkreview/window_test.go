package linkreview

import "testing"

func TestDecisionSucceeded(t *testing.T) {
	tests := []struct {
		name   string
		result string
		want   bool
	}{
		{name: "success", result: `{"success":true}`, want: true},
		{name: "api error", result: `{"error":"save failed"}`, want: false},
		{name: "invalid json", result: `not json`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := decisionSucceeded(test.result); got != test.want {
				t.Fatalf("decisionSucceeded(%q) = %v, want %v", test.result, got, test.want)
			}
		})
	}
}
