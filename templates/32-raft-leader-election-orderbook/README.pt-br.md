# CHALLENGE 32: RAFT LEADER ELECTION

**Categoria**: Distributed
**Tempo**: 2h
**Builda em cima de**: 31-quorum-write-orderbook-integration

## Estudo antes

Todo nó está em um de 3 estados: Follower, Candidate, Leader. Todo
nó tem um "term" (número que só cresce, nunca reseta). Follower vira
Candidate se não ouvir heartbeat do líder dentro de um timeout aleatório
(randomizado especificamente pra evitar dois nós virarem candidate ao
mesmo tempo e empatarem eleição repetidamente). Candidate incrementa o
próprio term, vota em si mesmo, manda RequestVote pros outros e quem
recebe só vota se ainda não votou nesse term E o term do candidato é
igual ou maior que o seu. Ganhar maioria dos votos = vira Leader, começa
a mandar heartbeat (o seu Heartbeat de hoje já existe, mas ainda não
carrega term nem serve pra resetar timeout de eleição de ninguém).

## Contexto

Hoje qualquer nó pode propor a qualquer momento, e é isso que
causa a divergência documentada no challenge 31 (duas propostas
concorrentes de nós diferentes, sem ordem total garantida). Eleição de
líder é o primeiro passo pra resolver isso: só o líder eleito vai propor,
depois que isso estiver pronto (essa restrição em si é o próximo
challenge hoje só a eleição, ninguém ainda é impedido de propor).

O que construir:

1. Estado do nó: NodeState (Follower/Candidate/Leader), CurrentTerm
   (uint64, persiste só em memória por enquanto), VotedFor (identidade
   de quem esse nó votou nesse term, ou vazio)
2. Election timeout: timer com duração aleatória (ex: 150-300ms) que
   reseta toda vez que um heartbeat válido chega. Se estourar, nó vira
   Candidate
3. Dois novos tipos de mensagem: MsgRequestVote (term do candidato,
   identidade do candidato) e MsgVoteGranted (term, granted bool)
4. Lógica de Candidate: incrementa term, vota em si mesmo, manda
   RequestVote broadcast, conta votos recebidos, vira Leader se bater
   maioria, volta a Follower se receber heartbeat de term igual ou maior
   de outro nó (alguém já ganhou)
5. Heartbeat existente ganha campo de term Follower que recebe
   heartbeat com term >= o seu reseta o timeout de eleição e reconhece
   o remetente como líder atual

Requisitos obrigatórios:

- VotedFor é resetado a cada novo term (nó pode votar de novo quando o
  term muda)
- Um nó nunca vota duas vezes no mesmo term
- Split vote (ninguém bate maioria, ex: 3 nós, 3 candidatos ao mesmo
  tempo) resolve sozinho, porque o timeout aleatório do próximo round
  não vai ser igual pra todo mundo de novo documenta isso no teste,
  não precisa forçar cenário artificial, só confirmar que reeleição
  acontece se o primeiro round empatar
- Teste de integração: mata o processo do líder atual (ou simula
  timeout), confirma que um novo líder é eleito em algum dos 2
  sobreviventes, com term maior que o anterior

O que será observado: se term é tratado como fonte de verdade pra decidir
quem manda (term maior sempre vence, não importa quem "chegou primeiro"
na rede); se o timeout randomizado realmente varia por nó (timeout fixo
igual em todo mundo reintroduziria empate infinito)

---
Primeiro passo, pra pensar antes de codar: hoje seu MessageType já tem
Hello/Proposal/Ack/Heartbeat/Commit. RequestVote e VoteGranted entram
nesse mesmo enum, certo (mesmo padrão que já vale pra Hello)? E o campo
Term você adicionaria ele em CADA tipo de mensagem existente (Proposal,
Ack, Heartbeat, Commit), ou só nas duas novas (RequestVote/VoteGranted)
por enquanto? Pensa no motivo: um Follower recebendo uma mensagem de
qualquer tipo vinda de um term MAIOR que o seu precisa saber disso pra se
atualizar isso muda sua resposta?