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

// TrieNode representa um node na trie
// Cada node tem um map de children (um por caractere possível)
// e uma flag indicando se é fim de palavra
type TrieNode struct {
	children map[rune]*TrieNode
	isWord   bool
	mu       sync.RWMutex // Lock granular por node
}

// Trie é a estrutura principal
type Trie struct {
	root *TrieNode
}

// NewTrie cria uma nova trie vazia
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert adiciona uma palavra na trie
// Thread-safe: múltiplas goroutines podem inserir simultaneamente
// TODO: Implementar usando hand-over-hand locking
// Dica: você precisa adquirir write lock do node atual, depois do child,
// depois liberar o lock do node pai antes de continuar descendo
func (t *Trie) Insert(word string) {
	// TODO: Implementar
	// Lembre-se: ao criar um novo child node, você precisa inicializar
	// o map de children dele também (make)
}

// Search verifica se uma palavra completa existe na trie
// Retorna true apenas se a palavra existe E está marcada como fim de palavra
// TODO: Implementar usando read locks
// Dica: você pode usar RLock aqui porque não está modificando
func (t *Trie) Search(word string) bool {
	// TODO: Implementar
	return false
}

// StartsWith verifica se existe alguma palavra na trie que começa com o prefix dado
// TODO: Implementar
// Dica: diferente do Search, você não precisa verificar isEnd,
// basta chegar até o final do prefix
func (t *Trie) StartsWith(prefix string) bool {
	// TODO: Implementar
	return false
}

// AutoComplete retorna todas as palavras que começam com o prefix dado
// limit define o número máximo de sugestões (0 = sem limite)
// TODO: Esta é a parte mais desafiadora!
// Dica: você precisa:
//  1. Navegar até o node do prefix (com locks)
//  2. Fazer DFS a partir dali coletando palavras
//  3. O problema: você NÃO PODE segurar locks durante toda a DFS
//     porque isso bloquearia outras operações por muito tempo
//  4. Solução: copiar a estrutura necessária ou usar snapshot approach
func (t *Trie) AutoComplete(prefix string, limit int) []string {
	// TODO: Implementar
	return nil
}

// Delete remove uma palavra da trie
// TODO: Implementar (bonus se tiver tempo)
// Esta é complexa porque você precisa remover nodes que não são mais necessários
// Mas cuidado: não pode remover um node que é prefix de outra palavra
// Exemplo: se você tem "car" e "card", ao deletar "car" não pode remover o node 'r'
func (t *Trie) Delete(word string) bool {
	// TODO: Implementar
	return false
}

// Funções auxiliares que você provavelmente vai precisar

// findNode navega até o node que representa o final de um prefix
// Retorna o node encontrado (ou nil se prefix não existe)
// TODO: Implementar esta helper - será útil para Search, StartsWith e AutoComplete
func (t *Trie) findNode(prefix string) *TrieNode {
	// TODO: Implementar
	return nil
}

// collectWords faz DFS a partir de um node coletando todas as palavras
// currentWord é o prefix acumulado até chegar neste node
// TODO: Implementar esta helper para usar no AutoComplete
// Dica: esta função vai ser recursiva
func collectWords(node *TrieNode, currentWord string, words *[]string, limit int) {
	// TODO: Implementar DFS recursivo
	// Lembre-se de verificar se já coletou o limite de palavras
}
