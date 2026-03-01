package main

import (
	"sync"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

type VoteRequest struct {
	Term        int
	CandidateID int
}

type VoteResponse struct {
	Term    int
	Granted bool
}

type Heartbeat struct {
	Term     int
	LeaderID int
}

type Node struct {
	id          int
	state       NodeState
	currentTerm int
	votedFor    int // -1 se não votou nesse term
	votes       int
	mu          sync.Mutex

	// Channels para comunicação entre nós
	heartbeatCh chan Heartbeat
	voteReqCh   chan VoteRequest
	voteRespCh  chan VoteResponse

	// Referência ao cluster inteiro para broadcast
	peers []*Node

	electionTimeout   time.Duration
	heartbeatInterval time.Duration

	once   sync.Once
	stopCh chan struct{}
}
