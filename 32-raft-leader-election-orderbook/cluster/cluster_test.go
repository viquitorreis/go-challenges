package cluster_test

import (
	"context"
	"net"
	"testing"
	"time"

	"raft_orderbook/cluster"
	"raft_orderbook/framing"
	"raft_orderbook/orderbook"
	"raft_orderbook/peer"
	"raft_orderbook/raft"
)

// connectPeers wires two clusters together with an in-memory net.Pipe
// connection, running the handshake on both sides and registering the
// resulting Peer under each cluster the same flow that happens over
// real TCP in production, minus the socket.
func connectPeers(t *testing.T, ctxA context.Context, clA *cluster.Cluster, addrA string,
	ctxB context.Context, clB *cluster.Cluster, addrB string, f *framing.Framing) {
	t.Helper()

	connA, connB := net.Pipe()

	var idA string
	var errB error
	done := make(chan struct{})

	// Launch B's handshake FIRST, in a goroutine, so it's already
	// waiting to exchange bytes by the time A's blocking call runs.
	go func() {
		idA, errB = peer.Handshake(connB, f, addrB)
		close(done)
	}()

	idB, errA := peer.Handshake(connA, f, addrA)
	<-done // wait for B's goroutine to finish before reading idA/errB

	if errA != nil || errB != nil {
		t.Fatalf("handshake failed: %v / %v", errA, errB)
	}

	pA := peer.NewPeer(ctxA, idB, connA, f, clA.InboundChan())
	pB := peer.NewPeer(ctxB, idA, connB, f, clB.InboundChan())

	if !clA.TryRegister(idB, pA) {
		t.Fatalf("failed to register peer %s on cluster A", idB)
	}
	if !clB.TryRegister(idA, pB) {
		t.Fatalf("failed to register peer %s on cluster B", idA)
	}
}

func TestRaftFailover_NewLeaderElectedAfterLeaderDies(t *testing.T) {
	parentCtx := context.Background()
	f := framing.NewFraming(4)

	addrs := []string{"node-a", "node-b", "node-c"}
	ctxs := make([]context.Context, 3)
	cancels := make([]context.CancelFunc, 3)
	clusters := make([]*cluster.Cluster, 3)

	for i, addr := range addrs {
		ctx, cancel := context.WithCancel(parentCtx)
		ctxs[i] = ctx
		cancels[i] = cancel
		ob := orderbook.NewOrderBook("BTC-USD")
		clusters[i] = cluster.NewCluster(ctx, addr, ob, raft.NewRaft(uint64(len(addrs))))
	}
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	// full mesh: connect every pair once
	for i := range addrs {
		for j := i + 1; j < len(addrs); j++ {
			connectPeers(t, ctxs[i], clusters[i], addrs[i], ctxs[j], clusters[j], addrs[j], f)
		}
	}

	// wait for an initial leader to emerge
	leaderIdx := waitForLeader(t, clusters, 3*time.Second)
	initialTerm := clusters[leaderIdx].CurrentTerm()
	t.Logf("initial leader: %s, term: %d", addrs[leaderIdx], initialTerm)

	// kill the leader cancel its context, which tears down its Peer
	// goroutines and connections on both ends, simulating the process
	// dying mid-cluster
	cancels[leaderIdx]()

	survivors := make([]*cluster.Cluster, 0, 2)
	survivorAddrs := make([]string, 0, 2)
	for i := range clusters {
		if i != leaderIdx {
			survivors = append(survivors, clusters[i])
			survivorAddrs = append(survivorAddrs, addrs[i])
		}
	}

	newLeaderIdx := waitForLeader(t, survivors, 3*time.Second)
	newTerm := survivors[newLeaderIdx].CurrentTerm()
	t.Logf("new leader: %s, term: %d", survivorAddrs[newLeaderIdx], newTerm)

	if newTerm <= initialTerm {
		t.Fatalf("expected new leader's term (%d) to be greater than the old leader's term (%d)", newTerm, initialTerm)
	}
}

// waitForLeader polls the given clusters until exactly one reports
// itself as leader, or the timeout expires.
func waitForLeader(t *testing.T, clusters []*cluster.Cluster, timeout time.Duration) int {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for i, cl := range clusters {
			if cl.IsLeader() {
				return i
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("no leader elected within timeout")
	return -1
}
