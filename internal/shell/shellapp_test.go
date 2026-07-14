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
		{
			name:  "shell example",
			input: "'shell example'",
			want:  []string{"shell example"},
		},
		{
			name:  "echo",
			input: "echo test     shell",
			want:  []string{"echo", "test", "shell"},
		},
		{
			name:  "start with quote",
			input: "echo 'hello world' hello",
			want:  []string{"echo", "hello world", "hello"},
		},
		{
			name:  "mixed spaces and single quotes",
			input: "echo 'test     hello' 'world''shell' example''script",
			want:  []string{"echo", "test     hello", "worldshell", "examplescript"},
		},

		// double quote tests
		{
			name:  "double quotes with multiple spaces",
			input: `echo "hello    world"`,
			want:  []string{"echo", "hello    world"},
		},
		{
			name:  "adjacent double quoted strings are concatenated",
			input: `echo "hello""world"`,
			want:  []string{"echo", "helloworld"},
		},
		{
			name:  "double quoted and unquoted strings are concatenated",
			input: `echo "hello"world`,
			want:  []string{"echo", "helloworld"},
		},
		{
			name:  "double quoted strings separated by space",
			input: `echo "hello" "world"`,
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "single quotes inside double quotes are literal",
			input: `echo "shell's test"`,
			want:  []string{"echo", "shell's test"},
		},
		{
			name:  "escaping spaces",
			input: `echo three\ \ \ spaces`,
			want:  []string{"echo", "three   spaces"},
		},
		{
			name:  "escaping spaces more complex",
			input: `echo before\     after`,
			want:  []string{"echo", "before ", "after"},
		},
		{
			name:  "backslash-n becomes just n",
			input: `echo test\nexample`,
			want:  []string{"echo", "testnexample"},
		},
		{
			name:  "escaped backslash is literal",
			input: `echo hello\\world`,
			want:  []string{"echo", "hello\\world"},
		},
		{
			name:  "escaped single quotes are literal",
			input: `echo \'hello\'`,
			want:  []string{"echo", "'hello'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse(tt.input)
			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("wanted:%v len:%d got:%v len:%d", tt.want, len(tt.want), got, len(got))
			}
		})
	}
}
