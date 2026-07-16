package main

// Cluster gerencia o conjunto de nós
type Cluster struct {
	nodes []*Node
}

func NewCluster(size int) *Cluster {
	peers := make([]*Node, size)
	for i := range size {
		peers[i] = NewNode(i, nil)
	}

	for _, peer := range peers {
		peer.peers = peers
	}

	return &Cluster{
		nodes: peers,
	}
}

func (c *Cluster) Start() {
	for _, p := range c.nodes {
		go p.Start()
	}
}

func (c *Cluster) GetLeader() *Node {
	for _, p := range c.nodes {
		p.mu.Lock()

		if p.state == Leader {
			p.mu.Unlock()
			return p
		}

		p.mu.Unlock()
	}

	return nil
}

func (c *Cluster) KillLeader() {
	for _, p := range c.nodes {
		p.mu.Lock()
		if p.state == Leader {
			p.mu.Unlock()
			p.Stop()
			return
		}
		p.mu.Unlock()
	}
}

func (c *Cluster) Stop() {
	for _, p := range c.nodes {
		p.Stop()
	}
}
