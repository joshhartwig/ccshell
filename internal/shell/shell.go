package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Shell struct {
	Builtins   map[string]func(s ...string) error //command and associate functions
	Paths      []string                           // used to access PATH folders
	WorkingDir string                             // keeps track of current working directory
	reader     io.Reader
	writer     io.Writer
}

// fetches all directories in PATH
func (sh *Shell) ParsePath(str string) {
	seperator := os.PathListSeparator
	folders := strings.SplitSeq(str, string(seperator))
	for f := range folders {
		sh.Paths = append(sh.Paths, f)
	}
}

func New(r io.Reader, w io.Writer) *Shell {
	s := Shell{
		Builtins: make(map[string]func(s ...string) error),
		Paths:    make([]string, 0),
		reader:   r,
		writer:   w,
	}

	s.registerBuiltins()
	wd, _ := os.Getwd()
	s.WorkingDir = wd

	path := os.Getenv("PATH")
	s.ParsePath(path)

	return &s
}

// returns true is file is executable
func isExec(mode os.FileMode) bool {
	return mode&0111 == 0111
}

// registers commands for the shell
func (sh *Shell) registerBuiltins() {
	sh.Builtins["exit"] = sh.exitHandler
	sh.Builtins["pwd"] = sh.pwdHandler
	sh.Builtins["echo"] = sh.echoHandler
	sh.Builtins["type"] = sh.typeHandler
	sh.Builtins["cd"] = sh.revisedChangeDirHandler
}

func (sh *Shell) ProcessInput() {

	scanner := bufio.NewScanner(sh.reader)

	// keep looping and getting input
	for {
		fmt.Fprint(sh.writer, "$ ")
		if !scanner.Scan() {
			break
		}

		// split cmd, args if both space if found between words
		cmd, args, found := strings.Cut(scanner.Text(), " ") // split cmd and args
		if !found {
			cmd = scanner.Text() // single cmd found
		}

		splitArgs := strings.Split(args, " ")

		f, ok := sh.Builtins[cmd]
		if !ok {
			path := sh.findExePath(cmd) // determine if cmd is in PATH
			if path == "" {
				fmt.Fprintf(sh.writer, "%s: command not found\n", scanner.Text()) // command not found and is not builtin return err and start from scratch
				continue
			}

			err := sh.executeCommand(cmd, splitArgs...) // exec command
			if err != nil {
				fmt.Fprintf(sh.writer, "error running %s: %v\n", cmd, err)
			}
			continue
		}

		// if found run the command
		err := f(args)
		if err != nil {
			fmt.Fprintf(sh.writer, "%v\n", err)
		}
	}
}

// looks for an executable in the PATH and returns the path of the EXE
// example return for 'ls' is `/bin/ls`
func (sh *Shell) findExePath(cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return ""
	}
	return path
}

func (sh *Shell) executeCommand(cmd string, args ...string) error {

	var execCmd *exec.Cmd
	if args[0] == "" {
		execCmd = exec.Command(cmd)
	} else {
		execCmd = exec.Command(cmd, args...)
	}

	bytes, err := execCmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Fprint(sh.writer, string(bytes))
	return nil
}
