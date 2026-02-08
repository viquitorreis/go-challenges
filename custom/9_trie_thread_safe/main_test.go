package main

import (
	"fmt"
	"sync"
	"testing"
)

// TestBasicInsertAndSearch testa a funcionalidade básica de inserção e busca
func TestBasicInsertAndSearch(t *testing.T) {
	trie := NewTrie()

	// Insere algumas palavras
	words := []string{"cat", "car", "card", "care", "careful"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Testa que todas as palavras inseridas são encontradas
	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Expected to find word '%s', but Search returned false", word)
		}
	}

	// Testa que prefixos de palavras não são considerados palavras completas
	// Por exemplo, "car" é uma palavra, mas "ca" não é
	notWords := []string{"c", "ca", "car", "care"}
	for _, prefix := range notWords {
		// "car" e "care" são palavras reais, então pulamos elas
		if prefix == "car" || prefix == "care" {
			continue
		}
		if trie.Search(prefix) {
			t.Errorf("Expected NOT to find '%s' as a complete word, but Search returned true", prefix)
		}
	}

	// Testa que palavras não inseridas não são encontradas
	if trie.Search("dog") {
		t.Error("Expected NOT to find 'dog', but Search returned true")
	}
}

// TestStartsWith verifica se a detecção de prefixos funciona corretamente
func TestStartsWith(t *testing.T) {
	trie := NewTrie()

	words := []string{"apple", "application", "apply"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Testa prefixos que existem
	validPrefixes := []string{"a", "ap", "app", "appl", "apple"}
	for _, prefix := range validPrefixes {
		if !trie.StartsWith(prefix) {
			t.Errorf("Expected prefix '%s' to exist, but StartsWith returned false", prefix)
		}
	}

	// Testa prefixos que não existem
	invalidPrefixes := []string{"b", "app1", "applications"}
	for _, prefix := range invalidPrefixes {
		if trie.StartsWith(prefix) {
			t.Errorf("Expected prefix '%s' NOT to exist, but StartsWith returned true", prefix)
		}
	}
}

// TestAutoCompleteBasic testa a funcionalidade básica de autocomplete
func TestAutoCompleteBasic(t *testing.T) {
	trie := NewTrie()

	words := []string{"cat", "car", "card", "care", "careful", "dog", "dodge", "door"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Testa autocomplete para "ca" - deve retornar todas as palavras que começam com "ca"
	results := trie.AutoComplete("ca", 10)
	expected := map[string]bool{
		"cat": true, "car": true, "card": true, "care": true, "careful": true,
	}

	if len(results) != len(expected) {
		t.Errorf("Expected %d results for prefix 'ca', got %d", len(expected), len(results))
	}

	for _, word := range results {
		if !expected[word] {
			t.Errorf("Unexpected word '%s' in autocomplete results for 'ca'", word)
		}
	}

	// Verifica que todas as palavras esperadas estão presentes
	for word := range expected {
		found := false
		for _, result := range results {
			if result == word {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected word '%s' in autocomplete results for 'ca', but it was not found", word)
		}
	}
}

// TestAutoCompleteLimit testa se o limite de resultados é respeitado
func TestAutoCompleteLimit(t *testing.T) {
	trie := NewTrie()

	words := []string{"test1", "test2", "test3", "test4", "test5", "test6"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Testa com limite de três resultados
	results := trie.AutoComplete("test", 3)
	if len(results) != 3 {
		t.Errorf("Expected exactly 3 results with limit=3, got %d", len(results))
	}

	// Testa com limite maior que o número de palavras disponíveis
	results = trie.AutoComplete("test", 100)
	if len(results) != 6 {
		t.Errorf("Expected 6 results (all available words), got %d", len(results))
	}

	// Testa com limite zero, que significa sem limite
	results = trie.AutoComplete("test", 0)
	if len(results) != 6 {
		t.Errorf("Expected 6 results with limit=0 (no limit), got %d", len(results))
	}
}

// TestAutoCompleteNonExistentPrefix testa autocomplete com prefixo que não existe
func TestAutoCompleteNonExistentPrefix(t *testing.T) {
	trie := NewTrie()

	trie.Insert("apple")
	trie.Insert("application")

	results := trie.AutoComplete("banana", 10)
	if len(results) != 0 {
		t.Errorf("Expected empty results for non-existent prefix, got %d results", len(results))
	}
}

// TestEmptyTrie testa operações em uma trie vazia
func TestEmptyTrie(t *testing.T) {
	trie := NewTrie()

	if trie.Search("anything") {
		t.Error("Expected Search to return false on empty trie")
	}

	if trie.StartsWith("any") {
		t.Error("Expected StartsWith to return false on empty trie")
	}

	results := trie.AutoComplete("any", 10)
	if len(results) != 0 {
		t.Errorf("Expected AutoComplete to return empty slice on empty trie, got %d results", len(results))
	}
}

// TestSingleCharacterWords testa palavras de um único caractere
func TestSingleCharacterWords(t *testing.T) {
	trie := NewTrie()

	trie.Insert("a")
	trie.Insert("i")

	if !trie.Search("a") {
		t.Error("Expected to find single character word 'a'")
	}

	if !trie.Search("i") {
		t.Error("Expected to find single character word 'i'")
	}

	results := trie.AutoComplete("a", 10)
	found := false
	for _, word := range results {
		if word == "a" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'a' in autocomplete results")
	}
}

// TestOverlappingWords testa palavras onde uma é prefixo da outra
func TestOverlappingWords(t *testing.T) {
	trie := NewTrie()

	// "car" é prefixo de "card" e "careful"
	trie.Insert("car")
	trie.Insert("card")
	trie.Insert("careful")

	// Todas as três devem ser encontradas como palavras completas
	if !trie.Search("car") {
		t.Error("Expected to find 'car' as a complete word")
	}
	if !trie.Search("card") {
		t.Error("Expected to find 'card' as a complete word")
	}
	if !trie.Search("careful") {
		t.Error("Expected to find 'careful' as a complete word")
	}

	// Autocomplete de "car" deve incluir todas as três
	results := trie.AutoComplete("car", 10)
	expected := []string{"car", "card", "careful"}

	for _, word := range expected {
		found := false
		for _, result := range results {
			if result == word {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find '%s' in autocomplete for 'car'", word)
		}
	}
}

// TestConcurrentInserts testa inserções concorrentes
func TestConcurrentInserts(t *testing.T) {
	trie := NewTrie()
	var wg sync.WaitGroup

	// Cria cem goroutines inserindo palavras simultaneamente
	numGoroutines := 100
	wordsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < wordsPerGoroutine; j++ {
				word := fmt.Sprintf("word_%d_%d", id, j)
				trie.Insert(word)
			}
		}(i)
	}

	wg.Wait()

	// Verifica que todas as palavras foram inseridas corretamente
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < wordsPerGoroutine; j++ {
			word := fmt.Sprintf("word_%d_%d", i, j)
			if !trie.Search(word) {
				t.Errorf("Expected to find word '%s' after concurrent insertion", word)
			}
		}
	}
}

// TestConcurrentReads testa leituras concorrentes
func TestConcurrentReads(t *testing.T) {
	trie := NewTrie()

	// Insere algumas palavras antes de iniciar as leituras concorrentes
	words := []string{"apple", "application", "apply", "banana", "band", "bandana"}
	for _, word := range words {
		trie.Insert(word)
	}

	var wg sync.WaitGroup
	numReaders := 100

	// Múltiplas goroutines lendo simultaneamente
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Faz várias operações de leitura
			for _, word := range words {
				if !trie.Search(word) {
					t.Errorf("Expected to find word '%s' during concurrent reads", word)
				}
			}

			if !trie.StartsWith("app") {
				t.Error("Expected to find prefix 'app' during concurrent reads")
			}

			results := trie.AutoComplete("ba", 10)
			if len(results) == 0 {
				t.Error("Expected non-empty autocomplete results during concurrent reads")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentMixedOperations testa leituras e escritas concorrentes
func TestConcurrentMixedOperations(t *testing.T) {
	trie := NewTrie()
	var wg sync.WaitGroup

	// Algumas goroutines inserindo
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				word := fmt.Sprintf("insert_%d_%d", id, j)
				trie.Insert(word)
			}
		}(i)
	}

	// Outras goroutines fazendo buscas
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				word := fmt.Sprintf("insert_%d_%d", id, j)
				// Pode ou não encontrar, dependendo do timing, mas não deve dar panic
				trie.Search(word)
			}
		}(i)
	}

	// Outras fazendo autocomplete
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				// Autocomplete pode retornar resultados parciais, mas não deve dar panic
				trie.AutoComplete("insert", 5)
			}
		}()
	}

	wg.Wait()

	// Verifica que pelo menos algumas palavras foram inseridas
	found := 0
	for i := 0; i < 50; i++ {
		for j := 0; j < 10; j++ {
			word := fmt.Sprintf("insert_%d_%d", i, j)
			if trie.Search(word) {
				found++
			}
		}
	}

	if found == 0 {
		t.Error("Expected to find at least some words after concurrent mixed operations")
	}
}

