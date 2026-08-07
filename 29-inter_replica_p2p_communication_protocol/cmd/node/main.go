package main

import (
	"context"
	"inter_comm/cluster"
	"inter_comm/framing"
	"inter_comm/peer"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
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
	// LISTEN_ADDR=:9001 go run ./cmd/node localhost:9002 localhost:9003

	// # terminal 2
	// LISTEN_ADDR=:9002 go run ./cmd/node localhost:9001 localhost:9003

	// # terminal 3
	// LISTEN_ADDR=:9003 go run ./cmd/node localhost:9001 localhost:9002
	cfg := Config{
		ListenAddr: os.Getenv("LISTEN_ADDR"), // ex: ":9001"
		PeerAddrs:  os.Args[1:],              // ex: :9002 :9003
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f := framing.NewFraming(4)
	cl := cluster.NewCluster(ctx)

	go listenLoop(ctx, cfg.ListenAddr, f, cl)

	for _, addr := range cfg.PeerAddrs {
		if cfg.ListenAddr < addr {
			// if its smaller we call it, just a simple solution for unrepeated connections
			// (better solution on next challenges)
			go dialWithRetry(ctx, addr, f, cl)
		}
	}

	slog.Info("node is up and running", "port", cfg.ListenAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	slog.Info("shutting down node")
	cancel()
}

// listenLoop accepts inbound connections from others on the cluster
func listenLoop(ctx context.Context, addr string, f *framing.Framing, cl *cluster.Cluster) {
	ln, err := net.Listen("tcp", addr)
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

		registerPeer(ctx, conn.RemoteAddr().String(), conn, f, cl)
	}
}

func registerPeer(
	ctx context.Context,
	addr string,
	conn net.Conn,
	f *framing.Framing,
	cl *cluster.Cluster,
) {
	p := peer.NewPeer(ctx, addr, conn, f, cl.InboundChan())
	cl.Register(p)
}

// dialWithRetry disks to a peer configured, retrying until it succeeds
// or the context closes. Without it, an up order of the nodes would matter
func dialWithRetry(ctx context.Context, addr string, f *framing.Framing, cl *cluster.Cluster) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			slog.Warn("dial failed, retrying", "addr", addr, "error", err)
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

		registerPeer(ctx, addr, conn, f, cl)

		return
	}
}
