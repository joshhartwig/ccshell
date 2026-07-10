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
	sh.Builtins["cd"] = sh.changeDirHandler
}

func (sh *Shell) ProcessInput() {

	scanner := bufio.NewScanner(sh.reader)
	for { // keep looping and getting input
		fmt.Fprint(sh.writer, "$ ")
		if !scanner.Scan() {
			break
		}

		// split cmd, args if both space if found between words
		cmd, args, found := strings.Cut(scanner.Text(), " ") // split cmd and args
		if !found {
			cmd = scanner.Text() // single cmd found
		}

		// check if the args are quoted before splitting
		var splitArgs = parse(args)
		// if strings.HasPrefix(args, "'") && strings.HasSuffix(args, "'") { // if the args do not start with single quote, split with space
		// 	pre, _ := strings.CutPrefix(args, "'")
		// 	post, _ := strings.CutSuffix(pre, "'")
		// 	splitArgs = append(splitArgs, post)
		// } else {
		// 	splitArgs = strings.Split(args, " ") // split the args into an array
		// }

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
		err := f(splitArgs...)
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

func parse(input string) []string {
	const (
		modeUnquoted = iota
		modeSingleQuoted
	)
	const spaceRune rune = ' '
	const singleQuoteRune rune = '\''

	mode := modeUnquoted
	spaceDelim := false
	tokens := []string{}
	token := []rune{}

	for _, c := range input {
		if mode == modeSingleQuoted { // in single quote mode
			if c == singleQuoteRune { // last singleq
				mode = modeUnquoted
				continue
			}
			token = append(token, c)
			continue
		}
		if c == singleQuoteRune { // starting singleq
			mode = modeSingleQuoted
			continue
		}

		if c == spaceRune {
			if spaceDelim {
				tokens = append(tokens, string(token)) // append word
				token = token[:0]                      // reset word
				spaceDelim = false
				continue
			}
			continue
		}
		token = append(token, c)
		spaceDelim = true // ignore space until we hit
	}

	if len(token) > 0 {
		tokens = append(tokens, string(token))
	}

	return tokens
}
