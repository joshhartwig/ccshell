package completion

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CompleteFileName(path string) ([]string, error) {
	var dir, prefix, cwd string

	cwd, _ = os.Getwd()
	dir = filepath.Dir(filepath.Clean(cwd + path))
	prefix = filepath.Base(path)
	// split on prefix after last /
	prefix = path[strings.LastIndex(path, "/")+1:]

	files, err := fs.Glob(os.DirFS(dir), "*")
	if err != nil {
		return []string{}, err
	}

	trie := NewTrie()
	for _, f := range files {
		trie.Insert(f)

	}
	return trie.Prefixes(prefix), nil
}
