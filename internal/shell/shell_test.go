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

		want := "$ cat: noexist: No such file or directory\n"
		got, err := output.ReadString('\n')
		if err != nil {
			t.Errorf("error processing output: %v", err)
		}
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
