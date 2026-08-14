package peer_test

import (
	"context"
	"net"
	"raft_orderbook/cluster"
	"raft_orderbook/framing"
	"raft_orderbook/orderbook"
	"raft_orderbook/peer"
	"raft_orderbook/raft"
	"slices"
	"testing"
	"time"
)

func TestHandshake_ExchangesIdentity(t *testing.T) {
	// net.Pipe gives us two ends of an in-memory connection no real
	// socket, no port, but behaves like a real net.Conn for read/write.
	connA, connB := net.Pipe()
	f := framing.NewFraming(4)

	// Handshake is synchronous and blocking on both ends (send then
	// read), so each side must run in its own goroutine, or they'd
	// deadlock waiting on each other.
	resultCh := make(chan string, 2)
	errCh := make(chan error, 2)

	go func() {
		identity, err := peer.Handshake(connA, f, "localhost:9001")
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- identity
	}()

	go func() {
		identity, err := peer.Handshake(connB, f, "localhost:9002")
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- identity
	}()

	var got []string
	for range 2 {
		select {
		case identity := <-resultCh:
			got = append(got, identity)
		case err := <-errCh:
			t.Fatalf("handshake failed: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("handshake timed out")
		}
	}

	// Side A should have learned B's identity, and vice-versa so the
	// two results should be "localhost:9002" and "localhost:9001", in
	// some order (goroutine scheduling isn't deterministic).
	if !(slices.Contains(got, "localhost:9001") && slices.Contains(got, "localhost:9002")) {
		t.Fatalf("expected both identities exchanged, got: %v", got)
	}
}

func TestCluster_RejectsSecondConnectionSameIdentity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl := cluster.NewCluster(
		ctx,
		"localhost:9002",
		orderbook.NewOrderBook("BTC-USD"),
		raft.NewRaft(uint64(2)),
	)
	f := framing.NewFraming(4)

	// First connection for identity "localhost:9002" should register.
	connA1, connB1 := net.Pipe()
	go peer.Handshake(connB1, f, "localhost:9002")
	identity, err := peer.Handshake(connA1, f, "localhost:9001")
	if err != nil {
		t.Fatalf("first handshake failed: %v", err)
	}
	p1 := peer.NewPeer(ctx, identity, connA1, f, cl.InboundChan())
	defer p1.Close()
	if !cl.TryRegister(identity, p1) {
		t.Fatal("expected first registration to succeed")
	}

	// Second connection, SAME identity should be rejected.
	connA2, connB2 := net.Pipe()
	go peer.Handshake(connB2, f, "localhost:9002")
	identity2, err := peer.Handshake(connA2, f, "localhost:9001")
	if err != nil {
		t.Fatalf("second handshake failed: %v", err)
	}
	p2 := peer.NewPeer(ctx, identity2, connA2, f, cl.InboundChan())
	if cl.TryRegister(identity2, p2) {
		t.Fatal("expected second registration to be rejected as duplicate")
	}
	p2.Close()
}
