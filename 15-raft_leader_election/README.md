# Raft Leader Election

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Distributed Systems
**Estimated time:** ~2 hours

## What it is

An implementation of the leader election portion of the Raft consensus algorithm, the same algorithm behind etcd, Kafka's KRaft mode, Consul, CockroachDB, and TiKV: nodes hold an election with randomized timeouts, and the winner sends heartbeats to hold onto leadership.

## What you'll learn

- The three Raft roles (Follower, Candidate, Leader) and the state transitions between them.
- Why election timeouts are randomized per node, to avoid every node starting a campaign simultaneously and splitting the vote indefinitely.
- Terms as a logical clock: every election increments the term, and any message carrying an older term is ignored, which prevents a stale or recovered node from disrupting a live leader.

## What's implemented

- `NewCluster(size int) *Cluster`, `Start()`, `GetLeader() *Node`, `KillLeader()`, `Stop()` to drive a simulated cluster.
- `NewNode(id int, peers []*Node) *Node` with `run()` dispatching to `runFollower()`, `runCandidate()`, and `runLeader()` based on current role.
- `sendHeartbeats()` for the leader to maintain authority and reset followers' election timers.
- Tests cover a single leader being elected, all nodes agreeing on the term, leader failover after `KillLeader()`, absence of race conditions, and a node stepping down when it sees a higher term.

## Design decisions

- Each node runs its own state machine loop (`run()`) reacting to timers and incoming messages, rather than a central coordinator driving the cluster.
- Election timeouts are randomized per node specifically to break symmetry and make split votes rare in practice, matching the real Raft paper's approach.

## How to run

```bash
go run .
go test -race ./...
```
