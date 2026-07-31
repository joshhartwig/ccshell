package shell

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// exits the shell
func (sh *Shell) exitHandler(s stage) error {
	os.Exit(0)
	return nil
}

// prints current working directory
func (sh *Shell) pwdHandler(s stage) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(sh.writer, "%s\n", wd)
	return nil
}

// echos text provided to echo command
func (sh *Shell) echoHandler(s stage) error {
	// if stage outputs to a text file write everything here to buffer then to file
	// overwrite if append = false
	var buf bytes.Buffer
	for i, c := range s.args {
		if i == len(s.args)-1 {
			//fmt.Fprintf(sh.writer, "%s", c) // print final no space
			buf.Write([]byte(fmt.Sprintf("%s", c)))
		} else {
			//fmt.Fprintf(sh.writer, "%s ", c) // print w/ space
			buf.Write([]byte(fmt.Sprintf("%s ", c)))
		}
	}

	// if output file is not empty get cwd
	if s.stdOutFile != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		filename := s.stdOutFile
		filepath := filepath.Join(cwd, filename)
		err = os.WriteFile(filepath, buf.Bytes(), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	sh.writer.Write(buf.Bytes())
	fmt.Fprintln(sh.writer, "")
	return nil
}

// returns if a command is built in or the path of the command if not
func (sh *Shell) typeHandler(s stage) error {
	for _, x := range s.args { // loop through args
		if x == "" {
			return fmt.Errorf("type requires one or more arguments") // stop empty type commands
		}

		_, found := sh.Builtins[x] // look for built in commands
		if !found {
			// look for executable in PATH

			path := sh.findExePath(x)
			if path == "" {
				return fmt.Errorf("%s: not found", x)
			}
			fmt.Fprintf(sh.writer, "%s is %s\n", x, path)
			return nil
		}
		fmt.Fprintf(sh.writer, "%s is a shell builtin\n", x)
	}
	return nil
}

func (sh *Shell) changeDirHandler(s stage) error {
	if len(s.args) == 0 {
		return fmt.Errorf("cd requires at least one argument")
	}
	path := s.args[0]

	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		err = os.Chdir(home)
		return nil
	}

	err := os.Chdir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cd: %s: No such file or directory", path)
		}
	}
	return nil
}
