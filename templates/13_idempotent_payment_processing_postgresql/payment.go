package main

import (
	"context"
	"database/sql"
	"hash/fnv"
)

type PaymentRequest struct {
	IdempotencyKey string
	UserID         string
	Amount         int64
	Currency       string
}

type PaymentResult struct {
	PaymentID      int64
	Status         string
	StripeChargeID string
	ErrorMessage   string
}

type PaymentService struct {
	db     *sql.DB
	shards []*Shard
}

// Shard processa pagamentos sequencialmente para um conjunto de chaves
type Shard struct {
	id    int
	reqCh chan *paymentTask
	// TODO: adicionar campos necessários
}

type paymentTask struct {
	req    *PaymentRequest
	result chan *paymentTaskResult
}

type paymentTaskResult struct {
	result *PaymentResult
	err    error
}

func NewPaymentService(db *sql.DB, numShards int) *PaymentService {
	// TODO: inicializar service com N shards
	// cada shard tem sua goroutine worker que processa a fila
	panic("not implemented")
}

// ProcessPayment enfileira um pagamento no shard correto baseado no hash da chave
// Retorna o resultado (novo ou anterior se já processado)
func (ps *PaymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	// TODO:
	// 1. hash(idempotencyKey) % numShards -> shard index
	// 2. Enfileira requisição no shard via channel
	// 3. Aguarda resultado do shard
	panic("not implemented")
}

// shardWorker processa requisições sequencialmente para um shard
// Cada requisição: SELECT + INSERT ON CONFLICT DO NOTHING
func (s *Shard) worker(db *sql.DB) {
	// TODO:
	// loop infinito: recebe task do s.reqCh
	// para cada task:
	//   1. SELECT * FROM payments WHERE idempotency_key = ? -> já existe?
	//      SIM: retorna resultado anterior
	//      NÃO: continua
	//   2. Processa pagamento (simula Stripe)
	//   3. INSERT ... ON CONFLICT DO NOTHING
	//   4. SELECT * para recuperar o resultado inserido
	//   5. Envia resultado via task.result
	panic("not implemented")
}

func hashKey(key string, shards int) int64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int64(h.Sum64() % uint64(shards))
}

func processStripeCharge(ctx context.Context, amount int64, userID string) (chargeID string, err error) {
	// TODO: simular processamento da Stripe
	// retorna um charge ID fictício e nil error
	panic("not implemented")
}
