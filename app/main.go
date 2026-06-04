package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	commands := make(map[string]func(s ...string) error)
	commands["exit"] = func(s ...string) error { return nil }

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		cmd := scanner.Text()
		if cmd == "exit" {
			return
		} else {
			fmt.Printf("%s : command not found\n", cmd)
		}
	}
}
