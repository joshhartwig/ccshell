package keylistener

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"

	"github.com/codecrafters-io/shell-starter-go/internal/completion"
)

const (
	keyEnter     = '\r'
	keyLF        = '\n'
	keySpace     = ' '
	keyBackspace = 127
)

type KeyStrokeListener struct {
	in               *os.File
	out              io.Writer
	commandCompleter func(prefix string) []string
	fileCompleter    func(prefix string) ([]string, error)
}

type keyState struct {
	lastKeyTab   bool
	lastKeySpace bool
}

// NewKeyStrokeListener creates a new keystrokeListener for shell required for autocomplete support
func NewKeyStrokeListener(r *os.File, w io.Writer, cc func(prefix string) []string, fc func(prefix string) ([]string, error)) *KeyStrokeListener {
	return &KeyStrokeListener{
		in:               r,
		out:              w,
		commandCompleter: cc,
		fileCompleter:    fc,
	}
}

// HandleInput puts the terminal into 'raw' mode which is required to capture keystrokes that is needed for
// autocomplete support as part of codecrafters https://app.codecrafters.io/courses/shell/stages/qp2
// this function supports input and checks if tab, backspace, enter, crtl-c is inputted as well as return string
func (ks *KeyStrokeListener) HandleInput(prompt string) (string, error) {
	// set terminal into raw mode
	oldState, err := term.MakeRaw(int(ks.in.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(ks.in.Fd()), oldState)

	// init keystate for double tab tracking
	keyState := keyState{
		lastKeyTab:   false,
		lastKeySpace: false,
	}
	fmt.Fprint(ks.out, prompt) // print the prompt

	var buf []rune // tracks input
	reader := bufio.NewReader(ks.in)

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case keyEnter, keyLF:
			fmt.Fprint(ks.out, "\r\n")
			return string(buf), nil

		case '\t': // fetch completions and iterate through them on each tab
			if commandAlreadyExists(buf) {
				partial := string(buf[strings.LastIndex(string(buf), " ")+1:])
				command := string(buf[:strings.LastIndex(string(buf), " ")+1])
				files, _ := ks.fileCompleter(partial)
				slices.Sort(files)

				switch len(files) {
				case 0:
					fmt.Fprintf(ks.out, "\a") // bell no match
				case 1:
					buf = []rune(command + files[0]) // find the first relevant keyword
					buf = append(buf, ' ')           // add a space to the match
					ks.Redraw(prompt, buf)           // redraw the prompt + matched buffer + space
				default:
					if keyState.lastKeyTab {
						keyState.lastKeyTab = false
						fmt.Fprintf(ks.out, "\r\n%s\r\n", strings.Join(files, " "))
						ks.Redraw(prompt, buf)
					} else {
						keyState.lastKeyTab = true
						buf = []rune(command + completion.FindLongestCommonPrefix(files))
						fmt.Fprintf(ks.out, "\a")
						ks.Redraw(prompt, buf)
					}
				}
				break
			}
			matches := ks.commandCompleter(string(buf))
			slices.Sort(matches)

			switch len(matches) {
			case 0: // no matches play 'bell'
				fmt.Fprintf(ks.out, "\a")
			case 1:
				buf = []rune(matches[0]) // find the first relevant keyword
				buf = append(buf, ' ')   // add a space to the match
				ks.Redraw(prompt, buf)   // redraw the prompt + matched buffer + space
			default: // multiple matches
				if keyState.lastKeyTab {
					keyState.lastKeyTab = false
					fmt.Fprintf(ks.out, "\r\n%s\r\n", strings.Join(matches, " "))
					ks.Redraw(prompt, buf)
				} else {
					keyState.lastKeyTab = true
					buf = []rune(completion.FindLongestCommonPrefix(matches))
					fmt.Fprintf(ks.out, "\a")
					ks.Redraw(prompt, buf)
				}
			}

		// support for backspace in terminal
		case keyBackspace, '\b':
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(ks.out, "\b \b")
			}
		case 3:
			return "", errors.New("interrupted")

		// append rune and print
		default:
			buf = append(buf, r)
			keyState.lastKeyTab = false
			fmt.Fprintf(ks.out, "%c", r)
		}
	}
}

// Redraw draws the prompt and buffer input to io.writer
func (ks *KeyStrokeListener) Redraw(prompt string, buf []rune) {
	fmt.Fprintf(ks.out, "\r\033[K%s%s", prompt, string(buf))
}

// isCommandFileCompletion returns true if function determines we should be in file completion mode
// should have a command and a space then a rune
func commandAlreadyExists(buf []rune) bool {
	stringBuf := string(buf)
	// if we have a space
	if strings.Contains(stringBuf, " ") {
		// if the space is not the first char
		if stringBuf[0] != keySpace || stringBuf[len(stringBuf)-1] != keySpace {
			// check if first char after space is
			idx := strings.LastIndex(stringBuf, " ")
			if idx+1 <= len(stringBuf) {
				return true
			}
			return false
		}
		return false
	}
	return false
}
