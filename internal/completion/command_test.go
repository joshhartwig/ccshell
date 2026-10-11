package completion

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

func TestFindLongestCommonPrefix(t *testing.T) {
	input := []string{"xyz_foo", "xyz_foo_bar", "xyz_foo_baz"}

	want := "xyz_foo"
	got := FindLongestCommonPrefix(input)

	if want != got {
		t.Errorf("wanted %s but got %s", want, got)
	}
}

func TestFindLongestCommonPrefixSingleWord(t *testing.T) {
	input := []string{"solo"}

	want := "solo"
	got := FindLongestCommonPrefix(input)

	if want != got {
		t.Errorf("wanted %s but got %s", want, got)
	}
}

func TestFindLongestCommonPrefixNoCommonPrefix(t *testing.T) {
	input := []string{"cat", "dog", "bird"}

	want := ""
	got := FindLongestCommonPrefix(input)

	if want != got {
		t.Errorf("wanted %s but got %s", want, got)
	}
}

func TestFindLongestCommonPrefixEmptyInput(t *testing.T) {
	input := []string{}

	want := ""
	got := FindLongestCommonPrefix(input)

	if want != got {
		t.Errorf("wanted %s but got %s", want, got)
	}
}

func TestSearchNotFound(t *testing.T) {
	tr := NewTrie()
	tr.Insert("cat")
	tr.Insert("cog")

	if tr.Search("dog") {
		t.Errorf("should not have found word dog")
	}

	if tr.Search("ca") {
		t.Errorf("should not have found partial word ca")
	}
}

func TestSearchEmptyTrie(t *testing.T) {
	tr := NewTrie()

	if tr.Search("anything") {
		t.Errorf("should not have found anything in empty trie")
	}
}

func TestPrefixesNoMatches(t *testing.T) {
	words := []string{"cat", "cog", "dog"}
	trie := NewTrie()

	for _, w := range words {
		trie.Insert(w)
	}

	got := trie.Prefixes("xyz")

	if len(got) != 0 {
		t.Errorf("expected no matches but got: %v", got)
	}
}

func TestPrefixesExactMatch(t *testing.T) {
	words := []string{"cat", "cog", "dog"}
	trie := NewTrie()

	for _, w := range words {
		trie.Insert(w)
	}

	want := []string{"cat"}
	got := trie.Prefixes("cat")

	slices.Sort(want)
	slices.Sort(got)

	if !slices.Equal(want, got) {
		t.Errorf("slices should be equal got: %v want: %v", got, want)
	}
}

func TestPrefixesEmptyPrefix(t *testing.T) {
	words := []string{"cat", "cog", "dog"}
	trie := NewTrie()

	for _, w := range words {
		trie.Insert(w)
	}

	want := []string{"cat", "cog", "dog"}
	got := trie.Prefixes("")

	slices.Sort(want)
	slices.Sort(got)

	if !slices.Equal(want, got) {
		t.Errorf("slices should be equal got: %v want: %v", got, want)
	}
}

func TestInsertDuplicateWord(t *testing.T) {
	tr := NewTrie()
	tr.Insert("cat")
	tr.Insert("cat")

	if !tr.Search("cat") {
		t.Errorf("should have found word cat")
	}

	got := tr.Prefixes("cat")
	want := []string{"cat"}

	if !slices.Equal(want, got) {
		t.Errorf("expected no duplicates, got: %v want: %v", got, want)
	}
}

func TestInsertEmptyWord(t *testing.T) {
	tr := NewTrie()
	tr.Insert("")

	if !tr.Search("") {
		t.Errorf("should have found empty word")
	}
}
