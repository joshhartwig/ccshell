package completion

import (
	"io/fs"
	"os"
)

func List(prefix string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	root := os.DirFS(cwd)
	files, err := fs.Glob(root, "*.*")
	return files, nil
}
