package peer

type MessageType byte

const (
	MsgWriteProposal MessageType = iota
	MsgWriteAck
	MsgHeartbeat
	MsgHello  // first msg on any connection, announces identity
	MsgCommit // broadcast by the proposer once quorum is reached
)

type InboundMsg struct {
	From *Peer // peer complete addr
	Type MessageType
	Body []byte
}

type HelloBody struct {
	ListenAddr string
}
