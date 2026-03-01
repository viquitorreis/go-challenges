package main

import (
	"fmt"
	"sync"
)

func main() {
	trie := NewTrie()

	fmt.Println("=== Testando Insert e Search ===")
	words := []string{"cat", "car", "card", "care", "careful", "dog", "dodge", "door"}

	for _, word := range words {
		trie.Insert(word)
		fmt.Printf("Inserted: %s\n", word)
	}

	fmt.Println("\n=== Testando Search (exact match) ===")
	testWords := []string{"car", "card", "careful", "cat", "can", "do"}
	for _, word := range testWords {
		found := trie.Search(word)
		fmt.Printf("Search('%s'): %v\n", word, found)
	}

	fmt.Println("\n=== Testando StartsWith ===")
	prefixes := []string{"ca", "car", "do", "doo", "cat", "x"}
	for _, prefix := range prefixes {
		exists := trie.StartsWith(prefix)
		fmt.Printf("StartsWith('%s'): %v\n", prefix, exists)
	}

	fmt.Println("\n=== Testando AutoComplete ===")
	testPrefixes := []string{"ca", "car", "do"}
	for _, prefix := range testPrefixes {
		suggestions := trie.AutoComplete(prefix, 5)
		fmt.Printf("AutoComplete('%s', limit=5): %v\n", prefix, suggestions)
	}

	fmt.Println("\n=== Testando AutoComplete sem limite ===")
	allCar := trie.AutoComplete("car", 0)
	fmt.Printf("AutoComplete('car', no limit): %v\n", allCar)

	fmt.Println("\n=== Testando operações concorrentes ===")
	var wg sync.WaitGroup

	// 10 goroutines inserindo palavras simultaneamente
	newWords := []string{
		"apple", "application", "apply", "banana", "band", "bandana",
		"can", "candy", "candle", "canon",
	}

	for _, word := range newWords {
		wg.Add(1)
		go func(w string) {
			defer wg.Done()
			trie.Insert(w)
		}(word)
	}

	// 10 goroutines fazendo searches simultaneamente
	for _, word := range testWords {
		wg.Add(1)
		go func(w string) {
			defer wg.Done()
			trie.Search(w)
		}(word)
	}

	// 10 goroutines fazendo autocomplete simultaneamente
	for _, prefix := range testPrefixes {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			trie.AutoComplete(p, 3)
		}(prefix)
	}

	wg.Wait()

	fmt.Println("\n=== Verificando palavras inseridas concorrentemente ===")
	for _, word := range newWords {
		found := trie.Search(word)
		fmt.Printf("Search('%s'): %v\n", word, found)
	}

	fmt.Println("\n=== Testando AutoComplete em 'can' ===")
	canWords := trie.AutoComplete("can", 10)
	fmt.Printf("AutoComplete('can'): %v\n", canWords)
}

type TrieNode struct {
	val      rune
	children map[rune]*TrieNode
	isWord   bool
	mu       sync.RWMutex // lock granular por node
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

func (t *Trie) Insert(word string) {
	fmt.Printf("\n=== Inserindo '%s' ===\n", word)
	curr := t.root

	for i, char := range word {
		fmt.Printf("Iteração %d, char='%c'\n", i, char)
		curr.mu.Lock()
		val, ok := curr.children[char]
		if !ok {
			newNode := &TrieNode{
				val:      char,
				children: make(map[rune]*TrieNode),
			}
			curr.children[char] = newNode
			curr.mu.Unlock()
			curr = newNode
			continue
		}
		curr.mu.Unlock()
		curr = val
	}

	curr.mu.Lock()
	curr.isWord = true
	curr.mu.Unlock()
}

func (t *Trie) Search(word string) bool {
	curr := t.root

	for _, char := range word {
		curr.mu.RLock()
		val, ok := curr.children[char]
		curr.mu.RUnlock()
		if !ok {
			return false
		}
		curr = val
	}

	curr.mu.RLock()
	defer curr.mu.RUnlock()

	return curr.isWord
}

func (t *Trie) StartsWith(prefix string) bool {
	curr := t.root

	for _, char := range prefix {
		curr.mu.RLock()
		val, ok := curr.children[char]
		curr.mu.RUnlock()
		if !ok {
			return false
		}
		curr = val
	}

	return true
}

func (t *Trie) AutoComplete(prefix string, limit int) []string {
	getPrefixNode := func(p string) *TrieNode {
		curr := t.root

		for _, char := range p {
			curr.mu.RLock()
			val, ok := curr.children[char]
			curr.mu.RUnlock()
			if !ok {
				return nil
			}
			curr = val
		}

		return curr
	}

	prefixNode := getPrefixNode(prefix)
	words := []string{}

	if prefixNode == nil {
		return []string{}
	}

	prefixNode.mu.RLock()
	isWord := prefixNode.isWord
	prefixNode.mu.RUnlock()

	if isWord {
		words = append(words, prefix)
	}

	collectWords(prefixNode, prefix, &words, limit)

	return words
}

func collectWords(node *TrieNode, currentWord string, words *[]string, limit int) {
	if limit > 0 && len(*words) >= limit {
		return
	}

	node.mu.Lock()
	type childSnapshot struct {
		char   rune
		node   *TrieNode
		isWord bool
	}

	childrensSnapshot := make([]childSnapshot, 0, len(node.children))

	for char, child := range node.children {
		child.mu.Lock()

		childrensSnapshot = append(childrensSnapshot, childSnapshot{
			char:   char,
			node:   child,
			isWord: child.isWord,
		})

		child.mu.Unlock()
	}
	node.mu.Unlock()

	for _, item := range childrensSnapshot {
		if limit > 0 && len(*words) >= limit {
			return
		}

		newWord := currentWord + string(item.char)

		if item.isWord {
			*words = append(*words, newWord)
		}

		collectWords(item.node, newWord, words, limit)
	}
}
