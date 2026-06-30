package shell

import (
	"fmt"
	"os"
)

// exits the shell
func (sh *Shell) exitHandler(s ...string) error {
	os.Exit(0)
	return nil
}

// prints current working directory
func (sh *Shell) pwdHandler(s ...string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(sh.writer, "%s\n", wd)
	return nil
}

// echos text provided to echo command
func (sh *Shell) echoHandler(args ...string) error {
	for _, c := range args {
		fmt.Fprintf(sh.writer, "%s", c)
	}
	fmt.Fprintln(sh.writer, "")
	return nil
}

// returns if a command is built in or the path of the command if not
func (sh *Shell) typeHandler(s ...string) error {
	for _, x := range s { // loop through args
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

func (sh *Shell) changeDirHandler(s ...string) error {
	path := s[0]

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
