package main

import (
	"math/rand"
	"time"
)

func NewNode(id int, peers []*Node) *Node {
	return &Node{
		id:                id,
		state:             Follower,
		votedFor:          -1,
		heartbeatCh:       make(chan Heartbeat),
		voteReqCh:         make(chan VoteRequest),
		voteRespCh:        make(chan VoteResponse),
		peers:             peers,
		electionTimeout:   time.Duration(rand.Intn(300-150+1)+150) * time.Millisecond,
		heartbeatInterval: time.Duration(time.Millisecond * 100),
		stopCh:            make(chan struct{}),
	}
}

func (n *Node) Start() {
	go n.run()
}

func (n *Node) run() {
	for {
		n.mu.Lock()
		state := n.state
		n.mu.Unlock()

		switch state {
		case Follower:
			n.runFollower()
		case Candidate:
			n.runCandidate()
		case Leader:
			n.runLeader()
		}

		select {
		case <-n.stopCh:
			return
		default:
		}
	}
}

func (n *Node) runFollower() {
	timer := time.NewTimer(n.electionTimeout)
	defer timer.Stop()

	for {
		select {
		case <-n.stopCh:
			return

		case <-timer.C:
			// timeout expirou sem receber o heartbeat do lider, vira candidato
			n.mu.Lock()
			n.state = Candidate
			n.mu.Unlock()
			return // vai parar de rodar como follower e chamar runCandidate

		case hb := <-n.heartbeatCh:
			n.mu.Lock()
			if hb.Term >= n.currentTerm {
				n.currentTerm = hb.Term
				n.state = Follower
			}
			n.mu.Unlock()

			// reset timer
			if !timer.Stop() {
				select {
				case <-timer.C: // dreana o channel se ja expirou
				default:
				}
			}
			timer.Reset(n.electionTimeout)

		case req := <-n.voteReqCh:
			n.mu.Lock()
			granted := false

			// ainda não votou
			if req.Term > n.currentTerm {
				n.currentTerm = req.Term
				n.votedFor = -1
			}

			if req.Term >= n.currentTerm && n.votedFor == -1 {
				n.votedFor = req.CandidateID
				granted = true
				timer.Reset(n.electionTimeout)
			}
			n.mu.Unlock()

			// response ao candidato, caso candidato esteja lendo, se não default
			for _, p := range n.peers {
				if p.id == req.CandidateID {
					select {
					case p.voteRespCh <- VoteResponse{
						Term:    n.currentTerm,
						Granted: granted,
					}:
					default:
					}
					break
				}
			}
		}
	}
}

func (n *Node) runCandidate() {
	n.mu.Lock()
	n.votedFor = n.id
	n.currentTerm++
	n.votes = 1
	term := n.currentTerm
	n.mu.Unlock()

	for _, peer := range n.peers {
		if peer == n {
			continue
		}

		// se um peer estiver ocupado ou morto, não podemos travar o envio
		go func(p *Node) {
			select {
			case p.voteReqCh <- VoteRequest{
				Term:        term,
				CandidateID: n.id,
			}:
			case <-n.stopCh:
			}
		}(peer)
	}

	timer := time.NewTimer(n.electionTimeout)
	defer timer.Stop()

	for {
		select {
		case <-n.stopCh:
			return

		case <-timer.C:
			// timer termina sem quorum, reseta o timer, incrementa o term e reseta votes para 1
			n.mu.Lock()
			timer.Reset(n.electionTimeout)
			n.currentTerm++
			n.mu.Unlock()

			// run vai chamar o runCandidate() novamente
			return
		case hb := <-n.heartbeatCh:
			n.mu.Lock()
			if hb.Term >= n.currentTerm {
				n.currentTerm = hb.Term
			}

			n.state = Follower
			n.mu.Unlock()

			return

		case vote := <-n.voteRespCh:
			n.mu.Lock()
			if vote.Granted && n.currentTerm == vote.Term {
				n.votes++ // recebeu voto tem que somar
				if n.votes >= (len(n.peers)/2)+1 {
					n.state = Leader
					n.mu.Unlock()
					return // se virou lider retorna, vai chamar runLeader()
				}
			}
			n.mu.Unlock()
		}
	}
}

func (n *Node) runLeader() {
	// TODO: envia heartbeats periódicos para todos os peers
	// TODO: se recebe mensagem com term maior -> volta para Follower (step down)
	ticker := time.NewTicker(n.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			n.sendHeartbeats()
		case hb := <-n.heartbeatCh:
			n.mu.Lock()
			if hb.Term >= n.currentTerm {
				n.state = Follower
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()

		case msg := <-n.voteReqCh:
			n.mu.Lock()
			if msg.Term >= n.currentTerm {
				n.state = Follower
				n.mu.Unlock()
				return // vai rodar runFollower
			}
			n.mu.Unlock()
		}
	}
}

func (n *Node) sendHeartbeats() {
	n.mu.Lock()
	term := n.currentTerm
	id := n.id
	n.mu.Unlock()

	for _, peer := range n.peers {
		if peer == n {
			continue
		}

		select {
		case peer.heartbeatCh <- Heartbeat{
			Term:     term,
			LeaderID: id,
		}:
		default:
		}
	}
}

func (n *Node) Stop() {
	n.mu.Lock()
	n.once.Do(func() {
		close(n.stopCh)
	})
	n.mu.Unlock()
}
