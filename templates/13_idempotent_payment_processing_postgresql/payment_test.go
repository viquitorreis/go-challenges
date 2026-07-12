package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres",
		"host=localhost user=postgres password=postgres dbname=payments_test sslmode=disable")
	require.NoError(t, err)

	_, err = db.Exec(`DROP TYPE IF EXISTS payment_status CASCADE`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed')`)
	require.NoError(t, err)
	_, err = db.Exec(`DROP TABLE IF EXISTS payments`)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE payments (
			id BIGSERIAL PRIMARY KEY,
			idempotency_key VARCHAR(255) NOT NULL UNIQUE,
			user_id VARCHAR(255) NOT NULL,
			amount BIGINT NOT NULL,
			currency VARCHAR(3) NOT NULL,
			status payment_status NOT NULL,
			stripe_charge_id VARCHAR(255),
			error_message VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	return db
}

// TestIdempotency — mesma chave retorna mesmo resultado
func TestIdempotency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewPaymentService(db, 4)
	ctx := context.Background()

	req := &PaymentRequest{
		IdempotencyKey: "idempotent-key-1",
		UserID:         "alice",
		Amount:         10000,
		Currency:       "USD",
	}

	result1, err := service.ProcessPayment(ctx, req)
	require.NoError(t, err)
	assert.NotZero(t, result1.PaymentID)
	assert.Equal(t, "completed", result1.Status)

	result2, err := service.ProcessPayment(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, result1.PaymentID, result2.PaymentID)
	assert.Equal(t, result1.Status, result2.Status)
	assert.Equal(t, result1.StripeChargeID, result2.StripeChargeID)
}

// TestConcurrentIdempotency — 10 goroutines com mesma chave não duplicam
func TestConcurrentIdempotency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewPaymentService(db, 4)
	ctx := context.Background()

	req := &PaymentRequest{
		IdempotencyKey: "concurrent-key-1",
		UserID:         "bob",
		Amount:         5000,
		Currency:       "USD",
	}

	results := make([]*PaymentResult, 10)
	errs := make([]error, 10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = service.ProcessPayment(ctx, req)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "request %d failed", i)
	}

	firstID := results[0].PaymentID
	for i, result := range results {
		assert.Equal(t, firstID, result.PaymentID,
			"request %d got different payment ID", i)
	}

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM payments WHERE idempotency_key = $1`,
		req.IdempotencyKey).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have exactly 1 payment in DB")
}

// TestDifferentKeysParallel — chaves diferentes processam em paralelo, sem bloquear
func TestDifferentKeysParallel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewPaymentService(db, 4)
	ctx := context.Background()

	results := make([]*PaymentResult, 10)
	errs := make([]error, 10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &PaymentRequest{
				IdempotencyKey: fmt.Sprintf("key-%d", idx),
				UserID:         fmt.Sprintf("user-%d", idx),
				Amount:         10000,
				Currency:       "USD",
			}
			results[idx], errs[idx] = service.ProcessPayment(ctx, req)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "request %d failed", i)
	}

	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			assert.NotEqual(t, results[i].PaymentID, results[j].PaymentID,
				"different keys should create different payments")
		}
	}
}
