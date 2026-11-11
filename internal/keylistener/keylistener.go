package keylistener

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	in               *bufio.Reader
	raw              *os.File
	out              io.Writer
	commandCompleter func(prefix string) []string          // used for auto complete on commands
	fileCompleter    func(prefix string) ([]string, error) // used for auto complete on files
}

type keyState struct {
	lastKeyTab   bool
	lastKeySpace bool
}

// NewKeyStrokeListener creates a new keystrokeListener for shell required for autocomplete support
func NewKeyStrokeListener(r io.Reader, w io.Writer, cc func(prefix string) []string, fc func(prefix string) ([]string, error)) *KeyStrokeListener {
	ks := &KeyStrokeListener{
		in:               bufio.NewReader(r),
		out:              w,
		commandCompleter: cc,
		fileCompleter:    fc,
	}

	if f, ok := r.(*os.File); ok {
		ks.raw = f
	}
	return ks
}

func (ks *KeyStrokeListener) HandleInput(prompt string) (string, error) {
	if ks.raw != nil && term.IsTerminal(int(ks.raw.Fd())) {
		old, err := term.MakeRaw(int(ks.raw.Fd()))
		if err != nil {
			return "", err
		}

		defer term.Restore(int(ks.raw.Fd()), old)
	}

	// init keystate for double tab tracking
	keyState := keyState{
		lastKeyTab:   false,
		lastKeySpace: false,
	}
	fmt.Fprint(ks.out, prompt) // print the prompt

	var buf []rune // tracks input
	reader := ks.in

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case keyEnter, keyLF:
			fmt.Fprint(ks.out, "\r\n")
			return string(buf), nil

		// tab completions
		case '\t':
			stringBuffer := string(buf)

			spaceIdx := strings.LastIndex(stringBuffer, " ")
			command := stringBuffer[:spaceIdx+1] // fetch text before space ex 'cat' <--
			partial := stringBuffer[spaceIdx+1:] // fetch text after space ex cat --> '/app/mai'

			folderPath, _ := filepath.Split(partial)

			// if we already have a command and tab is pressed, enter file completion mode
			if commandExists(buf) {
				files, err := ks.fileCompleter(partial) // fetch all files in the path
				if err != nil {
					break // unable to fetch folder path do nothing
				}

				slices.Sort(files) // sort found files

				switch len(files) {
				case 0: // no match play bell
					fmt.Fprintf(ks.out, "\a")
				case 1:
					if filepath.Ext(files[0]) != "" { // if the extension is present its file add a space
						files[0] = files[0] + " "
					}
					buf = []rune(command + folderPath + files[0]) // combine command + folderpath + files
					ks.Redraw(prompt, buf)
				default: // multiple files display them on the next line
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

			matches := ks.commandCompleter(stringBuffer)
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
func commandExists(buf []rune) bool {
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
