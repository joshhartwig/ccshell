package shell

import (
	"bufio"
	"errors"
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

type pipeline struct {
	commands []command
}

type command struct {
	name string
	args []string
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

	// keep looping and getting input
	for {
		fmt.Fprint(sh.writer, "$ ")
		if !scanner.Scan() {
			break
		}

		// fetch input and parse into tokens
		line := scanner.Text()
		tokens := parse(line)

		// TODO: convert []tokens to Command struct check if err creating
		// struct. If len(0)
		if len(tokens) == 0 { // if empty continue
			continue
		}

		f, ok := sh.Builtins[tokens[0]]
		if !ok {
			// if not found determine if the command is in PATH
			path := sh.findExePath(tokens[0])
			if path == "" {
				fmt.Fprintf(sh.writer, "%s: command not found\n", scanner.Text()) // command not found and is not builtin return err and start from scratch
				continue
			}

			err := sh.executeCommand(tokens[0], tokens[1:]...) // exec command
			if err != nil {
				fmt.Fprintf(sh.writer, "error running %s: %v\n", tokens[0], err)
			}
			continue
		}

		// if found run the command
		err := f(tokens[1:]...)
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

// add backslash support makes next char literal
// /<space> adds space and does not split
// echo hello\\world
func parse(input string) []string {
	const (
		modeUnquoted = iota
		modeSingleQuoted
		modeDoubleQuoted
	)
	const spaceRune rune = ' '
	const singleQuoteRune rune = '\''
	const doubleQuotedRune rune = '"'
	const backslashRune rune = '\\'
	const dollarRune rune = '$'
	const asteriskRune rune = '*'
	const newlineRune rune = '\n'

	// this needs to be outside of iota as escaping can happen
	// inside double quoted text
	escapeMode := false

	mode := modeUnquoted
	spaceDelim := false
	tokens := []string{}
	token := []rune{}

	for _, c := range input {

		// in single quote mode
		if mode == modeSingleQuoted {
			// if last single quote
			if c == singleQuoteRune {
				mode = modeUnquoted
				continue
			}
			token = append(token, c)
			spaceDelim = true
			continue
		}

		// in double quoted mode
		if mode == modeDoubleQuoted {
			// last double quote & reset to unquoted
			if escapeMode {
				if c == doubleQuotedRune || c == singleQuoteRune || c == backslashRune || c == dollarRune || c == asteriskRune || c == newlineRune {
					token = append(token, c)
					escapeMode = false
					continue
				}
			}
			if c == doubleQuotedRune {
				mode = modeUnquoted
				continue
			}

			if c == backslashRune && escapeMode == false {
				escapeMode = true
				continue
			}
			token = append(token, c)
			spaceDelim = true
			escapeMode = false
			continue

		}

		// if we hit an escape character and not in escape mode
		// set to true
		if c == backslashRune && escapeMode == false {
			escapeMode = true
			continue
		}

		// found first double quote
		if c == doubleQuotedRune && escapeMode == false {
			mode = modeDoubleQuoted
			continue
		}

		if c == singleQuoteRune && escapeMode == false { // starting singleq
			mode = modeSingleQuoted
			continue
		}

		if c == spaceRune && escapeMode == false {
			if spaceDelim {
				tokens = append(tokens, string(token)) // append word
				token = token[:0]                      // reset word
				spaceDelim = false
				continue
			}
			continue
		}

		// append the token to tokens
		token = append(token, c)
		// reset modes
		mode = modeUnquoted
		spaceDelim = true
		escapeMode = false
	}

	// append last token before return
	if len(token) > 0 {
		tokens = append(tokens, string(token))
	}

	return tokens
}

// creates a pipeline of commands to exec splitting on | or >
func createPipeline(input []string) (pipeline, error) {
	pipeline := pipeline{}
	cmd := command{}

	if len(input) == 0 {
		return pipeline, errors.New("input length is zero")
	}
	for i, arg := range input {
		if i == 0 { // first arg is cmd
			cmd.name = input[i]
		} else {
			if arg == ">" { // pipe found append command to pipeline
				pipeline.commands = append(pipeline.commands, cmd) // append current command
				cmd = command{                                     // create new command and add arg
					name: arg,
				}
			} else {
				cmd.args = append(cmd.args, arg)
			}
		}
	}
	pipeline.commands = append(pipeline.commands, cmd)

	return pipeline, nil
}
