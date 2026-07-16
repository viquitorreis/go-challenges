# Raft Leader Election

🇺🇸 [English version](README.md)

**Categoria:** Sistemas Distribuídos
**Tempo estimado:** ~2 horas

## O que é

Uma implementação da parte de eleição de líder do algoritmo de consenso Raft, o mesmo algoritmo por trás do etcd, do modo KRaft do Kafka, do Consul, CockroachDB e TiKV: nodes fazem uma eleição com timeouts randomizados, e quem vence manda heartbeats pra manter a liderança.

## O que você aprende

- Os três papéis do Raft (Follower, Candidate, Leader) e as transições de estado entre eles.
- Por que os timeouts de eleição são randomizados por node, pra evitar que todos os nodes comecem uma campanha ao mesmo tempo e dividam o voto indefinidamente.
- Terms como um relógio lógico: cada eleição incrementa o term, e qualquer mensagem carregando um term mais antigo é ignorada, o que evita que um node desatualizado ou recém-recuperado atrapalhe um líder ativo.

## O que foi implementado

- `NewCluster(size int) *Cluster`, `Start()`, `GetLeader() *Node`, `KillLeader()`, `Stop()` pra controlar um cluster simulado.
- `NewNode(id int, peers []*Node) *Node` com `run()` despachando pra `runFollower()`, `runCandidate()` e `runLeader()` conforme o papel atual.
- `sendHeartbeats()` pro líder manter sua autoridade e resetar os timers de eleição dos followers.
- Os testes cobrem um único líder sendo eleito, todos os nodes concordando no term, failover de líder depois de `KillLeader()`, ausência de race conditions, e um node dando step-down ao ver um term maior.

## Decisões de design

- Cada node roda seu próprio loop de máquina de estados (`run()`) reagindo a timers e mensagens recebidas, em vez de um coordenador central controlando o cluster.
- Os timeouts de eleição são randomizados por node especificamente pra quebrar a simetria e tornar votos divididos raros na prática, seguindo a abordagem do paper original do Raft.

## Como rodar

```bash
go run .
go test -race ./...
```
