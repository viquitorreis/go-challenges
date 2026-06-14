# Go Systems Challenges

A collection of 15 self designed, senior level Go challenges built to deepen practical knowledge of concurrency, data structures, network programming, and distributed systems patterns the kind of problems that show up in real
backend systems and in interviews at infrastructure-focused companies.

I built these for myself because I couldn't find resources that went beyond "here's how goroutines and channels work" into the trade-offs that actually matter in production: lock granularity, backpressure, graceful shutdown, idempotency under retries, consensus. Sharing them in case they're useful for other engineers preparing for the same kind of interviews.
 
Each challenge includes a problem statement, a starting skeleton with TODOs, and a test suite (most runnable with `go test -race`). Pick one, implement it, and run the tests. If you write up your solution or learnings, feel free to share happy to see how others approached the same problems.

## Challenges
 
| # | Challenge | Core concepts |
|---|-----------|----------------|
| 1 | [Event Bus (Fan-out)](./1_event-bus-fan-out) - [problem statement](./README_FULL_NOTES.md#1-event-bus) | Pub/sub fan-out, per-subscriber channels, `sync.RWMutex` |
| 2 | [Log Aggregator (Fan-in)](./2_log-aggregator-fan-in) - [problem statement](./README_FULL_NOTES.md#2-log-aggregator-fan-in) | N:1 fan-in, `WaitGroup` coordination, layered graceful shutdown |
| 3 | [Image Processing Pipeline](./3_image-processor-pipeline) - [problem statement](./README_FULL_NOTES.md#3-image-processor-pipeline) | 4-stage concurrent pipeline, context cancellation |
| 4 | [TCP Chat Server](./4_tcp-chat-server) - [problem statement](./README_FULL_NOTES.md#4-tcp-chat-server) | Raw TCP, `sync.Cond` lobby, per-client broadcast channels |
| 5 | [Rate Limiter (Token Bucket)](./5_rate_limiter_token_bucket) - [problem statement](./README_FULL_NOTES.md#5-rate-limiter-com-token-bucket) | Token bucket algorithm, lazy refill, concurrent-safe |
| 6 | [Worker Pool with Priority Queue](./6_worker_pool_priority_queue) - [problem statement](./README_FULL_NOTES.md#6-worker-pool-com-priority-quieu) | `container/heap`, `sync.Cond`, bounded queue with backpressure |
| 7 | [LRU Cache with TTL](./7_lru_cache_thread_safe_with_ttl) - [problem statement](./README_FULL_NOTES.md#7-lru-cache-thread-safe-com-ttl) | HashMap + doubly linked list, O(1) eviction, active TTL cleanup |
| 8 | [Health Check Poller with Circuit Breaker](./8_health_check_poller_with_circuit_breaker) - [problem statement](./README_FULL_NOTES.md#8---health-check-poller-com-circuit-breaker) | Concurrent HTTP polling, circuit breaker, status aggregation |
| 9 | [Thread-Safe Trie](./9_trie_thread_safe) - [problem statement](./README_FULL_NOTES.md#9-trie-thread-safe) | Autocomplete trie, fine-grained per-node (hand-over-hand) locking |
| 10 | [Skip List](./10_skip_list_thread_safe) - [problem statement](./README_FULL_NOTES.md#10-skip-list-thread-safe) | Probabilistic skip list (Redis sorted-set style), update-array pattern |
| 11 | [Exchange Order Book](./11_exchange_order_book) - [problem statement](./README_FULL_NOTES.md#11-exchange-order-book) | Price-time priority matching engine, bid/ask order book |
| 12 | [TCP Server with Worker Pool & Backpressure](./12_tcp_server_worker_pools) - [problem statement](./README_FULL_NOTES.md#12-tcp-server-com-worker-pool-e-backpressure) | Bounded connection queue, graceful shutdown, load shedding |
| 13 | [Idempotent Payment Processing (PostgreSQL)](./13_idempotent_payment_processing_postgresql) - [problem statement](./README_FULL_NOTES.md#13-idempotent-payment-processing-with-postgresql) | Sharded idempotency keys, `ON CONFLICT DO NOTHING`, concurrent dedup |
| 14 | [Mining Pool with Stratum Protocol](./14_mining_pool_with_stratum_protocol) - [problem statement](./README_FULL_NOTES.md#14-mining-pool-with-stratum-protocol) | JSON-RPC over TCP, marketplace vs. order-book pricing models |
| 15 | [Raft Leader Election](./15_raft_leader_election) - [problem statement](./README_FULL_NOTES.md#15-raft-leader-election) | Raft consensus, terms, randomized election timeouts, heartbeats |
 
## Why this exists

Most Go concurrency resources stop at "here's how goroutines and channels
work." There wasn't much available that pushed into the trade-offs that
matter in production: lock granularity, backpressure, graceful shutdown
ordering, idempotency under concurrent retries, consensus. This repo is the
curriculum I built for myself to close that gap - one problem at a time,
each with a clear set of learnings and a test suite that enforces correctness
with `-race`.
