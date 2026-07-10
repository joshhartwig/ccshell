package shell

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single string with spaces",
			input: "abc abc",
			want:  []string{"abc", "abc"},
		},
		{
			name:  "multiple spaces between args",
			input: "hello    world",
			want:  []string{"hello", "world"},
		},
		{
			name:  "leading and trailing spaces",
			input: "  foo bar  ",
			want:  []string{"foo", "bar"},
		},
		{
			name:  "single argument no spaces",
			input: "single",
			want:  []string{"single"},
		},
		{
			name:  "quoted string with spaces",
			input: "'hello    world'",
			want:  []string{"hello    world"},
		},
		{
			name:  "empty quotes is ignored",
			input: "hello''world",
			want:  []string{"helloworld"},
		},
		{
			name:  "adjacent quoted string are concatenated",
			input: "'hi''mom'",
			want:  []string{"himom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse(tt.input)
			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("wanted %v got %v", tt.want, got)
			}
		})
	}
}
