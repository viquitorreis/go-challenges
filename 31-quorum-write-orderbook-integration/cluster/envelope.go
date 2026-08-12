package cluster

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// ProposalEnvelope bundles a proposal's ID with its operation payload,
// so both travel together as a single WriteProposal message body.
type ProposalEnvelope struct {
	ID string
	Op []byte
}

func encodeProposal(id string, op []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(ProposalEnvelope{ID: id, Op: op}); err != nil {
		return nil, fmt.Errorf("encode proposal: %w", err)
	}
	return buf.Bytes(), nil
}

func decodeProposal(body []byte) (id string, op []byte, err error) {
	var env ProposalEnvelope
	if err := gob.NewDecoder(bytes.NewReader(body)).Decode(&env); err != nil {
		return "", nil, fmt.Errorf("decode proposal: %w", err)
	}
	return env.ID, env.Op, nil
}
