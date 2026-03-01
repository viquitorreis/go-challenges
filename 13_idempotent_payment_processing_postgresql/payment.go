package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"time"
)

// said you'd shard requests to workers using idempotencyKey % numWorkers to avoid database locks.
// But what happens when a worker crashes mid-processing? What happens when you need to scale and
// add more workers - now the same idempotency key routes to a different worker.
// What happens during deployment when workers restart? Your approach has no persistence,
// no recovery mechanism, and breaks under the exact failure scenarios the interviewer cares about.

func NewPaymentService(db *DB, numShards int) *PaymentService {
	// cada shard tem sua goroutine worker que processa a fila
	ps := &PaymentService{
		db:     db.conn,
		shards: make([]*Shard, numShards),
	}

	for i := range numShards {
		shard := &Shard{
			id:    i,
			reqCh: make(chan *paymentTask),
		}
		ps.shards[i] = shard

		// spawn do worker
		go shard.worker(db)
	}

	return ps
}

func (ps *PaymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	shard := int(hashKey(req.IdempotencyKey, len(ps.shards)))

	// buffered channel, dessa forma cada worker processa no maximo um resultado por vez
	task := &paymentTask{
		req:    req,
		result: make(chan *paymentTaskResult, 1),
	}

	// envia task no channel de processamento desse shard para o worker processar depois
	ps.shards[shard].reqCh <- task

	// bloqueia até o worker processar e enviar resultado
	res := <-task.result

	return res.result, res.err
}

// a idempotencia dessa abordagem:

// mesma idempotencyKey -> sempre vai pro mesmo shard (via hash % numShards)
// esse shard processa sequencialmente (1 worker, 1 channel, FIFO)
// primeira requisição: SELECT (não existe) -> INSERT -> retorna resultado
// segunda requisição com mesma chave: SELECT (já existe) -> retorna resultado anterior, sem INSERT
func (s *Shard) worker(db *DB) {
	for task := range s.reqCh {
		ctx := context.Background()
		tx, err := db.conn.BeginTx(ctx, nil)
		if err != nil {
			slog.Error("error starting transaction", "error", err)
			task.result <- &paymentTaskResult{err: err}
			continue
		}

		var payment PaymentModel
		row := tx.QueryRow(`
			SELECT id, idempotency_key, user_id, amount, currency, status, stripe_charge_id, error_message, created_at, updated_at
			FROM payments 
			WHERE idempotency_key = $1
		`, task.req.IdempotencyKey)

		err = row.Scan(
			&payment.id,
			&payment.idempotencyKey,
			&payment.userID,
			&payment.amount,
			&payment.currency,
			&payment.status,
			&payment.stripeChargeID,
			&payment.errorMessage,
			&payment.createdAt,
			&payment.updatedAt,
		)

		if err == nil {
			tx.Commit()
			task.result <- &paymentTaskResult{
				result: &PaymentResult{
					PaymentID:      payment.id,
					Status:         payment.status,
					StripeChargeID: payment.stripeChargeID,
					ErrorMessage:   payment.errorMessage,
				},
			}
			continue
		}

		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("error selecting payment", "error", err)
			tx.Rollback()
			task.result <- &paymentTaskResult{err: err}
			continue
		}

		chargeID, err := processStripeCharge(ctx, task.req.Amount, task.req.UserID)
		if err != nil {
			tx.Rollback()
			task.result <- &paymentTaskResult{err: err}
			continue
		}

		var newPaymentID int64
		var newStatus string
		var newChargeID string
		var newErrorMsg string

		row = tx.QueryRow(`
			INSERT INTO payments (idempotency_key, user_id, amount, currency, status, stripe_charge_id, error_message)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (idempotency_key) DO NOTHING
			RETURNING id, status, stripe_charge_id, error_message
		`, task.req.IdempotencyKey, task.req.UserID, task.req.Amount, task.req.Currency, "completed", chargeID, "")

		err = row.Scan(&newPaymentID, &newStatus, &newChargeID, &newErrorMsg)

		if errors.Is(err, sql.ErrNoRows) {
			row = tx.QueryRow(`
				SELECT id, status, stripe_charge_id, error_message
				FROM payments
				WHERE idempotency_key = $1
			`, task.req.IdempotencyKey)
			err = row.Scan(&newPaymentID, &newStatus, &newChargeID, &newErrorMsg)
		}

		if err != nil {
			slog.Error("error inserting payment", "error", err)
			tx.Rollback()
			task.result <- &paymentTaskResult{err: err}
			continue
		}

		tx.Commit()
		task.result <- &paymentTaskResult{
			result: &PaymentResult{
				PaymentID:      newPaymentID,
				Status:         newStatus,
				StripeChargeID: newChargeID,
				ErrorMessage:   newErrorMsg,
			},
		}
	}
}

func hashKey(key string, shards int) int64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int64(h.Sum64() % uint64(shards))
}

func processStripeCharge(ctx context.Context, amount int64, userID string) (chargeID string, err error) {
	time.Sleep(10 * time.Millisecond)
	chargeID = fmt.Sprintf("ch_%s_%d", userID, time.Now().UnixNano())
	return chargeID, nil
}
