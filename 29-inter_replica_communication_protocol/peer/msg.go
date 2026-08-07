package peer

type MessageType byte

const (
	MsgWriteProposal MessageType = iota
	MsgWriteAck
	MsgHeartbeat
)

type InboundMsg struct {
	From *Peer // peer complete addr
	Type MessageType
	Body []byte
}
