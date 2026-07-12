package main

import (
	"math/rand"
	"sync"
)

const MaxLevel = 16

// SkipListNode representa um node em múltiplos níveis
type SkipListNode struct {
	score   int
	value   any
	forward []*SkipListNode // forward[i] = próximo node no nível i
	mu      sync.RWMutex    // opcional: se quiser locking por node
}

// SkipList é a estrutura principal
type SkipList struct {
	head     *SkipListNode
	maxLevel int
	p        float64 // probabilidade de subir de nível (geralmente 0.5)
	level    int     // nível atual máximo utilizado
	size     int
	mu       sync.RWMutex
	rng      *rand.Rand
}

// NewSkipList cria uma nova skip list
// maxLevel: altura máxima permitida
// p: probabilidade de um node subir de nível (0.5 = 50%)
// rng: source de random (injetado para testes determinísticos)
func NewSkipList(maxLevel int, p float64, rng *rand.Rand) *SkipList {
	// TODO: criar head node com maxLevel forward pointers (todos nil)
	// head.score pode ser -infinito (math.MinInt)
	panic("not implemented")
}

// randomLevel gera a altura de um novo node probabilisticamente
// Começa em 1 e incrementa enquanto rand < p E level < maxLevel
func (sl *SkipList) randomLevel() int {
	// TODO
	panic("not implemented")
}

// Insert insere ou atualiza um score+value na lista
// Usa o update array pattern:
//  1. Percorre do nível mais alto até 0, guardando predecessores em update[]
//  2. Gera randomLevel para o novo node
//  3. Reconecta ponteiros usando update[]
func (sl *SkipList) Insert(score int, value any) {
	// TODO
	panic("not implemented")
}

// Search busca por score. Retorna (value, true) se encontrado.
// Percorre de cima para baixo: enquanto forward[level] existe e score < target,
// avança. Quando não pode mais avançar, desce um nível.
func (sl *SkipList) Search(score int) (any, bool) {
	// TODO
	panic("not implemented")
}

// Delete remove o node com o score dado, se existir.
// Mesmo padrão do Insert: update array primeiro, depois reconecta ponteiros.
func (sl *SkipList) Delete(score int) bool {
	// TODO
	panic("not implemented")
}

// RangeSearch retorna todos os valores com score entre min e max (inclusive),
// em ordem crescente. Hint: chegue até min usando a lógica de Search,
// depois percorra o nível 0 enquanto score <= max.
func (sl *SkipList) RangeSearch(min, max int) []any {
	// TODO
	panic("not implemented")
}

// Size retorna o número de elementos na lista
func (sl *SkipList) Size() int {
	// TODO
	panic("not implemented")
}
