package shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/internal/completions"
	"github.com/codecrafters-io/shell-starter-go/internal/keylistener"
)

type Shell struct {
	Builtins    map[string]func(stage stage) (result, error) //command and associate functions
	Paths       []string                                     // used to access PATH folders
	WorkingDir  string                                       // keeps track of current working directory
	reader      io.Reader
	writer      io.Writer
	interceptor *interceptor
	completer   *completions.Trie
}

// a pipeline is a slice of commands
type pipeline struct {
	stages []stage
}

// defines the stage in a command that contains the command and its input or output
type stage struct {
	name         string
	args         []string
	stdOutFile   string
	stdErrorFile string
	appendStdOut bool
	appendStdErr bool
	output       map[int]func(st stage, res result) error // 0: normal output to stdout 1: > stdout to file 2: > stderror to file
	outputfunc   func(st stage, res result) error
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
	if s.appendStdOut {
		if err := appendFile(res.stdOut.Bytes(), s.stdOutFile); err != nil {
			return err
		}
	} else {
		if err := createPathAndFile(res.stdOut.Bytes(), s.stdOutFile); err != nil {
			return err
		}
	}

	if res.stdError.Len() > 0 {
		if _, err := sh.writer.Write(res.stdError.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// pipe content out 2> should redirect stderr but not stdout
func (sh *Shell) redirectStdError(s stage, res result) error {
	if res.stdOut.Len() > 0 {
		if _, err := sh.writer.Write(res.stdOut.Bytes()); err != nil {
			return err
		}
	}

	if s.appendStdErr {
		return appendFile(res.stdError.Bytes(), s.stdErrorFile)
	}

	return createPathAndFile(res.stdError.Bytes(), s.stdErrorFile)
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
		completer:   completions.NewTrie(),
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

	sh.completer.Insert("exit")
	sh.completer.Insert("pwd")
	sh.completer.Insert("echo")
	sh.completer.Insert("type")
	sh.completer.Insert("cd")

	// adds all exes in path
	for _, e := range sh.findAllExeOSPath() {
		sh.completer.Insert(e)
	}
}

// test function for putting terminal in raw mode
func (sh *Shell) RunKeystrokeMode() {
	ks := keylistener.NewKeyStrokeListener(sh.reader.(*os.File), sh.writer, sh.completer.Prefixes)
	for {
		if !sh.interceptor.LastByteNewLine() {
			fmt.Fprintf(sh.writer, "\r\n")
		}

		str, err := ks.HandleInput("$ ")
		if err != nil {
			return
		}
		sh.processLine(str)
	}
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

// isExecAll returns true if the file is exectuable by all
func IsExecAll(mode os.FileMode) bool {
	return mode&0111 == 0111
}

// findAllExePath returns all exes in the current os.PATH
func (sh *Shell) findAllExeOSPath() []string {
	osPath := os.Getenv("PATH")
	var executables []string
	paths := strings.Split(osPath, string(filepath.ListSeparator))

	for _, p := range paths {
		entries, err := os.ReadDir(p)
		if err != nil {
			return executables
		}

		for _, entry := range entries {
			// Skip directories
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				// Skip files that error out when fetching info
				continue
			}

			if isExecutable(info) {
				executables = append(executables, entry.Name())
			}
		}
	}

	return executables
}

// isExecutable checks if a file is an executable depending on the Operating System.
func isExecutable(info fs.FileInfo) bool {
	if runtime.GOOS == "windows" {
		// On Windows, check for common executable file extensions
		ext := filepath.Ext(info.Name())
		return ext == ".exe" || ext == ".bat" || ext == ".cmd"
	}

	// On Unix/Linux/macOS, check if any executable bit is set (User, Group, or Other)
	return info.Mode().Perm()&0111 != 0
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

	redirectStdOut := false   //  > & 1>
	redirectStdError := false //  2>
	appendStdOut := false     // >>
	appendStdError := false   // support for 2>>

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
			}
			if appendStdOut {
				f, _ := stage.output[1]
				stage.outputfunc = f
				stage.stdOutFile = tokens[i]
				stage.appendStdOut = true
				continue
			}
			if appendStdError {
				f, _ := stage.output[2]
				stage.outputfunc = f
				stage.stdErrorFile = tokens[i]
				stage.appendStdErr = true
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
			case ">>", "1>>":
				appendStdOut = true
			case "2>>":
				appendStdError = true
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

// helper function that creates directory and file
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

// helper function that appends to existing file
func appendFile(data []byte, path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file closes when the function exits

	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}
