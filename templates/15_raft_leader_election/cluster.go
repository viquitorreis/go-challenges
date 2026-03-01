package main

// Cluster gerencia o conjunto de nós
type Cluster struct {
	nodes []*Node
}

func NewCluster(size int) *Cluster {
	// TODO: cria nós, conecta peers (cada nó conhece todos os outros)
	panic("implement me")
}

func (c *Cluster) Start() {
	// TODO: chama Start() em cada nó
	panic("implement me")
}

func (c *Cluster) GetLeader() *Node {
	// TODO: retorna o nó com state == Leader, ou nil
	panic("implement me")
}

func (c *Cluster) KillLeader() {
	// TODO: chama Stop() no nó líder atual
	panic("implement me")
}

func (c *Cluster) Stop() {
	// TODO: para todos os nós
	panic("implement me")
}
