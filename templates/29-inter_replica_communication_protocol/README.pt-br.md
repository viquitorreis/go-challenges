# INTER-REPLICA COMMUNICATION PROTOCOL

**Categoria**: Distributed
**Tempo**: 2h
**Builda em cima de**: distributed_matching_engine (challenge 25, order book single-node) + framing binário do stream broker (22-tcp-multiplexed-stream-broker)

## Estudo antes (10-15min):

### Replicação simples vs consenso

Por que o order book não precisa de Raft completo pra ter consistência aceitável?

A ideia central: quorum de escrita (W) sobre um conjunto de N réplicas garante que
qualquer leitura subsequente com quorum de leitura (R) enxerga o dado mais
recente, desde que R+W>N.

Pense em qual granularidade replicar: o **trade log inteiro** (cada execução, ordenada) ou **snapshot periódico do book**. Trade log é
mais barato de propagar e dá replay determinístico; snapshot é mais simples
de aplicar mas mais pesado. Isso precisa Decidido antes de codar.

## Contexto

O order book até agora roda num processo só. Pra virar distribuído
de verdade, os nós precisam trocar mensagens entre si quem primary, quem
replica, quando uma escrita é considerada "commitada". Hoje você constrói só
o canal e o protocolo de mensagens entre réplicas, não a lógica de decisão
de quorum ainda (isso é o próximo challenge).

O que construir:

1. Reaproveitar o framing length-prefix (4 bytes tamanho + payload) já usado
   no stream broker como base do protocolo inter-réplica não reinventar
   framing novo
2. Três tipos de mensagem: WriteProposal (uma operação de trade log ex:
   novo pedido, cancelamento serializada), WriteAck (id da proposta + nó
   que confirmou), Heartbeat (liveness, sem payload além do id do nó e
   timestamp)
3. Cada nó mantém conexão TCP persistente com os outros nós do cluster
   (topologia fixa via config, não descoberta dinâmica ainda)
4. `SetNoDelay(true)` em toda conexão nova, antes de qualquer I/O isso é
   requisito, não bonus, já que o protocolo faz seu próprio framing
5. Um nó consegue mandar WriteProposal pros outros e receber WriteAck de
   volta, e todos trocam Heartbeat em intervalo fixo (ex: a cada 200ms)

## Requisitos obrigatórios:

- Framing reaproveitado do stream broker, não uma versão nova
- Serialização das 3 mensagens (pode ser struct + encoding/gob, ou binário
  manual sua escolha, mas documente o porquê)
- Uma goroutine dona da conexão com cada peer (nunca múltiplas goroutines
  escrevendo na mesma conn sem coordenação)
- Teste de integração com `net.Pipe()` ou TCP real em 127.0.0.1:0 simulando
  2-3 nós trocando as 3 mensagens
- Roda com -race limpo

**Bonus (se sobrar tempo)**:

- timeout de heartbeat que marca um peer como suspeito (não derruba ainda, só loga), isso vira a base do failover de
para outro challenge

## O que será observado

- Se a decisão de granularidade de replicação (trade log vs snapshot) foi pensada antes de codar ou decidida no meio
- Se o framing foi reaproveitado de verdade ou reinventado
- Se a goroutine-por-peer evita mistura de mutex e channel protegendo a mesma coisa

Decisões de design: documenta no README (ou comentário no topo do arquivo
principal) a escolha entre trade log e snapshot, e o porquê.

---

Primeiro passo: antes de código, qual granularidade você vai replicar,
trade log ou snapshot do book e por quê? Pensa nisso primeiro, depois
vai pro esqueleto do protocolo.