package main

import "sync"

type Side int

const (
	Bid Side = iota
	Ask
)

// Order representa uma ordem individual no book
type Order struct {
	ID        string
	UserID    string
	Side      Side
	Price     int // em centavos para evitar float
	Quantity  int
	Timestamp int64 // unix nano — usado para price-time priority
}

// Trade representa uma execução — quando bid e ask se encontram
type Trade struct {
	BidOrderID string
	AskOrderID string
	Price      int
	Quantity   int
}

// PriceLevel agrupa todas as ordens num mesmo preço (fila FIFO)
type PriceLevel struct {
	Price  int
	Orders []*Order
}

// OrderBook mantém bids e asks para um único ticker
type OrderBook struct {
	Symbol string

	// TODO: estrutura para bids ordenados (maior preço primeiro)
	// TODO: estrutura para asks ordenados (menor preço primeiro)
	// TODO: índice de orderID -> Order para cancelamento O(1)

	mu sync.Mutex
}

func NewOrderBook(symbol string) *OrderBook                        { panic("not implemented") }
func NewOrder(id, userID string, side Side, price, qty int) *Order { panic("not implemented") }

// AddOrder adiciona uma ordem ao lado correto do book
func (ob *OrderBook) AddOrder(o *Order) { panic("not implemented") }

// Match executa todas as ordens que se cruzam e retorna os trades gerados.
// Regra: enquanto max(bid) >= min(ask), executa pelo preço do ask (ou bid, sua escolha).
// Quantidade executada = min(bid.Quantity, ask.Quantity).
// Ordens parcialmente executadas permanecem no book com quantidade reduzida.
func (ob *OrderBook) Match() []Trade { panic("not implemented") }

// Cancel remove uma ordem do book por ID
func (ob *OrderBook) Cancel(orderID string) bool { panic("not implemented") }

// BidDepth retorna o número de ordens ativas no lado bid
func (ob *OrderBook) BidDepth() int { panic("not implemented") }

// AskDepth retorna o número de ordens ativas no lado ask
func (ob *OrderBook) AskDepth() int { panic("not implemented") }
