package tokenizer

type Token struct {
	Value string
	Start int
	End   int
}

const (
	keyEnter     = '\r'
	keyLF        = '\n'
	keySpace     = ' '
	keyBackspace = 127
)

// func Tokenize(in *os.File) error {
// 	termState, err := term.MakeRaw(int(in.Fd()))
// 	if err != nil {
// 		return err
// 	}

// 	defer term.Restore(int(in.Fd()), termState)

// 	var buf []rune
// 	reader := bufio.NewReader(in)

// 	for {
// 		r, _, err := reader.ReadRune()
// 		if err != nil {
// 			return err
// 		}

// 		switch r {
// 		case keyEnter, keyLF:

// 		}
// 	}
// }
