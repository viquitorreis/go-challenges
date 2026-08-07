package cluster

import (
	"context"
	"inter_comm/peer"
	"log"
	"sync"
)

type Cluster struct {
	peers map[string]*peer.Peer
	mu    sync.RWMutex

	register   chan *peer.Peer
	unregister chan *peer.Peer

	inbound chan peer.InboundMsg // peer.Peer develivers here what have read from the network
	propose chan []byte          // local ask to propose a write on the cluster
}

func NewCluster(ctx context.Context) *Cluster {
	cluster := &Cluster{
		peers:      make(map[string]*peer.Peer),
		register:   make(chan *peer.Peer),
		unregister: make(chan *peer.Peer),
		inbound:    make(chan peer.InboundMsg),
		propose:    make(chan []byte),
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
			c.RegPeer(p)
			log.Printf("REGISTERED: peer %s, total: %d", p.Addr, len(c.peers))
		case p := <-c.unregister:
			c.UnregPeer(p)
		case msg := <-c.inbound:
			c.handleInboundMsg(msg)
		case payload := <-c.propose:
			c.broadcastProposal(payload)
		}
	}
}

func (b *Cluster) RegPeer(p *peer.Peer) {
	b.mu.Lock()
	b.peers[p.Addr] = p
	b.mu.Unlock()
}

func (b *Cluster) UnregPeer(p *peer.Peer) {
	b.mu.Lock()
	delete(b.peers, p.Addr)
	b.mu.Unlock()
}

func (c *Cluster) handleInboundMsg(m peer.InboundMsg) {
	switch m.Type {
	case peer.MsgWriteProposal:
		// only response is ack for now... next challenge will have a real quorum
		m.From.Send(peer.MsgWriteAck, nil)
	case peer.MsgWriteAck:
		log.Printf("ack received from %s", m.From.Addr)
	case peer.MsgHeartbeat:
		m.From.MarkAlive()
		log.Printf("heartbeat from %s", m.From.Addr)
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
