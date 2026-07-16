package main

import (
	"math"
	"math/rand"
	"sync"
)

const MaxLevel = 16

type SkipListNode struct {
	score   int
	value   any
	forward []*SkipListNode // forward[i] = próximo node no nível i
	mu      sync.RWMutex    // opcional: se quiser locking por node
}

type SkipList struct {
	head     *SkipListNode
	maxLevel int
	p        float64 // probability on going up on levels (generally 0.5)
	level    int     // level is the highest level currently is usage
	size     int
	mu       sync.RWMutex
	rng      *rand.Rand
}

func NewSkipList(maxLevel int, p float64, rng *rand.Rand) *SkipList {
	// head is a sentinel, doesnt count as an element, so size starts at 0
	return &SkipList{
		head: &SkipListNode{
			score:   math.MinInt,
			forward: make([]*SkipListNode, maxLevel),
		},
		maxLevel: maxLevel,
		p:        p,
		level:    1,
		size:     0,
		rng:      rng,
	}
}

func (sl *SkipList) randomLevel() int {
	i := 1
	for i < sl.maxLevel && sl.rng.Float64() < sl.p {
		i++
	}
	return i
}

// Insert insere ou atualiza um score+value na lista
// Usa o update array pattern:
//  1. Percorre do nível mais alto até 0, guardando predecessores em update[]
//  2. Gera randomLevel para o novo node
//  3. Reconecta ponteiros usando update[]
func (sl *SkipList) Insert(score int, value any) {
	predecessors := make([]*SkipListNode, sl.maxLevel)

	sl.mu.Lock()
	defer sl.mu.Unlock()

	curr := sl.head
	// level is the highest level currently is usage
	for i := sl.level - 1; i >= 0; i-- {
		for curr.forward[i] != nil && curr.forward[i].score < score {
			curr = curr.forward[i] // go to next level
		}
		predecessors[i] = curr
	}

	// if score exists already, we update it
	candidate := curr.forward[0]
	if candidate != nil && candidate.score == score {
		candidate.value = value
		return
	}

	// new node
	level := sl.randomLevel()
	newNode := &SkipListNode{
		score:   score,
		value:   value,
		forward: make([]*SkipListNode, level),
	}

	// if node have more level than the list currently uses
	// extra levels have head as predecessor because none existing node can reach the new level
	// without it predecessors[n] being n an innexisting level, would be nil and therefore panic
	// when we try to access it
	if level > sl.level {
		// we loop all the levels that already exists
		for i := sl.level; i < level; i++ {
			predecessors[i] = sl.head
		}
		sl.level = level
	}

	// we insert on each level
	// the new node needs to be connected to all the levels that he is part of
	// the physical node is only one, but have multiple references for other levels
	for i := 0; i < level; i++ {
		// pointing to what came before
		// predecessors[i] --> predecessors[i].forward[i]
		newNode.forward[i] = predecessors[i].forward[i]
		// predecessor point to the new node
		// predecessors[i] --> newNode --> predecessors[i].forward[i]
		predecessors[i].forward[i] = newNode
	}

	sl.size++
}

// Search busca por score. Retorna (value, true) se encontrado.
// Percorre de cima para baixo: enquanto forward[level] existe e score < target,
// avança. Quando não pode mais avançar, desce um nível.
func (sl *SkipList) Search(score int) (any, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	curr := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		// search on all nodes on this level from smallest to the biggest
		// as its ordered we will stop right over curr.score < score, so the next can be possible candidate
		// because curr will be the only node that can have its score < target score
		for curr.forward[i] != nil && curr.forward[i].score < score {
			curr = curr.forward[i]
		}
	}

	candidate := curr.forward[0]
	if candidate != nil && candidate.score == score {
		return candidate.value, true
	}

	return nil, false
}

func (sl *SkipList) Delete(score int) bool {
	// 1. update array just like the insert. We find the predecessor of each level
	// by the end curr.forward[0] is a candidate
	predecessors := make([]*SkipListNode, sl.maxLevel)

	sl.mu.Lock()
	defer sl.mu.Unlock()

	curr := sl.head
	// level is the highest level currently is usage
	for i := sl.level - 1; i >= 0; i-- {
		for curr.forward[i] != nil && curr.forward[i].score < score {
			curr = curr.forward[i] // go to next level
		}
		predecessors[i] = curr
	}

	// if score exists already, we update it
	target := curr.forward[0]
	if target == nil || target.score != score {
		return false
	}

	// desconnect target from each level by pointing predecessors to target's forward nodes
	for i := 0; i < len(target.forward); i++ {
		predecessors[i].forward[i] = target.forward[i]
	}

	// if any level is now empty we lower the amount of levels available
	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}

	sl.size--
	return true
}

func (sl *SkipList) RangeSearch(min, max int) []any {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	curr := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		// search on all nodes on this level from smallest to the biggest
		// as its ordered we will stop right over curr.score < score, so the next can be possible candidate
		// because curr will be the only node that can have its score < min score
		for curr.forward[i] != nil && curr.forward[i].score < min {
			curr = curr.forward[i]
		}
	}

	elements := []any{}

	curr = curr.forward[0]
	for curr != nil && curr.score <= max {
		elements = append(elements, curr.value)
		curr = curr.forward[0]
	}

	return elements
}

func (sl *SkipList) Size() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.size
}
