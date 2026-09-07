package keylistener

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type KeyStrokeListener struct {
	in        *os.File
	out       io.Writer
	completer func(prefix string) []string
}

// NewKeyStrokeListener creates a new keystrokeListener for shell required for autocomplete support
func NewKeyStrokeListener(r *os.File, w io.Writer, f func(prefix string) []string) *KeyStrokeListener {
	return &KeyStrokeListener{
		in:        r,
		out:       w,
		completer: f,
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

	// print the prompt
	fmt.Fprint(ks.out, prompt)

	// read input 1 byte at a time
	var buf []rune
	reader := bufio.NewReader(ks.in)

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case '\r', '\n':
			fmt.Fprint(ks.out, "\r\n")
			return string(buf), nil

		case '\t': // fetch completions and iterate through them on each tab
			matches := ks.completer(string(buf))
			switch len(matches) {
			case 0: // do nothing
				fmt.Fprintf(ks.out, "\a")
			case 1:
				buf = []rune(matches[0]) // find the first relevant keyword
				buf = append(buf, ' ')   // add a space to the match
				ks.Redraw(prompt, buf)   // redraw the prompt + matched buffer + space
			default:
				fmt.Fprintf(ks.out, "\r\n%s\r\n", strings.Join(matches, " "))
				ks.Redraw(prompt, buf)
			}

		// support for backspace in terminal
		case 127, '\b':
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(ks.out, "\b \b")
			}
		case 3:
			return "", errors.New("interrupted")

		// append rune and print
		default:
			buf = append(buf, r)
			fmt.Fprintf(ks.out, "%c", r)
		}
	}
}

// Redraw draws the prompt and buffer input to io.writer
func (ks *KeyStrokeListener) Redraw(prompt string, buf []rune) {
	fmt.Fprintf(ks.out, "\r\033[K%s%s", prompt, string(buf))
}