// TestUnicodeCharacters testa a trie com caracteres Unicode
func TestUnicodeCharacters(t *testing.T) {
	trie := NewTrie()

	// Palavras com caracteres acentuados e outros caracteres Unicode
	words := []string{"café", "naïve", "résumé", "日本語", "hello"}
	for _, word := range words {
		trie.Insert(word)
	}

	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Expected to find Unicode word '%s'", word)
		}
	}

	// Testa autocomplete com prefixos Unicode
	results := trie.AutoComplete("caf", 10)
	fmt.Printf("DEBUG: AutoComplete('caf') retornou: %v\n", results)
	found := false
	for _, result := range results {
		if result == "café" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'café' in autocomplete results for 'caf'")
	}
}

// TestLargeDataset testa a trie com um dataset maior
func TestLargeDataset(t *testing.T) {
	trie := NewTrie()

	// Insere mil palavras
	numWords := 1000
	for i := 0; i < numWords; i++ {
		word := fmt.Sprintf("word_%04d", i)
		trie.Insert(word)
	}

	// Verifica algumas palavras aleatórias
	testWords := []int{0, 100, 500, 999}
	for _, i := range testWords {
		word := fmt.Sprintf("word_%04d", i)
		if !trie.Search(word) {
			t.Errorf("Expected to find word '%s' in large dataset", word)
		}
	}

	// Testa autocomplete com muitos resultados
	results := trie.AutoComplete("word_", 50)
	if len(results) != 50 {
		t.Errorf("Expected 50 autocomplete results, got %d", len(results))
	}
}

