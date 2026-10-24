package completion

import (
	"io/fs"
	"os"
)

func CompleteFileName(filename string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	trie := NewTrie()

	root := os.DirFS(cwd)
	files, err := fs.Glob(root, "*")
	if err != nil {
		return []string{}, err
	}

	for _, f := range files {
		trie.Insert(f)
	}
	return trie.Prefixes(filename), nil
}
