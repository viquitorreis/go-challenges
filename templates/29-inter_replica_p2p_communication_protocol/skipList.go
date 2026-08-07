package main

import (
	"math"
	"math/rand"
)

const MaxLevel = 16

type SkipListNode struct {
	score   int
	value   *PriceLevel     // stores a price + list of orders
	forward []*SkipListNode // forward[i] = next node on level i
}

type SkipList struct {
	listType Side
	head     *SkipListNode
	maxLevel int
	p        float64 // probability on going up on levels (generally 0.5)
	level    int     // level is the highest level currently is usage
	size     int
	rng      *rand.Rand
}

func NewSkipList(maxLevel int, listType Side, p float64, rng *rand.Rand) *SkipList {
	// head is a sentinel, doesnt count as an element, so size starts at 0
	return &SkipList{
		listType: listType,
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

// Insert inserts or updates a score+value on the list
// Uses the update array pattern:
//  1. Traverse the highest level until 0, storing all predecesses on []update
//  2. Generates randomLevel to a new node
//  3. Reconnects pointers using []update
func (sl *SkipList) Insert(score int, value *PriceLevel) {
	predecessors := make([]*SkipListNode, sl.maxLevel)

	curr := sl.head
	// level is the highest level currently is usage
	for i := sl.level - 1; i >= 0; i-- {
		for sl.shouldAdvance(curr.forward[i], score) {
			curr = curr.forward[i]
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

// Search searches by score. Returns (value, true) if found
// Runs from top to bottom, while forward[level] exists and score < target keep going
//
//	When it cant go anymore, it will go down to the next level
func (sl *SkipList) Search(score int) (any, bool) {
	curr := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		// search on all nodes on this level from smallest to the biggest
		// as its ordered we will stop right over curr.score < score, so the next can be possible candidate
		// because curr will be the only node that can have its score < target score
		for sl.shouldAdvance(curr.forward[i], score) {
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

	curr := sl.head
	// level is the highest level currently is usage
	for i := sl.level - 1; i >= 0; i-- {
		for sl.shouldAdvance(curr.forward[i], score) {
			curr = curr.forward[i]
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

// returns PriceLevel of best priority (smaller score on ask, and biggest on Bid) in O(1), since its always the first one
func (sl *SkipList) Front() (*PriceLevel, bool) {
	node := sl.head.forward[0]
	if node == nil {
		return nil, false
	}

	return node.value, true
}

func (sl *SkipList) All() []*PriceLevel {
	var result []*PriceLevel

	for node := sl.head.forward[0]; node != nil; node = node.forward[0] {
		result = append(result, node.value)
	}

	return result
}

func (sl *SkipList) Size() int {
	return sl.size
}

func (sl *SkipList) shouldAdvance(node *SkipListNode, target int) bool {
	if node == nil {
		return false
	}

	if sl.listType == Ask {
		return node.score < target
	}

	return node.score > target
}
