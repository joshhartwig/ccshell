package shell

import (
	"bytes"
	"testing"
)

func TestHandlers(t *testing.T) {

	t.Run("type exit correctly shows builtin", func(t *testing.T) {
		in := new(bytes.Buffer)
		out := new(bytes.Buffer)
		shell := New(in, out)

		input := "type exit"
		shell.processLine(input)

		want := "exit is a shell builtin"
		got := out.String()

		if got != want {
			t.Errorf("got: %s wanted: %s", got, want)
		}
	})

	t.Run("type hello shows error", func(t *testing.T) {
		in := new(bytes.Buffer)
		out := new(bytes.Buffer)
		shell := New(in, out)

		input := "type hello"
		shell.processLine(input)

		want := "hello: not found\n"
		got := out.String()

		if got != want {
			t.Errorf("got: %s wanted: %s", got, want)
		}
	})

}
