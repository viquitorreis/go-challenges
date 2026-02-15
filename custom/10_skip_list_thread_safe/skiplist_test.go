package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestSkipList() *SkipList {
	return NewSkipList(4, 0.5, rand.New(rand.NewSource(42)))
}

// --- Testes Básicos ---

func TestInsertAndSearch(t *testing.T) {
	sl := newTestSkipList()

	sl.Insert(10, "ten")
	sl.Insert(30, "thirty")
	sl.Insert(20, "twenty")

	val, ok := sl.Search(20)
	assert.True(t, ok)
	assert.Equal(t, "twenty", val)

	_, ok = sl.Search(99)
	assert.False(t, ok)
}

func TestInsertUpdate(t *testing.T) {
	sl := newTestSkipList()
	sl.Insert(10, "ten")
	sl.Insert(10, "TEN-updated") // mesmo score, deve atualizar

	val, ok := sl.Search(10)
	assert.True(t, ok)
	assert.Equal(t, "TEN-updated", val)
	assert.Equal(t, 1, sl.Size()) // não deve duplicar
}

func TestDelete(t *testing.T) {
	sl := newTestSkipList()
	sl.Insert(10, "ten")
	sl.Insert(20, "twenty")
	sl.Insert(30, "thirty")

	deleted := sl.Delete(20)
	assert.True(t, deleted)
	assert.Equal(t, 2, sl.Size())

	_, ok := sl.Search(20)
	assert.False(t, ok)

	deleted = sl.Delete(99) // não existe
	assert.False(t, deleted)
}

func TestRangeSearch(t *testing.T) {
	sl := newTestSkipList()
	for _, s := range []int{10, 20, 30, 40, 50, 60, 70} {
		sl.Insert(s, fmt.Sprintf("%d", s))
	}

	results := sl.RangeSearch(20, 50)
	assert.Equal(t, []any{"20", "30", "40", "50"}, results)

	// range vazio
	results = sl.RangeSearch(100, 200)
	assert.Empty(t, results)
}

func TestRangeSearchEmptyList(t *testing.T) {
	sl := newTestSkipList()
	results := sl.RangeSearch(0, 100)
	assert.Empty(t, results)
}

// --- Testes de Ordem ---

func TestInsertionOrder(t *testing.T) {
	sl := newTestSkipList()
	scores := []int{50, 10, 90, 30, 70, 20, 80, 40, 60}
	for _, s := range scores {
		sl.Insert(s, s)
	}

	results := sl.RangeSearch(10, 90)
	assert.Len(t, results, 9)
	for i := 1; i < len(results); i++ {
		assert.LessOrEqual(t, results[i-1].(int), results[i].(int))
	}
}

// --- Testes de Concorrência ---

func TestConcurrentInserts(t *testing.T) {
	sl := NewSkipList(8, 0.5, rand.New(rand.NewSource(0)))
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(score int) {
			defer wg.Done()
			sl.Insert(score, score)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, n, sl.Size())
	for i := 0; i < n; i++ {
		_, ok := sl.Search(i)
		assert.True(t, ok, "score %d not found", i)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	sl := NewSkipList(8, 0.5, rand.New(rand.NewSource(0)))
	for i := 0; i < 50; i++ {
		sl.Insert(i, i)
	}

	var wg sync.WaitGroup
	// Writers
	for i := 50; i < 100; i++ {
		wg.Add(1)
		go func(score int) {
			defer wg.Done()
			sl.Insert(score, score)
		}(i)
	}
	// Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(score int) {
			defer wg.Done()
			sl.Search(score)
		}(i)
	}
	// Range readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sl.RangeSearch(0, 100)
		}()
	}
	wg.Wait()
}

func TestConcurrentDeletes(t *testing.T) {
	sl := NewSkipList(8, 0.5, rand.New(rand.NewSource(0)))
	n := 50
	for i := 0; i < n; i++ {
		sl.Insert(i, i)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(score int) {
			defer wg.Done()
			sl.Delete(score)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 0, sl.Size())
}
