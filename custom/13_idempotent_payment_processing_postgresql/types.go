package main

import (
	"database/sql"
	"time"
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

type PaymentModel struct {
	id             int64
	idempotencyKey string
	userID         string
	amount         int64
	currency       string
	status         string
	stripeChargeID string
	errorMessage   string
	createdAt      time.Time
	updatedAt      time.Time
}

type Shard struct {
	id    int
	reqCh chan *paymentTask
}

type paymentTask struct {
	req    *PaymentRequest
	result chan *paymentTaskResult
}

type paymentTaskResult struct {
	result *PaymentResult
	err    error
}

type PaymentService struct {
	db     *sql.DB
	shards []*Shard
}
