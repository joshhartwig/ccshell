package completions

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
	for i := rune('a'); i <= rune('z'); i++ {
		next, ok := node.Children[i]
		if !ok {
			continue
		}
		dfsCollectCompletions(next, prefix+string(i), results)
	}
}
