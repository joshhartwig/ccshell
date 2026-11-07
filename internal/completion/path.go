package completion

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CompleteFileName(path string) ([]string, error) {
	var dir, file, cwd string

	cwd, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	dir = filepath.Dir(filepath.Clean(cwd + path))
	_, file = filepath.Split(path)

	files, err := fs.Glob(os.DirFS(dir), "*")
	if err != nil {
		return []string{}, err
	}

	trie := NewTrie()
	for _, f := range files {
		if filepath.Ext(f) == "" {
			f += "/"
		}
		trie.Insert(f)
	}

	// result := trie.Prefixes(file)
	// if len(result) == 0 {
	// 	return trie.Prefixes(""), nil
	// }

	return trie.Prefixes(file), nil
}
