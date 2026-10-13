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
	keyCtrlC     = 3
	keyCtrlD     = 4
	keyCtrlU     = 21
	keyEnter     = '\r'
	keyLF        = '\n'
	keyEscape    = 27
	keySpace     = ' '
	keyBackspace = 127
	keyUnknown   = 0xd800 /* UTF-16 surrogate area */ + iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyAltLeft
	keyAltRight
	keyHome
	keyEnd
	keyDeleteWord
	keyDeleteLine
	keyDelete
	keyClearScreen
	keyTranspose
	keyPasteStart
	keyPasteEnd
)

type KeyStrokeListener struct {
	in               *os.File
	out              io.Writer
	commandCompleter func(prefix string) []string
	fileCompleter    func(prefix string) []string
}

type keyState struct {
	lastKeyTab   bool
	lastKeySpace bool
}

type context struct {
	line        string
	currentWord string
	index       int
}

// NewKeyStrokeListener creates a new keystrokeListener for shell required for autocomplete support
func NewKeyStrokeListener(r *os.File, w io.Writer, f func(prefix string) []string) *KeyStrokeListener {
	return &KeyStrokeListener{
		in:               r,
		out:              w,
		commandCompleter: f,
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
			// TODO: if command set with space ex `cat te` -> `cat test.txt`
			// check if command is set if len greater 2 w/ space

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
