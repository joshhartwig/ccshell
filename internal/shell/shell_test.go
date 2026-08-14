package shell

import (
	"bytes"
	"fmt"
	"os"
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

		want := "$ cat: noexist: No such file or directory\n$ "
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

func TestCreatePathAndFile(t *testing.T) {
	t.Run("test create directory absolute", func(t *testing.T) {
		t1 := []byte("hello world")
		path := "testdata/t1/test1.txt"
		err := createPathAndFile(t1, path)
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

func TestCreatePipeline(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer

	sh := New(&input, &output)
	testInput := []string{"cat", "hello world", ">", "test.txt"}
	wantPipeline := pipeline{
		stages: []stage{
			{name: "cat", args: []string{"hello", "world"}, stdOutFile: "test.txt"},
		},
	}

	wantOutput := "cat: hello world: No such file or directory\n"

	got, err := createPipeline(testInput, *sh)
	if err != nil {
		t.Errorf("got error: %v", err)
	}

	err = sh.processPipeline(got)
	if err != nil {
		t.Errorf("got error: %v", err)
	}

	if len(got.stages) != len(wantPipeline.stages) {
		t.Errorf("length should be the same got:%d want:%d", len(got.stages), len(wantPipeline.stages))
		fmt.Println(got)
	}

	if got.stages[0].name != wantPipeline.stages[0].name {
		t.Errorf("name should be the same got %s want %s", got.stages[0].name, wantPipeline.stages[0].name)
	}

	if output.String() != wantOutput {
		t.Errorf("wanted: %s got: %s", wantOutput, output.String())
	}

}
