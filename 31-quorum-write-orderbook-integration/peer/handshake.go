package peer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"multi_node_p2p/framing"
	"net"
)

// Handshake exchanges Hello on a freshly opened connection, before any
// Peer goroutines start. It's synchronous and blocking: send our own
// identity, then block on reading theirs. This is safe over a single
// TCP connection because its full-duplex, writing doesnt block the read,
// so theres no risk of both sides deadlocking waiting on each others
// write to finish first
func Handshake(conn net.Conn, f *framing.Framing, localAddr string) (string, error) {
	var body bytes.Buffer
	if err := gob.NewEncoder(&body).Encode(
		HelloBody{ListenAddr: localAddr},
	); err != nil {
		return "", fmt.Errorf("ecode hello: %w", err)
	}

	// sends its address to other peer
	payload := append([]byte{byte(MsgHello)}, body.Bytes()...)

	// Write in a separate goroutine so it never blocks this call from
	// reaching Read. This matters for transports without OS-level
	// buffering (like net.Pipe in tests) without it, both sides can
	// deadlock waiting on each other's Write to be consumed before
	// either calls Read. Real TCP sockets usually mask this with
	// kernel buffering, but relying on that would be fragile.
	writeErrCh := make(chan error, 1)
	go func() {
		writeErrCh <- f.Write(conn, payload)
	}()

	// waits for peer response
	resp, err := f.Read(conn)
	if err != nil {
		return "", fmt.Errorf("read hello: %w", err)
	}

	if writeErr := <-writeErrCh; writeErr != nil {
		return "", fmt.Errorf("write hello: %w", writeErr)
	}

	if len(resp) < 1 || MessageType(resp[0]) != MsgHello {
		return "", fmt.Errorf("protocol violation: expected hello, got type %d", resp[0])
	}

	var hb HelloBody
	if err := gob.NewDecoder(bytes.NewReader(resp[1:])).Decode(&hb); err != nil {
		return "", fmt.Errorf("decode hello: %w", err)
	}

	return hb.ListenAddr, nil
}
