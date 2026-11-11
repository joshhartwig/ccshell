package keylistener

import (
	"bytes"
	"strings"
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
		got := commandExists(tt.input)
		if got != tt.wantResult {
			t.Errorf("test %d failed wanted:%v got: %v", i, tt.wantResult, got)
		}
	}
}

func TestTabCompletesSingleCommand(t *testing.T) {
	cmds := func(p string) []string { return filterPrefix([]string{"echo", "exit"}, p) }
	files := func(string) ([]string, error) { return nil, nil }
	var out bytes.Buffer
	ks := NewKeyStrokeListener(strings.NewReader("ec\t\r"), &out, cmds, files)

	got, err := ks.HandleInput("$ ")
	if err != nil {
		t.Error(err)
	}
	want := "echo "

	if got != want {
		t.Errorf("got:%s want:%s", got, want)
	}
}

func filterPrefix(words []string, prefix string) []string {
	var out []string
	for _, w := range words {
		if strings.HasPrefix(w, prefix) {
			out = append(out, w)
		}
	}
	return out
}
