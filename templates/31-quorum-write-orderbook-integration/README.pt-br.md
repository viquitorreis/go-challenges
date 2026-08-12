# CHALLENGE 31: QUORUM WRITE + ORDERBOOK INTEGRATION

**Categoria**: Distributed
**Tempo**: 2h
**Builda em cima de**: 30-multi_node_p2p_connection_identity + distributed_matching_engine (orderbook)

## Estudo antes (10-15min)

Revisão rápida de quorum de escrita (W): com N
nós no cluster, W = (N/2)+1 garante maioria dos votos. Uma proposta só é
considerada "commitada" quando o proponente recebeu WriteAck de W-1 outros nós (o
próprio nó já conta como 1 voto implícito). Pensa em: o que acontece se
um nó nunca responder (caiu, rede lenta)? A proposta precisa de timeout
sem isso, uma proposta pendente vive pra sempre esperando um voto que
nunca vem.

## Contexto

Hoje cluster e orderbook são dois mundos que não se tocam.
WriteProposal chega, cluster manda ack vazio, ninguém aplica nada. Esse
challenge conecta os dois: proposta local vira WriteProposal real pro
cluster antes de aplicar; proposta remota só aplica no orderbook local
depois de bater quorum.

O que construir:

1. PendingProposal: struct rastreando uma proposta em andamento ID
   (nodeAddr:counter), a operação em si (serializada), quantos acks já
   recebeu, de quais peers, timestamp de criação (pra timeout)

2. Cluster.Propose(op Operation) entrypoint local: gera ID novo
   (incrementa contador interno), registra PendingProposal, broadcast
   WriteProposal pra todos os peers, conta o próprio voto imediatamente

3. handleInboundMsg, case WriteProposal: ao receber proposta de outro nó,
   responde WriteAck referenciando o ID recebido (não mais vazio como
   hoje) mas NÃO aplica no orderbook ainda, só confirma recebimento

4. handleInboundMsg, case WriteAck: casa o ack com o PendingProposal pelo
   ID, incrementa contador de votos daquele proposal. Se atingiu quorum
   (calculado dinamicamente a partir de len(peers)+1), aplica a operação
   no orderbook local E remove o PendingProposal do tracking

5. Timeout de proposta pendente: goroutine (ou ticker no Bootstrap) que
   periodicamente varre PendingProposals antigas (ex: >5s sem quorum),
   loga como falha, remove do tracking não trava o cluster esperando
   pra sempre

## Requisitos obrigatórios

- ID de proposta é nodeAddr:counter, único por proposta, nunca reusado
- Quorum é calculado (len(peers)+1)/2 + 1, nunca hardcoded pro número 3
- PendingProposals protegido por mutex próprio (acesso concorrente:
  chega ack de peer diferente ao mesmo tempo que timeout varre a lista)
- Ack duplicado do mesmo peer pra mesma proposta não conta dois votos
  (idempotência um peer só vota uma vez por proposta, mesmo que
  reenvie)
- Timeout remove proposta sem crashar nada, só loga

## Bonus (se sobrar tempo)

Quando uma proposta expira por timeout, decide
se vale reenviar automaticamente ou só reportar falha pro caller original
 documenta a escolha, não precisa implementar retry completo hoje

## O que será observado

Se quorum é calculado, não fixo; se ack duplicado
é tratado corretamente; se o timeout existe e funciona (proposta que
nunca bate quorum não vaza memória nem trava nada)

**Decisões de design**

Documenta no README o motivo do esquema de ID
(nodeAddr:counter) e por que quorum é calculado em vez de fixo mesmo
sendo só 3 nós hoje, isso é o padrão que sistemas reais usam pra não
precisar reescrever a lógica quando o cluster cresce.

---

Primeiro passo: antes de qualquer código, pensa no formato do
PendingProposal ele precisa saber "quais peers já votaram", não só
"quantos votos". Por quê isso importa, dado o requisito de idempotência
(ack duplicado não pode contar duas vezes)? Que estrutura de dado Go
resolve "conjunto de peers que já votaram" de forma natural?