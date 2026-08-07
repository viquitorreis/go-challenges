# CHALLENGE 30: CONNECTION IDENTITY + INTEGRATION TEST + MULTI-NODE SPAWN

**Categoria**: Distributed
**Tempo**: 2h
**Builda em cima de**: 29-inter_replica_communication_protocol

## Contexto

O challenge 29 fechou o canal de mensagens (WriteProposal/
WriteAck/Heartbeat) mas deixou 3 débitos:

- Identidade de conexão baseada em RemoteAddr() efêmero (quebra com conexão aceita)
- Sem teste de integração formal
- Sem validação de N nós reais rodando e se comunicando.

Esse challenge fecha os 3, sem entrar em lógica de quorum (isso é o challenge 31).

**O que construir**:

1. Handshake de identidade: primeira mensagem trocada em qualquer conexão
   nova (antes de Proposal/Ack/Heartbeat) anuncia "sou o nó X, meu endereço
   de listen é Y". Substitui a regra provisória de dial lexicográfico
   com identidade real, conexão duplicada entre o mesmo par de nós pode
   ser detectada e fechada explicitamente, em vez de evitada por convenção
2. Teste de integração: net.Pipe() ou TCP real em 127.0.0.1:0 simulando
   2-3 nós completando handshake e trocando as 3 mensagens
3. -race limpo em toda a suíte
4. Spawn de 3+ processos reais (não simulados em teste) rodando o
   binário, confirmando handshake + heartbeat trafegando entre todos

## Requisitos obrigatórios:

- Handshake vira o primeiro caso tratado no dispatch de mensagem do peer,
  antes de qualquer outra lógica
- Conexão duplicada entre o mesmo par (mesma identidade dos dois lados)
  detectada e uma das duas fechada, com log explicando qual sobreviveu
- Teste de integração cobre: handshake completa, peer registrado no
  cluster com identidade correta (não RemoteAddr), heartbeat MarkAlive
  disparado
- 3+ nós reais sobem, se conectam, handshake completa em todos os pares,
  sem conexão duplicada sobrando

O que será observado: se identidade resolve os dois problemas (RemoteAddr
efêmero E conexão duplicada) ou só um dos dois; se o teste de integração
testa o handshake de verdade, não só as mensagens que já existiam

---

Primeiro passo: MsgHello entra no mesmo enum MessageType que já existe
(WriteProposal/WriteAck/Heartbeat), ou o handshake é um protocolo
separado, fora desse envelope tipo+body? Pensa em como o Peer.readLoop
despacha hoje (lê tipo, switch) antes de decidir.