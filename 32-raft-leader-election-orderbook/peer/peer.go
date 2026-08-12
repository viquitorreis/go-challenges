package peer

import (
	"bytes"
	"context"
	"encoding/gob"
	"log/slog"
	"net"
	"raft_orderbook/errors"
	"raft_orderbook/framing"
	"time"
)

type Peer struct {
	Addr    string
	conn    net.Conn
	framing *framing.Framing
	outbox  chan []byte
	inbound chan<- InboundMsg // cluster chan, peer only writes on it

	lastRcvHeartbeat time.Time

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer receives a connection already stablished
// and the cluster channel channel on where it will develivers everything that arrives on him
func NewPeer(
	parent context.Context,
	addr string,
	conn net.Conn,
	f *framing.Framing,
	inbound chan<- InboundMsg,
) *Peer {
	ctx, cancel := context.WithCancel(parent)
	p := &Peer{
		Addr:    addr,
		conn:    conn,
		framing: f,
		outbox:  make(chan []byte, 16),
		inbound: inbound,
		ctx:     ctx,
		cancel:  cancel,
	}

	go p.readLoop()
	go p.writeLoop()

	return p
}

// Send enqueue a payload already encoded (type+body) to be written
// Never call conn.Write directly from outsite, only writeLoop writes on conn
func (p *Peer) Send(msgType MessageType, body []byte) {
	buf := make([]byte, 0, len(body)+1)
	buf = append(buf, byte(msgType))
	buf = append(buf, body...)

	select {
	case p.outbox <- buf:
	case <-p.ctx.Done():
	}
}

func (p *Peer) MarkAlive() {
	p.lastRcvHeartbeat = time.Now()
}

func (p *Peer) Close() {
	p.cancel()
	p.conn.Close()
}

func (p *Peer) readLoop() {
	for {
		payload, err := p.framing.Read(p.conn)
		if err != nil {
			slog.Error(errors.ErrReadingFromConn.Error(), "error reading from peer", err.Error())
			return
		}

		if len(payload) < 1 {
			continue
		}

		msg := InboundMsg{
			From: p,
			Type: MessageType(payload[0]),
			Body: payload[1:],
		}

		select {
		case p.inbound <- msg:
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Peer) writeLoop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case buf := <-p.outbox:
			if err := p.framing.Write(p.conn, buf); err != nil {
				slog.Error(errors.ErrWritingToConn.Error(), "error", err, "peer", "peer write failed", "addr", p.Addr)
				p.Close()
				return
			}
		case <-ticker.C:
			var body bytes.Buffer
			// sending the time can only be useful on the case of metric latency on the receiver side
			gob.NewEncoder(&body).Encode(time.Now())
			hb := append([]byte{byte(MsgHeartbeat)}, body.Bytes()...)

			if err := p.framing.Write(p.conn, hb); err != nil {
				slog.Error(errors.ErrWritingToConn.Error(), "error", err, "description", "heartbeat write failed", "addr", p.Addr)
				p.Close()
				return
			}
		}
	}
}
