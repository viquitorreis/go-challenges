package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"raft_orderbook/orderbook"
	"raft_orderbook/peer"
	"sync"
	"sync/atomic"
)

type Cluster struct {
	peers map[string]*peer.Peer
	mu    sync.RWMutex

	register   chan *peer.Peer
	unregister chan *peer.Peer

	inbound chan peer.InboundMsg // peer.Peer develivers here what have read from the network
	propose chan []byte          // local ask to propose a write on the cluster

	PendingProposals *PendingProposals
	proposalCounter  atomic.Uint64
	ownAddr          string // needs to know its own address to create the ID
	ob               *orderbook.OrderBook
}

func NewCluster(ctx context.Context, ownAddr string, ob *orderbook.OrderBook) *Cluster {
	cluster := &Cluster{
		peers:            make(map[string]*peer.Peer),
		register:         make(chan *peer.Peer),
		unregister:       make(chan *peer.Peer),
		inbound:          make(chan peer.InboundMsg),
		propose:          make(chan []byte),
		ownAddr:          ownAddr,
		PendingProposals: NewPendingProposals(),
		ob:               ob,
	}

	go cluster.Bootstrap(ctx)

	return cluster
}

func (c *Cluster) Bootstrap(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-c.register:
			if c.TryRegister(p.Addr, p) {
				slog.Info("peer registered after handshake", "peer", p.Addr)
			} else {
				slog.Warn("duplicate connection to already-registered peer, closing", "peer", p.Addr)
			}
		case p := <-c.unregister:
			c.UnregPeer(p)
		case msg := <-c.inbound:
			c.handleInboundMsg(msg)
		case payload := <-c.propose:
			c.broadcastProposal(payload)
		}
	}
}

func (c *Cluster) TryRegister(identity string, p *peer.Peer) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.peers[identity]; exists {
		return false
	}

	c.peers[identity] = p

	return true
}

func (b *Cluster) UnregPeer(p *peer.Peer) {
	b.mu.Lock()
	delete(b.peers, p.Addr)
	b.mu.Unlock()
}

func (c *Cluster) handleInboundMsg(m peer.InboundMsg) {
	switch m.Type {
	case peer.MsgWriteProposal:
		id, op, err := decodeProposal(m.Body) // extrai ID + operação do payload
		if err != nil {
			slog.Error("malformed proposal", "error", err, "from", m.From.Addr)
			return
		}

		// receiving side now DOES track it, so it has the operation
		// ready when the commit message arrives later
		c.PendingProposals.Register(id, op)
		m.From.Send(peer.MsgWriteAck, []byte(id)) // ack now references the real id

	case peer.MsgWriteAck:
		id := string(m.Body)

		_, exists := c.PendingProposals.RecordVote(id, m.From.Addr)
		if !exists {
			// proposal already commited, timed out or acked
			// something this node never proposed so we ignore, its not an errors
			return
		}

		c.maybeCommit(id)

	case peer.MsgCommit:
		id := string(m.Body)
		c.applyCommitted(id) // new: apply locally, same logic maybeCommit used to inline

	case peer.MsgHeartbeat:
		m.From.MarkAlive()

	case peer.MsgHello:

	}
}

func (c *Cluster) broadcastProposal(payload []byte) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.peers {
		p.Send(peer.MsgWriteProposal, payload)
	}
}

func (c *Cluster) InboundChan() chan<- peer.InboundMsg {
	return c.inbound
}

func (c *Cluster) Register(p *peer.Peer) {
	c.register <- p
}

// Propose originates a new write proposal from THIS node. Generates a
// unique ID, tracks it as pending, broadcasts to every peer, and counts this node's
// own implicit vote immediately (it doesnt need to ack itself over the network)
func (c *Cluster) Propose(op []byte) string {
	counter := c.proposalCounter.Add(1)
	id := fmt.Sprintf("%s:%d", c.ownAddr, counter)

	c.PendingProposals.Register(id, op)
	// count our own vote right away, quorum counts us as 1 without
	// needing a round-trip ack to ourselves
	c.PendingProposals.RecordVote(id, c.ownAddr)

	payload, err := encodeProposal(id, op)
	if err != nil {
		slog.Error("failed to encode proposal", "error", err)
		return "" // err signaling??
	}

	c.mu.RLock()
	for _, p := range c.peers {
		p.Send(peer.MsgWriteProposal, payload)
	}
	c.mu.RUnlock()

	c.maybeCommit(id) // in case it reached quorum with only its own vote (cluster with only 1 node)

	return id
}

// maybeCommit checks where a proposal has reached quorum, and if so,
// applies it to the local order book and stops tracking it
func (c *Cluster) maybeCommit(id string) {
	c.mu.RLock()
	quorum := (len(c.peers)+1)/2 + 1 // +1 counting this node on total this clusters have
	c.mu.RUnlock()

	// we need to expose this actual vote without incrementing the new one, adjust in
	// PendingProposals: a read only method, like VoteCount(id)
	votes := c.PendingProposals.VoteCount(id)
	if votes < quorum {
		return
	}

	// quorum reached broadcast commit so every peer applies too
	c.mu.RLock()
	for _, p := range c.peers {
		p.Send(peer.MsgCommit, []byte(id))
	}
	c.mu.RUnlock()

	c.applyCommitted(id) // apply on the proposer itself too
}

func (c *Cluster) applyCommitted(id string) {
	pp := c.PendingProposals.Get(id)
	if pp == nil {
		return // already applied, or this node never tracked it
	}

	op, err := orderbook.DecodeOperation(pp.Operation)
	if err != nil {
		slog.Error("failed to decode committed operation", "id", id, "error", err)
		c.PendingProposals.Remove(id)
		return
	}

	c.ob.Apply(op)
	trades := c.ob.Match()
	c.PendingProposals.Remove(id)
	slog.Info("proposal committed", "id", id, "trades", len(trades))
}

func (c *Cluster) GetOrders() []orderbook.Order {
	return c.ob.AllOrders()
}
