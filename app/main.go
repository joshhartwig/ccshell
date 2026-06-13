package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

type shellApplication struct {
	builtins map[string]func(s ...string) error //command and associate functions
	paths    []string                           // used to access PATH folders
}

func main() {
	var shellApp shellApplication
	shellApp.builtins = make(map[string]func(s ...string) error)
	shellApp.registerBuiltins()

	path := os.Getenv("PATH")
	if path == "" {
		log.Panic("error path value empty")
	}

	shellApp.parsePath(path) // parse PATH and store those in shellApp

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		err := shellApp.processInput(scanner.Text())
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

// get input break commands and arg and
func (a *shellApplication) processInput(s string) error {

	cmd, args, found := strings.Cut(s, " ") // split cmd and args
	if !found {
		cmd = s // single cmd found
	}

	f, ok := a.builtins[cmd]
	if !ok {
		path := a.findExePath(cmd) // determine if cmd is in PATH
		if path == "" {
			return fmt.Errorf("%s: command not found\n", s) // command not found and is not builtin return err and start from scratch
		}

		err := a.executeCommand(cmd, args) // exec command
		if err != nil {
			fmt.Printf("error running %s: %v\n", cmd, err)
		}
		return nil // return as we ran the command
	}

	// if found run the command
	err := f(args)
	if err != nil {
		return err
	}
	return nil
}

func (a *shellApplication) registerBuiltins() {
	a.builtins["exit"] = func(s ...string) error {
		os.Exit(0)
		return nil
	}
	a.builtins["echo"] = func(s ...string) error {
		for _, x := range s {
			fmt.Printf("%s", x)
		}
		fmt.Print("\n")
		return nil
	}
	a.builtins["type"] = func(s ...string) error {
		for _, x := range s { // loop through args
			if x == "" {
				return fmt.Errorf("type requires one or more arguments") // stop empty type commands
			}

			_, found := a.builtins[x] // look for built in commands
			if !found {
				// look for executable in PATH

				path := a.findExePath(x)
				if path == "" {
					return fmt.Errorf("%s: not found\n", x)
				}
				fmt.Printf("%s is %s\n", x, path)
				return nil
			}
			fmt.Printf("%s is a shell builtin\n", x)
		}
		return nil
	}
}

// fetches all directories in PATH
func (a *shellApplication) parsePath(s string) {
	seperator := os.PathListSeparator
	folders := strings.SplitSeq(s, string(seperator))
	for f := range folders {
		a.paths = append(a.paths, f)
	}
}

// looks for an executable in the PATH and returns the path of the EXE
// example return for 'ls' is `/bin/ls`
func (a *shellApplication) findExePath(cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return ""
	}
	return path
}

func (a *shellApplication) executeCommand(cmd, args string) error {
	execCmd := exec.Command(cmd, args)
	bytes, err := execCmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Print(string(bytes))

	return nil
}

func IsExecAll(mode os.FileMode) bool {
	return mode&0111 == 0111
}
