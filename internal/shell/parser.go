package shell

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
