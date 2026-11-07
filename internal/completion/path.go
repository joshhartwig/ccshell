package completion

import (
	"io/fs"
	"os"
	"strings"
)

// TODO: this currently works with paths but only returns the file name not the original path + filename
// also issue with returning just the Trie results example cat /app/ma returns main.go
func CompleteFileName(path string) ([]string, error) {
	var dir, prefix string

	// if we do not have a / then set the directory to '.'
	if !strings.Contains(path, "/") {
		prefix = path
		dir = "."
	} else {
		// fetch everything prior to the last /
		dir = path[:strings.LastIndex(path, "/")]
	}

	// split on prefix after last /
	prefix = path[strings.LastIndex(path, "/")+1:]
	if prefix == "" {
		prefix = path
	}

	files, err := fs.Glob(os.DirFS("."+dir), "*")
	if err != nil {
		return []string{}, err
	}

	trie := NewTrie()
	for _, f := range files {
		trie.Insert(f)

	}
	return trie.Prefixes(prefix), nil
}
