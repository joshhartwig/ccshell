package completions

import (
	"slices"
	"strings"
)

type TrieNode struct {
	Children  map[rune]*TrieNode
	EndOfWord bool
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Root: NewTrieNode(),
	}
}

// TrieNode acts as a node in our Trie with Children and Marker for if end of word
func NewTrieNode() *TrieNode {
	return &TrieNode{
		Children:  make(map[rune]*TrieNode),
		EndOfWord: false,
	}
}

// Inserts a word into the Trie data structure
func (t *Trie) Insert(word string) {
	word = strings.ToLower(word)   // normalize the word removing spacing
	word = strings.Trim(word, " ") // and leading + trailing spaces

	// iterate through each char node looking for the next char
	current := t.Root
	for _, char := range word {
		node, ok := current.Children[char]
		if !ok {
			newNode := NewTrieNode()
			current.Children[char] = newNode
			current = newNode
			continue
		}
		current = node
	}
	current.EndOfWord = true
}

// Search for word in Trie data structure
func (t *Trie) Search(word string) bool {
	current := t.Root

	for _, char := range word {
		node, ok := current.Children[char]
		if !ok {
			return false
		}
		current = node
	}
	return current.EndOfWord
}

// pass in ca return cat, can
func (t *Trie) Prefixes(prefix string) []string {
	current := t.Root
	for _, char := range prefix {
		node, ok := current.Children[char]
		if !ok {
			return []string{}
		}
		current = node
	}

	results := []string{}
	dfsCollectCompletions(current, prefix, &results)
	return results
}

func dfsCollectCompletions(node *TrieNode, prefix string, results *[]string) {
	if node.EndOfWord {
		*results = append(*results, prefix)
	}
	for ch, next := range node.Children {
		dfsCollectCompletions(next, prefix+string(ch), results)
	}
}

func FindLongestCommonPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}
	slices.Sort(words)
	first := words[0]
	last := words[len(words)-1]

	var commonPrefix []rune
	for idx, char := range first {
		if idx > len(last)-1 {
			return string(commonPrefix)
		}
		if first[idx] == last[idx] {
			commonPrefix = append(commonPrefix, char)
		}
	}
	return string(commonPrefix)
}
