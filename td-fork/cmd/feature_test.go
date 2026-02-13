package cmd

import "testing"

func TestParseBoolString(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{input: "true", want: true},
		{input: "1", want: true},
		{input: "yes", want: true},
		{input: "on", want: true},
		{input: "false", want: false},
		{input: "0", want: false},
		{input: "no", want: false},
		{input: "off", want: false},
		{input: "maybe", wantErr: true},
	}

	for _, tt := range tests {
		got, err := parseBoolString(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseBoolString(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseBoolString(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseBoolString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
