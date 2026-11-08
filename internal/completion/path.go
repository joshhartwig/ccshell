package completion

import (
	"os"
	"path/filepath"
	"strings"
)

func CompleteFileName(path string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir, file := filepath.Split(path)
	dir = strings.TrimPrefix(dir, string(filepath.Separator))
	searchDir := filepath.Join(cwd, dir)

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return []string{}, err
	}

	trie := NewTrie()
	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			name += "/"
		}

		trie.Insert(name)
	}

	return trie.Prefixes(file), nil
}
