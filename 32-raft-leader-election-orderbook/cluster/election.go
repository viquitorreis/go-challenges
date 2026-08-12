package cluster

import (
	"log/slog"
	"math/rand"
	"raft_orderbook/peer"
	"time"
)

const (
	electionTimeoutMin   = 150 * time.Millisecond
	electionTimeoutMax   = 300 * time.Millisecond
	leaderHeartbeatEvery = 75 * time.Millisecond // must be well under electionTimeoutMin
)

func randomElectionTimeout() time.Duration {
	return electionTimeoutMin + time.Duration(rand.Int63n(int64(electionTimeoutMax-electionTimeoutMin)))
}

func (c *Cluster) startElection() {
	term := c.raft.BecomeCandidate(c.ownAddr)
	slog.Info("starting election", "term", term, "candidate", c.ownAddr)

	c.mu.RLock()
	peerCount := len(c.peers)
	c.mu.RUnlock()
	slog.Info("starting election", "term", term, "candidate", c.ownAddr, "known_peers", peerCount)

	c.electionMu.Lock()
	c.votesReceived = map[string]struct{}{c.ownAddr: {}} // vote on itself
	c.electionMu.Unlock()

	payload, err := encodeRequestVote(term, c.ownAddr)
	if err != nil {
		slog.Error("failed to encode request vote", "error", err)
		return
	}

	c.mu.RLock()
	for _, p := range c.peers {
		p.Send(peer.MsgRequestVote, payload)
	}
	c.mu.RUnlock()

	c.resetElectionTimer() // if this round fails too, try again after another random wait
}

func (c *Cluster) resetElectionTimer() {
	if !c.electionTimer.Stop() {
		select {
		case <-c.electionTimer.C:
		default:
		}
	}

	c.electionTimer.Reset(randomElectionTimeout())
}

func (c *Cluster) broadcastLeaderHeartbeat() {
	payload, err := encodeLeaderHeartbeat(c.raft.CurrentTerm(), c.ownAddr)
	if err != nil {
		slog.Error("failed to encode leader heartbeat", "error", err)
		return
	}

	c.mu.RLock()
	for _, p := range c.peers {
		p.Send(peer.MsgLeaderHeartbeat, payload)
	}
	c.mu.RUnlock()
}
