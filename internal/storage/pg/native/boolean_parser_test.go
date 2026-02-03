package native

import (
	"testing"
)

func TestBooleanParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple AND",
			input:    "climate AND change",
			expected: "climate & change",
		},
		{
			name:     "simple OR",
			input:    "renewable OR sustainable",
			expected: "renewable | sustainable",
		},
		{
			name:     "NOT",
			input:    "climate AND NOT politics",
			expected: "climate & ! politics",
		},
		{
			name:     "parenthesized OR with AND",
			input:    "(renewable OR sustainable) AND energy",
			expected: "( renewable | sustainable ) & energy",
		},
		{
			name:     "complex nested",
			input:    "(climate OR weather) AND change AND NOT politics",
			expected: "( climate | weather ) & change & ! politics",
		},
		{
			name:     "quoted phrase uses followed-by",
			input:    `"climate change" AND energy`,
			expected: "climate <-> change & energy",
		},
		{
			name:     "implicit AND between adjacent words",
			input:    "climate change",
			expected: "climate & change",
		},
		{
			name:     "implicit AND before paren",
			input:    "climate (change OR warming)",
			expected: "climate & ( change | warming )",
		},
		{
			name:     "implicit AND after paren",
			input:    "(climate OR weather) change",
			expected: "( climate | weather ) & change",
		},
		{
			name:     "case insensitive operators",
			input:    "climate and change or warming",
			expected: "climate & change | warming",
		},
		{
			name:     "single word",
			input:    "climate",
			expected: "climate",
		},
		{
			name:    "empty expression",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
	}

	p := NewBooleanParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got result: %q", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
