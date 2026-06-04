package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

type app struct {
	commands map[string]func(s ...string) error
}

func main() {
	// TODO: Uncomment the code below to pass the first stage
	var app app
	app.commands = make(map[string]func(s ...string) error)
	app.commands["exit"] = func(s ...string) error {
		os.Exit(0)
		return nil
	}
	app.commands["echo"] = func(s ...string) error {
		for _, x := range s {
			fmt.Printf("%s", x)
		}
		fmt.Print("\n")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		err := app.processInput(scanner.Text())
		if err != nil {
			fmt.Printf("%v", err)
		}
	}
}

// get input break commands and arg and
func (a *app) processInput(s string) error {
	cmd, args, found := strings.Cut(s, " ")
	if !found {
		cmd = s // only single command no args
	}

	// try and find the command in the map
	f, ok := a.commands[cmd]
	if !ok {
		return fmt.Errorf("%s: command not found\n", s)
	}

	// if found run the command
	err := f(args)
	if err != nil {
		return err
	}
	return nil
}
