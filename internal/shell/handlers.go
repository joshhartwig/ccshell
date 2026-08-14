package shell

import (
	"bytes"
	"fmt"
	"os"
)

// exits the shell
func (sh *Shell) exitHandler(s stage) (result, error) {
	os.Exit(0)
	return result{}, nil
}

// prints current working directory
func (sh *Shell) pwdHandler(s stage) (result, error) {
	res := result{}

	wd, err := os.Getwd()
	if err != nil {
		return result{}, err
	}
	res.stdOut.Write([]byte(fmt.Sprintf("%s", wd)))
	return res, nil
}

// echos text provided to echo command
func (sh *Shell) echoHandler(s stage) (result, error) {
	// if stage outputs to a text file write everything here to buffer then to file
	// overwrite if append = false
	var buf bytes.Buffer
	for i, c := range s.args {
		if i == len(s.args)-1 {
			buf.Write([]byte(fmt.Sprintf("%s", c)))
		} else {
			buf.Write([]byte(fmt.Sprintf("%s ", c)))
		}
	}

	return result{stdOut: buf}, nil
}

// returns if a command is built in or the path of the command if not
func (sh *Shell) typeHandler(s stage) (result, error) {
	res := result{}
	for _, x := range s.args { // loop through args
		if x == "" {
			return res, fmt.Errorf("type requires one or more arguments") // stop empty type commands
		}

		_, found := sh.Builtins[x] // look for built in commands
		if !found {
			// look for executable in PATH

			path := sh.findExePath(x)
			if path == "" {
				return res, fmt.Errorf("%s: not found", x)
			}
			res.stdOut.Write([]byte(fmt.Sprintf("%s is %s", x, path)))
			return res, nil
		}
		res.stdOut.Write([]byte(fmt.Sprintf("%s is a shell builtin", x)))
	}
	return res, nil
}

func (sh *Shell) changeDirHandler(s stage) (result, error) {
	if len(s.args) == 0 {
		return result{}, fmt.Errorf("cd requires at least one argument")
	}
	path := s.args[0]

	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return result{}, err
		}

		err = os.Chdir(home)
		return result{}, nil
	}

	err := os.Chdir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result{}, fmt.Errorf("cd: %s: No such file or directory", path)
		}
	}
	return result{}, nil
}
