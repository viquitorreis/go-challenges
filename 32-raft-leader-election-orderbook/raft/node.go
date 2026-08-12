package raft

type NodeState uint8

const (
	Leader NodeState = iota
	Candidate
	Follower
)
