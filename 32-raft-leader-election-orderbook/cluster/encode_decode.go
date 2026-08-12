package cluster

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type RequestVoteBody struct {
	Term        uint64
	CandidateID string
}

func encodeRequestVote(term uint64, candidateID string) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(RequestVoteBody{Term: term, CandidateID: candidateID}); err != nil {
		return nil, fmt.Errorf("encode request vote: %w", err)
	}

	return buf.Bytes(), nil
}

func decodeRequestVote(body []byte) (term uint64, candidateID string, err error) {
	var rv RequestVoteBody

	if err := gob.NewDecoder(bytes.NewReader(body)).Decode(&rv); err != nil {
		return 0, "", fmt.Errorf("decode request vote: %w", err)
	}

	return rv.Term, rv.CandidateID, nil
}

type VoteResponseBody struct {
	Term    uint64
	Granted bool
}

func encodeVoteResponse(term uint64, granted bool) []byte {
	var buf bytes.Buffer
	// gob encode of a simple struct never fails on valid input types,
	// but we still check err in decode since that's where malformed
	// data could actually appear (from network, not from our own struct)
	gob.NewEncoder(&buf).Encode(VoteResponseBody{Term: term, Granted: granted})

	return buf.Bytes()
}

func decodeVoteResponse(body []byte) (term uint64, granted bool, err error) {
	var vr VoteResponseBody

	if err := gob.NewDecoder(bytes.NewReader(body)).Decode(&vr); err != nil {
		return 0, false, fmt.Errorf("decode vote response: %w", err)
	}

	return vr.Term, vr.Granted, nil
}

type LeaderHeartbeatBody struct {
	Term     uint64
	LeaderID string
}

func encodeLeaderHeartbeat(term uint64, leaderID string) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(LeaderHeartbeatBody{Term: term, LeaderID: leaderID}); err != nil {
		return nil, fmt.Errorf("encode leader heartbeat: %w", err)
	}

	return buf.Bytes(), nil
}

func decodeLeaderHeartbeat(b []byte) (term uint64, leaderID string, err error) {
	var body LeaderHeartbeatBody

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&body); err != nil {
		return 0, "", fmt.Errorf("decode leader heartbeat: %w", err)
	}

	return body.Term, body.LeaderID, nil
}
