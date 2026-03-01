package main

func NewNode(id int, peers []*Node) *Node {
	// TODO: inicializa com state=Follower, currentTerm=0, votedFor=-1
	// TODO: election timeout randomizado entre 150-300ms
	// TODO: cria todos os channels
	panic("implement me")
}

func (n *Node) Start() {
	// TODO: lança goroutine principal: run()
	panic("implement me")
}

func (n *Node) run() {
	// TODO: loop que despacha para runFollower/runCandidate/runLeader
	// baseado no estado atual
	panic("implement me")
}

func (n *Node) runFollower() {
	// TODO: espera heartbeat ou election timeout
	// Se timeout -> transiciona para Candidate
	// Se heartbeat com term >= currentTerm -> reseta timer, atualiza term se necessário
	// Se VoteRequest com term >= currentTerm e ainda não votou -> concede voto
	panic("implement me")
}

func (n *Node) runCandidate() {
	// TODO: incrementa term, vota em si mesmo
	// TODO: envia VoteRequest para todos os peers
	// TODO: espera votos com timeout (outro election timeout)
	// Se ganhou quorum -> transiciona para Leader
	// Se recebe heartbeat com term >= currentTerm -> volta para Follower
	// Se timeout sem quorum -> começa nova eleição (loop)
	panic("implement me")
}

func (n *Node) runLeader() {
	// TODO: envia heartbeats periódicos para todos os peers
	// TODO: se recebe mensagem com term maior -> volta para Follower (step down)
	panic("implement me")
}

func (n *Node) sendHeartbeats() {
	// TODO: itera sobre peers e envia Heartbeat no channel de cada um
	// Use select + default para não bloquear se peer não está lendo
	panic("implement me")
}

func (n *Node) requestVotes() {
	// TODO: envia VoteRequest para cada peer
	// Cada peer deve processar e responder no voteRespCh do candidato
	panic("implement me")
}

func (n *Node) Stop() {
	// TODO: stop stopChan
}
