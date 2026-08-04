package shell

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("test new shell and input type exit equals shell built in", func(t *testing.T) {
		input := new(bytes.Buffer)
		input.WriteString("type exit\n")

		output := new(bytes.Buffer)
		shell := New(input, output)
		shell.Run()

		want := "$ exit is a shell builtin\n"
		got, err := output.ReadString('\n')
		if err != nil {
			t.Errorf("error processing output: %v", err)
		}
		if got != want {
			t.Errorf("want: %s got: %s", want, got)
		}
	})
	t.Run("test cat no exist outputs error message", func(t *testing.T) {
		input := new(bytes.Buffer)
		input.WriteString("cat noexist > output.txt\n")

		output := new(bytes.Buffer)
		shell := New(input, output)
		shell.Run()

		want := "$ cat: noexist: No such file or directory\n\n$ "
		got := output.String()
		if got != want {
			t.Errorf("want: %s got: %s", want, got)
		}
	})
}

func TestNewLineInterceptor(t *testing.T) {
	t.Run("last byte is newline", func(t *testing.T) {
		testInput := []byte(string("hello world\n"))
		want := byte('\n')

		buf := new(bytes.Buffer)
		inter := interceptor{
			writer: buf,
		}

		_, err := inter.Write(testInput)
		if err != nil {
			t.Errorf("error writing to buffer")
		}

		got := inter.lastByte
		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("last byte is newline function", func(t *testing.T) {
		testInput := []byte(string("hello world\n"))

		buf := new(bytes.Buffer)
		inter := interceptor{
			writer: buf,
		}

		_, err := inter.Write(testInput)
		if err != nil {
			t.Errorf("error writing to buffer")
		}

		if !inter.LastByteNewLine() {
			t.Errorf("last byte should have been newline")
		}
	})

	t.Run("last byte is not newline", func(t *testing.T) {
		testInput := []byte(string("hello world"))

		buf := new(bytes.Buffer)
		inter := interceptor{
			writer: buf,
		}

		_, err := inter.Write(testInput)
		if err != nil {
			t.Errorf("error writing to buffer")
		}

		if inter.LastByteNewLine() {
			t.Errorf("last byte should not have been newline")
		}
	})

	t.Run("last byte is newline with empty string", func(t *testing.T) {
		testInput := []byte(string("\n"))

		buf := new(bytes.Buffer)
		inter := interceptor{
			writer: buf,
		}

		_, err := inter.Write(testInput)
		if err != nil {
			t.Errorf("error writing to buffer")
		}

		if !inter.LastByteNewLine() {
			t.Errorf("last byte should have been newline")
		}
	})

	t.Run("last byte is space not newline", func(t *testing.T) {
		testInput := []byte(string("hello world "))

		buf := new(bytes.Buffer)
		inter := interceptor{
			writer: buf,
		}

		_, err := inter.Write(testInput)
		if err != nil {
			t.Errorf("error writing to buffer")
		}

		if inter.LastByteNewLine() {
			t.Errorf("last byte should not have been newline")
		}
	})
}

func TestCreateDirectory(t *testing.T) {
	t.Run("test create directory absolute", func(t *testing.T) {
		t1 := []byte("hello world")
		path := "testdata/t1/test1.txt"
		err := createDirectory(t1, path)
		if err != nil {
			t.Fatalf("error creating directory: %v", err)
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("error reading created file: %v", err)
		}

		if string(t1) != string(bytes) {
			t.Errorf("got:%s want:%s", string(bytes), string(t1))
		}
	})
}

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
		{
			name:  "double quotes part 1",
			input: `echo "hello'test'\\'world"`,
			want:  []string{"echo", `hello'test'\'world`},
		},
		{
			name:  "escaped backslash in double quotes",
			input: `echo "A \\ escapes itself"`,
			want:  []string{"echo", `A \ escapes itself`},
		},
		{
			name:  "escaped double quote in double quotes",
			input: `echo "A \" inside double quotes"`,
			want:  []string{"echo", `A " inside double quotes`},
		},
		{
			name:  "single quotes and backslash in double quotes",
			input: `echo "just'one'\\n'backslash"`,
			want:  []string{"echo", `just'one'\n'backslash`},
		},
		{
			name:  "escaped double quote outside double quotes",
			input: `echo "inside\"literal_quote."outside\"`,
			want:  []string{"echo", `inside"literal_quote.outside"`},
		},
		{
			name:  "backslash and double quote in double quotes part 1",
			input: `echo "number 1"`,
			want:  []string{"echo", `number 1`},
		},
		{
			name:  "backslash and double quote in double quotes part 2",
			input: `echo "doublequote \" 2"`,
			want:  []string{"echo", `doublequote " 2`},
		},
		{
			name:  "backslash and double quote in double quotes part 3",
			input: `echo "backslash \\ 3"`,
			want:  []string{"echo", `backslash \ 3`},
		},
		{
			name:  "mixed escape and quotes",
			input: `echo "mixed\"quote'script'\\"`,
			want:  []string{"echo", `mixed"quote'script'\`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLine(tt.input)
			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("wanted:%v len:%d got:%v len:%d", tt.want, len(tt.want), got, len(got))
			}
		})
	}
}

func TestCreatePipeline(t *testing.T) {
	testInput := []string{"cat", "hello world", ">", "test.txt"}
	want := pipeline{
		stages: []stage{
			{name: "cat", args: []string{"hello", "world"}, stdOutFile: "test.txt", append: false},
		},
	}

	got, err := createPipeline(testInput)
	if err != nil {
		t.Errorf("got error: %v", err)
	}

	if len(got.stages) != len(want.stages) {
		t.Errorf("length should be the same got:%d want:%d", len(got.stages), len(want.stages))
		fmt.Println(got)
	}

	if got.stages[0].name != want.stages[0].name {
		t.Errorf("name should be the same got %s want %s", got.stages[0].name, want.stages[0].name)
	}

}
