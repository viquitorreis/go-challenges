# CHALLENGE 30: CONNECTION IDENTITY + INTEGRATION TEST + MULTI-NODE SPAWN
**Category**: Distributed
**Time**: 2h
**Builds on**: 29-inter_replica_communication_protocol

## Context
Challenge 29 closed the message channel (WriteProposal/WriteAck/Heartbeat)
but left 3 debts:
- Connection identity based on ephemeral RemoteAddr() (breaks on accepted connections)
- No formal integration test
- No validation of N real nodes running and communicating.
This challenge closes all 3, without getting into quorum logic (that's challenge 31).

**What to build**:
1. Identity handshake: the first message exchanged on any new connection
   (before Proposal/Ack/Heartbeat) announces "I'm node X, my listen
   address is Y". Replaces the provisional lexicographic dial rule, with
   real identity, a duplicate connection between the same pair of nodes
   can be detected and closed explicitly, instead of avoided by convention
2. Integration test: net.Pipe() or real TCP on 127.0.0.1:0 simulating
   2-3 nodes completing the handshake and exchanging the 3 messages
3. Clean -race across the whole suite
4. Spawn of 3+ real processes (not simulated in a test) running the
   binary, confirming handshake + heartbeat flowing between all of them

## Required:
- Handshake becomes the first case handled in the peer's message
  dispatch, before any other logic
- Duplicate connection between the same pair (same identity on both
  sides) detected and one of the two closed, with a log explaining which
  one survived
- Integration test covers: handshake completes, peer registered in the
  cluster with the correct identity (not RemoteAddr), heartbeat MarkAlive
  triggered
- 3+ real nodes come up, connect, handshake completes on every pair, no
  duplicate connection left over

What will be observed: whether identity solves both problems (ephemeral
RemoteAddr AND duplicate connection) or just one of the two; whether the
integration test actually tests the handshake, not just the messages
that already existed

---
First step: does MsgHello go into the same MessageType enum that already
exists (WriteProposal/WriteAck/Heartbeat), or is the handshake a separate
protocol, outside that type+body envelope? Think about how Peer.readLoop
dispatches today (reads type, switch) before deciding.