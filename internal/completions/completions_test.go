package completions

import (
	"slices"
	"testing"
)

// creates trie and tests if the words match
func TestInsert(t *testing.T) {

	tests := []struct {
		word       string
		shouldPass bool
	}{
		{
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
		{
			word:       "ham_tree",
			shouldPass: true,
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

	want := []string{"cat", "canvas", "canva"}
	got := trie.Prefixes("ca")

	slices.Sort(want)
	slices.Sort(got)

	if !slices.Equal(want, got) {
		t.Errorf("slices should be equal got: %v want: %v", got, want)
	}
}
