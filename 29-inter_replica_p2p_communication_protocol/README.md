
# INTER-REPLICA COMMUNICATION PROTOCOL

**Category**: Distributed
**Time**: 3h
**Builds on**: distributed_matching_engine (order book, single-node) + stream broker binary framing (tcp-multiplexed-stream-broker)

## Study before (10-15min):

### Simple replication vs consensus

Why doesn't the order book need full Raft to have acceptable consistency?
Core idea: write quorum (W) over a set of N replicas guarantees any
subsequent read with read quorum (R) sees the latest data, as long as
R+W>N.
Think about replication granularity: the **full trade log** (every
execution, ordered) or **periodic book snapshot**. Trade log is cheaper to
propagate and gives deterministic replay; snapshot is simpler to apply but
heavier. This needs to be decided before coding.

## Context

The order book runs in a single process so far. To become truly
distributed, nodes need to exchange messages with each other, who's
primary, who replicates, when a write is considered "committed". Today you
build only the channel and the inter-replica message protocol, not the
quorum decision logic yet (that's the next challenge).

What to build:

1. Reuse the length-prefix framing (4 bytes size + payload) already used in
   the stream broker as the base of the inter-replica protocol, don't
   reinvent framing
2. Three message types: WriteProposal (a trade log operation, e.g. new
   order, cancellation, serialized), WriteAck (empty for now, referencing a
   specific proposal is deferred until quorum logic exists), Heartbeat
   (liveness, no payload, sender identity comes from the connection itself,
   timestamp is stamped by the receiver, not sent)
3. Each node keeps a persistent TCP connection with the other cluster nodes
   (fixed topology via config, no dynamic discovery yet)
4. `SetNoDelay(true)` on every new connection, before any I/O, this is a
   requirement, not a bonus, since the protocol does its own framing
5. A node can send WriteProposal to others and receive WriteAck back, and
   all nodes exchange Heartbeat at a fixed interval (e.g. every 200ms)

## Required:

- Framing reused from the stream broker, not a new version
- Serialization of the 3 message types (struct + encoding/gob, or manual
  binary, your choice, but document why)
- One goroutine owning the connection to each peer (never multiple
  goroutines writing to the same conn without coordination)
- Duplicate connection handling: with symmetric dialing, each pair of nodes
  ends up with two physical connections. Resolved today with a provisional
  lexicographic dial rule (only the node with the "smaller" address dials);
  the real fix (identity handshake on connect) is deferred to the next
  challenge
- Integration test with `net.Pipe()` or real TCP on 127.0.0.1:0 simulating
  2-3 nodes exchanging the 3 messages — **pending, carried over**
- Clean run with -race — **pending, carried over**

**Bonus (if time allows)**:
- Heartbeat timeout marking a peer as suspect (doesn't drop it yet, just
  logs), this becomes the base for failover in another challenge

## What will be observed
- Whether the replication granularity decision (trade log vs snapshot) was
  thought through before coding or decided mid-way
- Whether framing was actually reused or reinvented
- Whether goroutine-per-peer avoids mixing mutex and channel protecting the
  same thing
- Whether the connection identity issue (duplicate connections, ephemeral
  RemoteAddr on accepted conns) was noticed and handled, even provisionally

Design decisions: document in the README (or a comment at the top of the
main file) the choice between trade log and snapshot, and why. Also
document the lexicographic dial workaround and the plan to replace it with
identity handshake in the next challenge.

---
Next challenge preview: quorum decision logic (counting WriteAck per
proposal, deciding when a write is "committed"), plus connection identity
handshake to replace the lexicographic workaround.