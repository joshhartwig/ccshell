package keylistener

import (
	"testing"
)

func TestIsCommandFileCompletion(t *testing.T) {
	tests := []struct {
		name       string
		input      []rune
		wantResult bool
	}{
		{
			name:       "command at start of file, should pass",
			input:      []rune{'c', 'a', 't', ' ', 'c'},
			wantResult: true,
		},
		{
			name:       "command with start of file, should pass",
			input:      []rune{'c', ' ', 'a'},
			wantResult: true,
		},
		{
			name:       "space at start",
			input:      []rune{' ', 'c', ' '},
			wantResult: false,
		},
		{
			name:       "space at start",
			input:      []rune{' ', 'c', ' '},
			wantResult: false,
		},
		{
			name:       "space at start",
			input:      []rune{' ', ' '},
			wantResult: false,
		},
	}

	for i, tt := range tests {
		got := commandAlreadyExists(tt.input)
		if got != tt.wantResult {
			t.Errorf("test %d failed wanted:%v got: %v", i, tt.wantResult, got)
		}
	}
}
