package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

type app struct {
	commands map[string]func(s ...string) error
	paths    []string
}

func main() {
	// TODO: Uncomment the code below to pass the first stage
	/*
			If the command is not a builtin, your shell must go through every directory in PATH. For each directory:
		Check if a file with the command name exists.
		Check if the file has execute permissions.
		If the file exists and has execute permissions, print <command> is <full_path> and stop.
		If the file exists but lacks execute permissions, skip it and continue to the next directory.
		If no executable is found in any directory, print <command>: not found.
	*/
	var app app
	app.commands = make(map[string]func(s ...string) error)
	app.registerCommands()

	path := os.Getenv("PATH")
	if path == "" {
		log.Panic("error path value empty")
	}

	app.parsePath(path)

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

func (a *app) registerCommands() {
	a.commands["exit"] = func(s ...string) error {
		os.Exit(0)
		return nil
	}
	a.commands["echo"] = func(s ...string) error {
		for _, x := range s {
			fmt.Printf("%s", x)
		}
		fmt.Print("\n")
		return nil
	}
	a.commands["type"] = func(s ...string) error {
		// iterate through s
		for _, x := range s {
			// look for built in commands
			_, ok := a.commands[x]
			if !ok {
				// look for executable
				found := a.findExecutables(x)
				if !found {
					return fmt.Errorf("%s: not found\n", x)
				}
				return nil
			}
			fmt.Printf("%s is a shell builtin\n", x)
		}
		return nil
	}
}

// fetches all directory in path
func (a *app) parsePath(s string) {
	seperator := os.PathListSeparator
	folders := strings.SplitSeq(s, string(seperator))
	for f := range folders {
		a.paths = append(a.paths, f)
	}
}

// loop through all files in the folder list, determine if the folder has execute permissions
// if it does write command is <full path> else skip
// ex.. $type grep => grep is /usr/bin/grep
func (a *app) findExecutables(cmd string) bool {
	for _, dir := range a.paths {
		direntry, err := os.ReadDir(dir)
		if err != nil {
			return false
		}

		for _, f := range direntry {
			if !f.IsDir() && f.Name() == cmd { // not directory & name matches cmd
				fileInfo, err := f.Info()
				if err != nil {
					return false
				}
				// is executable and matches command
				if IsExecAll(fileInfo.Mode()) {
					fmt.Printf("%s is %s/%s\n", cmd, dir, cmd)
					return true
				}
			}
		}
	}
	return false
}

func IsExecAll(mode os.FileMode) bool {
	return mode&0111 == 0111
}
