package shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type interceptor struct {
	writer   io.Writer
	lastByte byte
}

type Shell struct {
	Builtins    map[string]func(stage stage) (result, error) //command and associate functions
	Paths       []string                                     // used to access PATH folders
	WorkingDir  string                                       // keeps track of current working directory
	reader      io.Reader
	writer      io.Writer
	interceptor *interceptor
}

// a pipeline is a slice of commands
type pipeline struct {
	stages []stage
}

// defines the stage in a command that contains the command and its input or output
type stage struct {
	name               string
	args               []string
	stdOutFile         string
	stdErrorFile       string
	appendStdErrorFile string
	output             map[int]func(st stage, res result) error // 0: normal output to stdout 1: > stdout to file 2: > stderror to file
	outputfunc         func(st stage, res result) error
}

// result represents the output from executeCommand, contains both stderror and stdout in bytes as well as exit code
type result struct {
	stdOut   bytes.Buffer
	stdError bytes.Buffer
	exitCode int
}

func (sh *Shell) noRedirection(s stage, res result) error {
	_, err := sh.writer.Write(res.stdOut.Bytes())
	return err
}

// pipe out > & 1> should redirect stdout but write out stderror
func (sh *Shell) redirectStdOut(s stage, res result) error {
	err := createPathAndFile(res.stdOut.Bytes(), s.stdOutFile)
	if err != nil {
		return err
	}

	if res.stdError.Len() != 0 {
		sh.writer.Write(res.stdError.Bytes())
	}
	return nil
}

// pipe content out 2> should redirect stderr but not stdout
func (sh *Shell) redirectStdError(s stage, res result) error {

	if res.stdOut.Len() > 0 {
		sh.writer.Write(res.stdOut.Bytes())
	}

	err := createPathAndFile(res.stdError.Bytes(), s.stdErrorFile)
	if err != nil {
		return err
	}

	return nil
}

// interceptor wraps io writer and tracks last byte written its purpose is to detect if a newline should be inserted
func (i *interceptor) Write(p []byte) (int, error) {
	for idx, b := range p {
		if idx == len(p)-1 {
			i.lastByte = b
		}
	}
	return i.writer.Write(p)
}

func (i *interceptor) LastByteNewLine() bool {
	return i.lastByte == '\n'
}

// fetches all directories in PATH
func (sh *Shell) ParsePath(str string) {
	seperator := os.PathListSeparator
	folders := strings.SplitSeq(str, string(seperator))
	for f := range folders {
		sh.Paths = append(sh.Paths, f)
	}
}

// creates a new shell and returns it for the caller
func New(r io.Reader, w io.Writer) *Shell {
	inter := interceptor{
		writer:   w,
		lastByte: '\n',
	}

	s := Shell{
		Builtins:    make(map[string]func(s stage) (result, error)),
		Paths:       make([]string, 0),
		reader:      r,
		writer:      &inter,
		interceptor: &inter,
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

// run runs the shell loop
func (sh *Shell) Run() {
	scanner := bufio.NewScanner(sh.reader)

	// keep looping and getting input
	for {
		if sh.interceptor.LastByteNewLine() {
			fmt.Fprint(sh.writer, "$ ")
		} else {
			fmt.Fprintf(sh.writer, "\n")
			fmt.Fprint(sh.writer, "$ ")
		}

		if !scanner.Scan() {
			break
		}

		sh.interceptor.lastByte = '\n'

		sh.processLine(scanner.Text())

	}

}

func (sh *Shell) processLine(line string) {
	tokens := ParseLine(line)
	pipeline, err := createPipeline(tokens, *sh)
	if err != nil {
		fmt.Fprintln(sh.writer, err)
		return
	}

	err = sh.processPipeline(pipeline)
	if err != nil {
		fmt.Fprintln(sh.writer, err)
		return
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

// executeCommandResult takes the command and the args and calls the operating system to execute
func (sh *Shell) executeCommandResult(stage stage) (result, error) {
	var res result

	execCmd := exec.Command(stage.name, stage.args...)
	execCmd.Stdout = &res.stdOut
	execCmd.Stderr = &res.stdError

	err := execCmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			res.exitCode = execCmd.ProcessState.ExitCode()
			return res, fmt.Errorf("error running %s: %v", stage.name, err)
		}
	}

	return res, nil
}

// parse takes input from the terminal and processes the tokens splitting on spaces, double quotes
// and single quotes. Each has distinct rules that dictate how they should be split. See unit tests for more details
func ParseLine(input string) []string {
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

// createPipeline composes stages for each command to parse
func createPipeline(tokens []string, sh Shell) (pipeline, error) {
	pipeline := pipeline{}
	stage := stage{
		output: make(map[int]func(stage, result) error),
	}

	stage.output[0] = sh.noRedirection
	stage.output[1] = sh.redirectStdOut
	stage.output[2] = sh.redirectStdError

	// set the default outputfunc to nodirection
	stage.outputfunc = sh.noRedirection

	if len(tokens) == 0 {
		return pipeline, errors.New("input length is zero")
	}

	redirectStdOut := false   // support for > & 1>
	redirectStdError := false // support for 2>

	for i, arg := range tokens {
		if i == 0 { // first arg is cmd
			stage.name = tokens[i]
		} else {
			if redirectStdOut {
				f, _ := stage.output[1]
				stage.outputfunc = f
				stage.stdOutFile = tokens[i]
				continue
			}
			if redirectStdError {
				f, _ := stage.output[2]
				stage.outputfunc = f
				stage.stdErrorFile = tokens[i]
				continue
			} else {
				f, _ := stage.output[0]
				stage.outputfunc = f
			}

			switch arg {
			case ">", "1>":
				redirectStdOut = true
			case "2>":
				redirectStdError = true
			default:
				stage.args = append(stage.args, arg)
			}

		}
	}
	pipeline.stages = append(pipeline.stages, stage)

	return pipeline, nil
}

// process pipeline works through each stage and executes the functions
func (sh *Shell) processPipeline(pipeline pipeline) error {
	for _, stage := range pipeline.stages {
		f, ok := sh.Builtins[stage.name]
		if !ok {
			path := sh.findExePath(stage.name)
			if path == "" {
				// command not found and is not builtin return err and start from scratch
				return fmt.Errorf("%s: command not found", stage.name)
			}

			// return a result type that contains stderr and output
			result, err := sh.executeCommandResult(stage) // exec command
			if err != nil {
				return err
			}

			// execute output func
			err = stage.outputfunc(stage, result)
			if err != nil {
				return err
			}

			return nil
		}

		// execute shell built in command
		res, err := f(stage)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		err = stage.outputfunc(stage, res)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}

// helper function that creates directory
func createPathAndFile(data []byte, path string) error {

	// check and see if the directory from the path exists
	folderPath := filepath.Dir(path)
	_, err := os.Stat(folderPath)
	if err != nil {
		// if the folder does not exist create it
		if os.IsNotExist(err) {
			err = os.MkdirAll(folderPath, 0755)
			if err != nil {
				return err
			}
		}
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
