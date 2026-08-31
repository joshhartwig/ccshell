package completions

import (
	"fmt"
	"testing"
)

// creates trie and tests if the words match
func TestInsert(t *testing.T) {

	tests := []struct {
		word       string
		shouldPass bool
	}{{
		word:       "cat",
		shouldPass: true,
	},
		{
			word:       "cog",
			shouldPass: true,
		},
		{
			word:       "dog",
			shouldPass: true,
		},
		{
			word:       "ham",
			shouldPass: false,
		},
	}

	tr := NewTrie()

	// add each word if it should pass the test
	for _, tt := range tests {
		if tt.shouldPass {
			tr.Insert(tt.word)
		}
	}

	for _, tt := range tests {
		if tt.shouldPass {
			if !tr.Search(tt.word) {
				t.Errorf("should have found word %s", tt.word)
			}
		}
	}
}

func TestPrefixes(t *testing.T) {
	words := []string{"cat", "cog", "dog", "deck", "deep", "canvas", "canva"}

	trie := NewTrie()

	for _, w := range words {
		trie.Insert(w)
	}

	results := trie.Prefixes("ca")
	fmt.Println(results)
}
