# Idempotent Payment Processing (PostgreSQL)

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Distributed Systems
**Estimated time:** ~2 hours

## What it is

A payment processing service that guarantees the same idempotency key never results in a double charge, even under concurrent retries, by sharding requests by key and processing each shard sequentially.

## What you'll learn

- Why idempotency matters for payment APIs: network timeouts and server crashes mean the client will retry, and retries must be safe.
- Sharding by idempotency key as an alternative to a single global advisory lock, so unrelated keys can process in parallel while identical keys are always serialized.
- Using `ON CONFLICT DO NOTHING` at the database layer as the final safety net beneath the in-process sharding.

## What's implemented

- `NewDB()` and `createTables(db *sql.DB) error` setting up the PostgreSQL schema.
- `NewPaymentService(db *DB, numShards int) *PaymentService`.
- `ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error)` as the public entry point.
- `hashKey(key string, shards int) int64` routing a request to a fixed shard based on its idempotency key.
- `Shard.worker(db *DB)`: one goroutine per shard, processing that shard's requests one at a time.
- `processStripeCharge(...)` as a mocked charge call (no real Stripe integration).
- Tests cover idempotency for a single key, concurrent requests with the same key, different keys processing in parallel, and multiple distinct keys.

## Design decisions

- Each shard has exactly one worker goroutine, so requests with the same idempotency key are naturally serialized without needing a per-key lock.
- `hashKey` is deterministic, so the same key always lands on the same shard across retries.
- The database `ON CONFLICT DO NOTHING` constraint is the last line of defense in case two requests with the same key somehow reach the database concurrently.

## How to run

Requires a running PostgreSQL instance (connection details in `db.go`).

```bash
make execute
# or
go run .
go test ./...
```
