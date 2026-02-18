package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db := NewDB()
	defer db.conn.Close()

	service := NewPaymentService(db, 4) // 4 shards

	// Primeira requisição — processa
	result1, err := service.ProcessPayment(context.Background(), &PaymentRequest{
		IdempotencyKey: "key-123",
		UserID:         "alice",
		Amount:         10000,
		Currency:       "USD",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Payment 1: ID=%d Status=%s\n", result1.PaymentID, result1.Status)

	// Segunda requisição com MESMA chave — retorna resultado anterior
	result2, err := service.ProcessPayment(context.Background(), &PaymentRequest{
		IdempotencyKey: "key-123",
		UserID:         "alice",
		Amount:         10000,
		Currency:       "USD",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Payment 2 (retry): ID=%d Status=%s\n", result2.PaymentID, result2.Status)

	if result1.PaymentID == result2.PaymentID {
		fmt.Println("Idempotência garantida: mesma chave, mesmo resultado")
	}

	// Terceira requisição com chave diferente — cria novo pagamento
	result3, err := service.ProcessPayment(context.Background(), &PaymentRequest{
		IdempotencyKey: "key-456",
		UserID:         "alice",
		Amount:         5000,
		Currency:       "USD",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Payment 3: ID=%d Status=%s\n", result3.PaymentID, result3.Status)
}
