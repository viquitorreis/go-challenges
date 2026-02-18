package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) *PaymentService {
	db := NewDB()

	_, err := db.conn.Exec(`TRUNCATE payments`)
	require.NoError(t, err)

	return NewPaymentService(db, 4)
}

// TestIdempotency — mesma chave retorna mesmo resultado
func TestIdempotency(t *testing.T) {
	service := setupTest(t)
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
	service := setupTest(t)
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
	err := service.db.QueryRow(
		`SELECT COUNT(*) FROM payments WHERE idempotency_key = $1`,
		req.IdempotencyKey,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have exactly 1 payment in DB")
}

// TestDifferentKeysParallel — chaves diferentes processam em paralelo
func TestDifferentKeysParallel(t *testing.T) {
	service := setupTest(t)
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

	// Verifica que criou 10 pagamentos diferentes
	var count int
	err := service.db.QueryRow(`SELECT COUNT(*) FROM payments`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 10, count, "should have 10 different payments")

	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			assert.NotEqual(t, results[i].PaymentID, results[j].PaymentID,
				"different keys should create different payments")
		}
	}
}

// TestMultipleDifferentKeys — chaves diferentes criam pagamentos diferentes
func TestMultipleDifferentKeys(t *testing.T) {
	service := setupTest(t)
	ctx := context.Background()

	req1 := &PaymentRequest{
		IdempotencyKey: "key-1",
		UserID:         "alice",
		Amount:         10000,
		Currency:       "USD",
	}
	req2 := &PaymentRequest{
		IdempotencyKey: "key-2",
		UserID:         "alice",
		Amount:         10000,
		Currency:       "USD",
	}

	result1, err := service.ProcessPayment(ctx, req1)
	require.NoError(t, err)

	result2, err := service.ProcessPayment(ctx, req2)
	require.NoError(t, err)

	assert.NotEqual(t, result1.PaymentID, result2.PaymentID)
}
