package shell

import (
	"bytes"
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

		want := "$ cat: nonexistent: No such file or directory\n"
		got, err := output.ReadString('\n')
		if err != nil {
			t.Errorf("error processing output: %v", err)
		}
		if got != want {
			t.Errorf("want: %s got: %s", want, got)
		}
	})
}
