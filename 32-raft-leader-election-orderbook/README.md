# CHALLENGE 32: RAFT LEADER ELECTION

**Category**: Distributed
**Time**: 2h
**Builds on**: 31-quorum-write-orderbook-integration

## Study before

Every node is in one of 3 states: Follower, Candidate, Leader. Every
node has a "term" (a number that only ever increases, never resets). A
Follower becomes a Candidate if it doesn't hear a heartbeat from the
leader within a random timeout (randomized specifically to avoid two
nodes becoming candidates at the same time and repeatedly tying the
election). A Candidate increments its own term, votes for itself, sends
RequestVote to the others, and whoever receives it only votes if it
hasn't voted in that term yet AND the candidate's term is equal to or
greater than its own. Winning a majority of votes = becomes Leader,
starts sending heartbeats (your existing Heartbeat today already exists,
but doesn't carry a term yet nor does it reset anyone's election timeout).

## Context

Today any node can propose at any moment, and that's what causes the
divergence documented in challenge 31 (two concurrent proposals from
different nodes, with no guaranteed total order). Leader election is the
first step toward fixing that: only the elected leader will propose,
once that's in place (that restriction itself is the next challenge,
today is just the election, nobody is prevented from proposing yet).

What to build:

1. Node state: NodeState (Follower/Candidate/Leader), CurrentTerm
   (uint64, persisted in memory only for now), VotedFor (identity of who
   this node voted for in this term, or empty)
2. Election timeout: a timer with a random duration (e.g. 150-300ms)
   that resets every time a valid heartbeat arrives. If it fires, the
   node becomes a Candidate
3. Two new message types: MsgRequestVote (candidate's term, candidate's
   identity) and MsgVoteGranted (term, granted bool)
4. Candidate logic: increments term, votes for itself, broadcasts
   RequestVote, counts votes received, becomes Leader if it reaches a
   majority, steps back to Follower if it receives a heartbeat with a
   term equal to or greater from another node (someone already won)
5. The existing Heartbeat gains a term field. A Follower that receives a
   heartbeat with term >= its own resets its election timeout and
   recognizes the sender as the current leader

## Required

- VotedFor is reset on every new term (a node can vote again once the
  term changes)
- A node never votes twice in the same term
- Split vote (nobody reaches a majority, e.g. 3 nodes, 3 candidates at
  the same time) resolves on its own, because the next round's random
  timeout won't be the same for everyone again, document this in the
  test, no need to force an artificial scenario, just confirm that
  re-election happens if the first round ties
- Integration test: kills the current leader's process (or simulates a
  timeout), confirms a new leader is elected among the 2 survivors, with
  a term greater than the previous one

What will be observed: whether term is treated as the source of truth
for deciding who's in charge (higher term always wins, no matter who
"arrived first" on the network); whether the randomized timeout actually
varies per node (a fixed timeout, the same for everyone, would
reintroduce an infinite tie)

---

First step, to think about before coding: today your MessageType already
has Hello/Proposal/Ack/Heartbeat/Commit. Do RequestVote and VoteGranted
go into that same enum (same pattern that already applies to Hello)? And
the Term field, would you add it to EVERY existing message type
(Proposal, Ack, Heartbeat, Commit), or only to the two new ones
(RequestVote/VoteGranted) for now? Think about why: a Follower receiving
a message of any type coming from a term GREATER than its own needs to
know that in order to update itself, does that change your answer?