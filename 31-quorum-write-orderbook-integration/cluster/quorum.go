package cluster

import (
	"sync"
	"time"
)

// PendingProposal tracks a write proposal awaiting quorum. It exists
// from the moment a WriteProposal is sent (or received) until either
// enough WriteAcks arrive to commit it, or it times out.
type PendingProposal struct {
	ID        string              // "nodeAddr:counter", unique across the cluster
	Operation []byte              // the serialized order book operation itself
	Voters    map[string]struct{} // which peer identities have acked (set, not count)
	CreatedAt time.Time           // for timeout sweeping
}

// PendingProposals is a share table of proposals, awaiting quorum.
// Separate type (not just a field on Cluster) because it has its
// ownb concurrency rules distinct from the peers map.
type PendingProposals struct {
	items map[string]*PendingProposal
	mu    sync.Mutex
}

func NewPendingProposals() *PendingProposals {
	return &PendingProposals{
		items: make(map[string]*PendingProposal),
	}
}

// Register registers a new proposal. Called both when THIS node originates
// a proposal, and is a no-op path for proposals received from others
// (they dont need pending-tracking on the receiving side)
func (p *PendingProposals) Register(id string, op []byte) *PendingProposal {
	p.mu.Lock()
	defer p.mu.Unlock()

	pp := &PendingProposal{
		ID:        id,
		Operation: op,
		Voters:    make(map[string]struct{}),
		CreatedAt: time.Now(),
	}

	p.items[id] = pp

	return pp
}

// RecordVote adds a peer's ack to a proposal's voter set. Returns the new
// vote count and whether this proposal exists at all, if it doesn't
// (already committed, already timed out, or ack for a proposal this node never originated),
// the caller should just ignore the ack. The map naturally gives idempotency: adding the
// same identity twice doesn't increase len(Voters).
func (p *PendingProposals) RecordVote(id, voterIdentity string) (votes int, exists bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pp, ok := p.items[id]
	if !ok {
		return 0, false
	}

	pp.Voters[voterIdentity] = struct{}{}

	return len(pp.Voters), true
}

// Remove takes a proposal out of tracking, its called both on successful commit
// (quorum reached) and on timeout (gave up waiting)
func (p *PendingProposals) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.items, id)
}

// SweepExpired returns IDs of proposals older than maxAge, for the
// timeout goroutine to log and remove.
func (p *PendingProposals) SweepExpired(maxAge time.Duration) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var expired []string
	for id, pp := range p.items {
		if time.Since(pp.CreatedAt) > maxAge {
			expired = append(expired, id)
		}
	}

	return expired
}

func (p *PendingProposals) VoteCount(id string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	pp, ok := p.items[id]
	if !ok {
		return 0
	}

	return len(pp.Voters)
}

func (p *PendingProposals) Get(id string) *PendingProposal {
	p.mu.Lock()
	defer p.mu.Unlock()

	pp, ok := p.items[id]
	if !ok {
		return nil
	}

	return pp
}
