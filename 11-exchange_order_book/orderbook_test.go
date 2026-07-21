package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndMatch(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 103, 5))
	ob.AddOrder(NewOrder("a1", "bob", Ask, 103, 3))

	trades := ob.Match()
	assert.Len(t, trades, 1)
	assert.Equal(t, 3, trades[0].Quantity)
	assert.Equal(t, 2, ob.BidDepth()) // 5 - 3 = 2 restantes
	assert.Equal(t, 0, ob.AskDepth())
}

func TestNoMatch(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 100, 5))
	ob.AddOrder(NewOrder("a1", "bob", Ask, 105, 3))

	trades := ob.Match()
	assert.Empty(t, trades)
}

func TestCancel(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 100, 5))
	assert.True(t, ob.Cancel("b1"))
	assert.Equal(t, 0, ob.BidDepth())
	assert.False(t, ob.Cancel("b1")) // já removida
}

func TestPriceTimePriority(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	// dois bids no mesmo preço — b1 chegou primeiro
	ob.AddOrder(NewOrder("b1", "alice", Bid, 103, 2))
	ob.AddOrder(NewOrder("b2", "bob", Bid, 103, 2))
	ob.AddOrder(NewOrder("a1", "carol", Ask, 103, 2))

	trades := ob.Match()
	assert.Len(t, trades, 1)
	assert.Equal(t, "b1", trades[0].BidOrderID) // b1 tem prioridade
}

func TestConcurrentAddOrders(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ob.AddOrder(NewOrder(
				fmt.Sprintf("b%d", n), "user", Bid, 100+n, 1,
			))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, ob.BidDepth())
}

// BenchmarkCancelMiddleOfLevel mede o custo de cancelar UMA ordem específica
// no meio da fila de um price level, variando quantas ordens existem nesse
// nível. Isso expõe o Problema 2 do diagnóstico: []*Order não suporta
// remoção O(1) do meio -- cada Cancel aqui escaneia o slice inteiro pra
// reconstruí-lo sem a ordem removida.
//
// Setup (criar o book e popular N ordens) fica fora da região cronometrada
// via StopTimer/StartTimer -- só o Cancel em si é medido.
func BenchmarkCancelMiddleOfLevel(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("depth_%d", n), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ob := NewOrderBook("BTC-USD")
				ids := make([]string, n)

				for j := range n {
					id := fmt.Sprintf("o%d", j)
					ids[j] = id
					ob.AddOrder(NewOrder(id, "user", Bid, 100, 1))
				}

				middleID := ids[n/2]
				b.StartTimer()

				ob.Cancel(middleID)
			}
		})
	}
}

// BenchmarkCancelManyDistinctLevels mede o custo de cancelar N ordens, cada
// uma num price level DIFERENTE (1 ordem por nível), esvaziando cada nível
// no processo. Isso expõe o Problema 1 do diagnóstico: o heap não suporta
// remoção arbitrária -- toda vez que um nível esvazia, o código atual
// reconstrói o heap inteiro (O(n log n)) em vez de remover pontualmente.
//
// "ns/op" aqui representa o custo de cancelar TODAS as N ordens dentro de
// uma iteração, não de uma única ordem -- divida por N pra comparar o custo
// médio por cancelamento individual entre tamanhos diferentes.
func BenchmarkCancelManyDistinctLevels(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("levels_%d", n), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ob := NewOrderBook("BTC-USD")
				ids := make([]string, n)

				for j := range n {
					id := fmt.Sprintf("o%d", j)
					ids[j] = id
					// preço distinto por ordem -- cada uma vira seu próprio
					// price level, com profundidade 1
					ob.AddOrder(NewOrder(id, "user", Bid, 100+j, 1))
				}

				b.StartTimer()

				for _, id := range ids {
					ob.Cancel(id)
				}
			}
		})
	}
}
