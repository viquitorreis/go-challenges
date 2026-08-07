package peer_test

import (
	"net"
	"slices"
	"testing"
	"time"

	"multi_node_p2p/framing"
	"multi_node_p2p/peer"
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
