package main

import (
	"testing"
	"time"
)

// Testa que exatamente um líder é eleito
func TestSingleLeaderElected(t *testing.T) {
	cluster := NewCluster(3)
	cluster.Start()
	defer cluster.Stop()

	time.Sleep(1 * time.Second)

	leaders := 0
	for _, n := range cluster.nodes {
		n.mu.Lock()
		if n.state == Leader {
			leaders++
		}
		n.mu.Unlock()
	}

	if leaders != 1 {
		t.Errorf("expected 1 leader, got %d", leaders)
	}
}

// Testa que todos os followers reconhecem o mesmo term
func TestAllNodesAgreeOnTerm(t *testing.T) {
	cluster := NewCluster(3)
	cluster.Start()
	defer cluster.Stop()

	time.Sleep(1 * time.Second)

	leader := cluster.GetLeader()
	if leader == nil {
		t.Fatal("no leader elected")
	}

	leader.mu.Lock()
	leaderTerm := leader.currentTerm
	leader.mu.Unlock()

	for _, n := range cluster.nodes {
		n.mu.Lock()
		nodeTerm := n.currentTerm
		n.mu.Unlock()

		if nodeTerm != leaderTerm {
			t.Errorf("node %d has term %d, leader has term %d", n.id, nodeTerm, leaderTerm)
		}
	}
}

// Testa re-eleição após morte do líder
func TestLeaderFailover(t *testing.T) {
	cluster := NewCluster(3)
	cluster.Start()
	defer cluster.Stop()

	time.Sleep(1 * time.Second)

	firstLeader := cluster.GetLeader()
	if firstLeader == nil {
		t.Fatal("no initial leader")
	}
	firstLeaderID := firstLeader.id

	cluster.KillLeader()

	// Espera nova eleição
	time.Sleep(2 * time.Second)

	newLeader := cluster.GetLeader()
	if newLeader == nil {
		t.Fatal("no new leader after failover")
	}

	if newLeader.id == firstLeaderID {
		t.Error("same node became leader after it was killed")
	}

	// Term deve ter aumentado
	newLeader.mu.Lock()
	newTerm := newLeader.currentTerm
	newLeader.mu.Unlock()

	firstLeader.mu.Lock()
	oldTerm := firstLeader.currentTerm
	firstLeader.mu.Unlock()

	if newTerm <= oldTerm {
		t.Errorf("new term %d should be > old term %d", newTerm, oldTerm)
	}
}

// Testa race conditions - run with: go test -race
func TestNoRaceConditions(t *testing.T) {
	cluster := NewCluster(5)
	cluster.Start()
	defer cluster.Stop()

	// Mata e verifica líderes rapidamente para estressar a eleição
	for i := 0; i < 3; i++ {
		time.Sleep(800 * time.Millisecond)
		cluster.KillLeader()
		time.Sleep(800 * time.Millisecond)
		leader := cluster.GetLeader()
		if leader == nil {
			t.Logf("iteration %d: no leader yet (may still be electing)", i)
		}
	}
}

// Testa que um nó com term maior derruba o líder atual
func TestHigherTermStepsDownLeader(t *testing.T) {
	cluster := NewCluster(3)
	cluster.Start()
	defer cluster.Stop()

	time.Sleep(1 * time.Second)

	leader := cluster.GetLeader()
	if leader == nil {
		t.Fatal("no leader")
	}

	// Injeta heartbeat com term muito maior em todos os nós
	for _, n := range cluster.nodes {
		select {
		case n.heartbeatCh <- Heartbeat{Term: 999, LeaderID: -1}:
		default:
		}
	}

	time.Sleep(300 * time.Millisecond)

	leader.mu.Lock()
	state := leader.state
	leader.mu.Unlock()

	if state == Leader {
		t.Error("old leader should have stepped down after seeing higher term")
	}
}
