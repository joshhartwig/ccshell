package shell

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	input := new(bytes.Buffer)
	input.WriteString("type exit\n")

	output := new(bytes.Buffer)
	shell := New(input, output)
	shell.ProcessInput()
	fmt.Println(output.String())
}
