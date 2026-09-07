package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/shell-starter-go/internal/shell"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	sh := shell.New(os.Stdin, os.Stdout)
	sh.RunKeystrokeMode()
}
