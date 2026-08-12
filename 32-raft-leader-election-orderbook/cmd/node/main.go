package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"raft_orderbook/cluster"
	"raft_orderbook/framing"
	"raft_orderbook/orderbook"
	"raft_orderbook/peer"
	"raft_orderbook/raft"
	"syscall"
	"time"
)

// fixed config topology: own address + peers addresses
type Config struct {
	ListenAddr string
	PeerAddrs  []string
}

func main() {
	// run as i.e.:
	// LISTEN_ADDR=localhost:9001 go run ./cmd/node localhost:9002 localhost:9003

	// # terminal 2
	// LISTEN_ADDR=localhost:9002 go run ./cmd/node localhost:9001 localhost:9003

	// # terminal 3
	// LISTEN_ADDR=localhost:9003 go run ./cmd/node localhost:9001 localhost:9002

	// # then on one terminal: "order bid 100 10" and "order ask 100 10" on another one
	// to cancel just "cancel <order_uuid>""
	cfg := Config{
		ListenAddr: os.Getenv("LISTEN_ADDR"), // ex: ":9001"
		PeerAddrs:  os.Args[1:],              // ex: :9002 :9003
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f := framing.NewFraming(4)
	ob := orderbook.NewOrderBook("BTC-USD")
	totalNodes := len(cfg.PeerAddrs) + 1
	cl := cluster.NewCluster(ctx, cfg.ListenAddr, ob, raft.NewRaft(uint64(totalNodes)))

	go listenLoop(ctx, cfg.ListenAddr, f, cl)

	for _, addr := range cfg.PeerAddrs {
		if cfg.ListenAddr < addr {
			go dialWithRetry(ctx, cfg.ListenAddr, addr, f, cl)
		}
	}

	go startCLI(cl)

	slog.Info("node is up and running", "port", cfg.ListenAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	slog.Info("shutting down node")
	cancel()
}

// listenLoop accepts inbound connections from others on the cluster
func listenLoop(ctx context.Context, ownAddr string, f *framing.Framing, cl *cluster.Cluster) {
	ln, err := net.Listen("tcp", ownAddr)
	if err != nil {
		log.Fatalf("err starting listener: %v", err)
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				slog.Error("accept failed", "error", err)
				continue
			}
		}

		// remove nagles algorithm
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.SetNoDelay(true)
		}

		registerPeer(ctx, ownAddr, conn, f, cl)
	}
}

// dialWithRetry disks to a peer configured, retrying until it succeeds
// or the context closes. Without it, an up order of the nodes would matter
func dialWithRetry(ctx context.Context, ownAddr, peerAddr string, f *framing.Framing, cl *cluster.Cluster) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := net.Dial("tcp", peerAddr)
		if err != nil {
			slog.Warn("dial failed, retrying", "addr", ownAddr, "error", err)
			select {
			case <-time.After(time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		if tc, ok := conn.(*net.TCPConn); ok {
			tc.SetNoDelay(true)
		}

		registerPeer(ctx, ownAddr, conn, f, cl)

		return
	}
}

func registerPeer(
	ctx context.Context,
	localAddr string,
	conn net.Conn,
	f *framing.Framing,
	cl *cluster.Cluster,
) {
	// Handshake happens BEFORE any Peer goroutines start, conn is used
	// synchronously here, then handed off to peer.NewPeer only after
	// identity is confirmed.
	identity, err := peer.Handshake(conn, f, localAddr)
	if err != nil {
		slog.Error("handshake failed", "error", err)
		conn.Close()
		return
	}

	p := peer.NewPeer(ctx, identity, conn, f, cl.InboundChan())

	if !cl.TryRegister(identity, p) {
		slog.Warn("duplicate connection to already-registered peer, closing", "peer", identity)
		p.Close()
		return
	}

	slog.Info("peer registered after handshake", "peer", identity)
}