// BenchmarkInsert mede a performance de inserção
func BenchmarkInsert(b *testing.B) {
	trie := NewTrie()
	words := make([]string, b.N)

	for i := 0; i < b.N; i++ {
		words[i] = fmt.Sprintf("benchmark_word_%d", i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		trie.Insert(words[i])
	}
}

// BenchmarkSearch mede a performance de busca
func BenchmarkSearch(b *testing.B) {
	trie := NewTrie()

	// Prepara a trie com mil palavras
	for i := 0; i < 1000; i++ {
		word := fmt.Sprintf("benchmark_word_%d", i)
		trie.Insert(word)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		word := fmt.Sprintf("benchmark_word_%d", i%1000)
		trie.Search(word)
	}
}

// BenchmarkAutoComplete mede a performance de autocomplete
func BenchmarkAutoComplete(b *testing.B) {
	trie := NewTrie()

	// Prepara a trie com mil palavras
	for i := 0; i < 1000; i++ {
		word := fmt.Sprintf("benchmark_word_%d", i)
		trie.Insert(word)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		trie.AutoComplete("benchmark", 10)
	}
}

// BenchmarkConcurrentInserts mede a performance de inserções concorrentes
func BenchmarkConcurrentInserts(b *testing.B) {
	trie := NewTrie()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			word := fmt.Sprintf("concurrent_word_%d", i)
			trie.Insert(word)
			i++
		}
	})
}
