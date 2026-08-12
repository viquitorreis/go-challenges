package orderbook

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type OpKind int

const (
	OpAddOrder OpKind = iota
	OpCancel
)

// Operation is what actually get replacted, a command, not a result
// result. Every node applies the same sequence of Operations and
// re-derives the same book state, including re-running Match() itself.
type Operation struct {
	Kind    OpKind
	Order   *Order // set when Kind == OpAddOrder
	OrderID string // set when Kind == OpCancel
}

// Apply executes a single operation agains this order book. This is the
// ONLY path mutation should take once replication exists, direct calls
// to AddOrder/Cancel bypass consensus.
func (ob *OrderBook) Apply(op Operation) {
	switch op.Kind {
	case OpAddOrder:
		ob.addOrder(op.Order)
	case OpCancel:
		ob.cancel(op.OrderID)
	}
}

func EncodeOperation(op Operation) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(op); err != nil {
		return nil, fmt.Errorf("encode operation: %w", err)
	}

	return buf.Bytes(), nil
}

func DecodeOperation(payload []byte) (Operation, error) {
	var op Operation

	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&op); err != nil {
		return Operation{}, fmt.Errorf("decode operation: %w", err)
	}

	return op, nil
}
