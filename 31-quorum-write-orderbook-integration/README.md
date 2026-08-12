# CHALLENGE 31: QUORUM WRITE + ORDERBOOK INTEGRATION

**Category**: Distributed
**Time**: 2h
**Builds on**: 30-multi_node_p2p_connection_identity + distributed_matching_engine (orderbook)

## Study before (10-15min)

Quick review of write quorum (W): with N nodes in the cluster, W = (N/2)+1
guarantees a majority of votes. A proposal is only considered "committed"
once the proposer has received WriteAck from W-1 other nodes (the node
itself already counts as 1 implicit vote). Think about: what happens if a
node never responds (crashed, slow network)? The proposal needs a
timeout, without one, a pending proposal lives forever waiting for a vote
that never comes.

## Context

Today cluster and orderbook are two worlds that don't touch. WriteProposal
arrives, cluster sends an empty ack, nobody applies anything. This
challenge connects the two: a local proposal becomes a real WriteProposal
to the cluster before applying; a remote proposal only applies to the
local orderbook after reaching quorum.

What to build:

1. PendingProposal: struct tracking a proposal in flight, ID
   (nodeAddr:counter), the operation itself (serialized), how many acks
   received so far, from which peers, creation timestamp (for timeout)
2. Cluster.Propose(op Operation) local entrypoint: generates a new ID
   (increments an internal counter), registers a PendingProposal,
   broadcasts WriteProposal to every peer, counts its own vote
   immediately
3. handleInboundMsg, case WriteProposal: on receiving a proposal from
   another node, responds with WriteAck referencing the received ID (no
   longer empty like today), but does NOT apply to the orderbook yet,
   only confirms receipt
4. handleInboundMsg, case WriteAck: matches the ack to the
   PendingProposal by ID, increments that proposal's vote count. If
   quorum was reached (calculated dynamically from len(peers)+1), applies
   the operation to the local orderbook AND removes the PendingProposal
   from tracking
5. Pending proposal timeout: a goroutine (or ticker in Bootstrap) that
   periodically sweeps old PendingProposals (e.g. >5s without quorum),
   logs it as a failure, removes it from tracking, doesn't leave the
   cluster waiting forever

## Required

- Proposal ID is nodeAddr:counter, unique per proposal, never reused
- Quorum is calculated as (len(peers)+1)/2 + 1, never hardcoded to the
  number 3
- PendingProposals protected by its own mutex (concurrent access: an ack
  from a different peer can arrive at the same time the timeout sweep
  runs)
- A duplicate ack from the same peer for the same proposal doesn't count
  as two votes (idempotency, a peer only votes once per proposal, even if
  it resends)
- Timeout removes the proposal without crashing anything, just logs

## Bonus (if time allows)

When a proposal expires by timeout, decide whether it's worth
automatically resending or just reporting failure back to the original
caller, document the choice, no need to implement full retry today

## What will be observed

Whether quorum is calculated, not fixed; whether a duplicate ack is
handled correctly; whether the timeout exists and works (a proposal that
never reaches quorum doesn't leak memory or hang anything)

**Design decisions**

Document in the README the reasoning behind the ID scheme
(nodeAddr:counter) and why quorum is calculated instead of fixed, even
with only 3 nodes today, this is the pattern real systems use so the
logic doesn't need rewriting when the cluster grows.

---

First step: before any code, think about the shape of PendingProposal, it
needs to know "which peers have already voted", not just "how many
votes". Why does this matter, given the idempotency requirement (a
duplicate ack must not count twice)? What Go data structure naturally
solves "set of peers that have already voted"?