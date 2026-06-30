package shell

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("test new type exit", func(t *testing.T) {
		input := new(bytes.Buffer)
		input.WriteString("type exit\n")

		output := new(bytes.Buffer)
		shell := New(input, output)
		shell.ProcessInput()

		want := "$ exit is a shell builtin\n"
		got, err := output.ReadString('\n')
		if err != nil {
			t.Errorf("error processing output: %v", err)
		}

		if got != want {
			t.Errorf("want: %s got: %s", want, got)
		}
	})

	t.Run("test arg parsing", func(t *testing.T) {
		input := new(bytes.Buffer)
		input.WriteString(`echo 'world     hello' 'script''shell' `)

		output := new(bytes.Buffer)
		shell := New(input, output)
		shell.ProcessInput()

		want := "$ world     hello scriptshell exampletest"
		got, err := output.ReadString('\n')
		if err != nil {
			t.Errorf("error processing output: %v", err)
		}

		if got != want {
			t.Errorf("want: %s\n got: %s\n", want, got)
		}
	})
}
