# ORDER BOOK REWRITE SKIP LIST + DOUBLY LINKED CANCELLATION

**Categoria:** Data Structures & Systems Design
**Tempo:** 3h+ (pode ser dividido em 2 sessões ver "Sequência sugerida" abaixo)
**Builda em cima de:** o challenge original do order book (heap-based, challenge 11). Se você ainda
não fez aquele, comece por lá este challenge assume que você já tem um matching
engine funcional e está reescrevendo a estrutura de dados interna, não o
comportamento.

## Como resolver este challenge

1. Pega a implementação original do order book (challenge 1, heap + `map[preço][]*Order`)
2. Copia ela pra uma pasta nova mantém o original intacto como referência de
   comportamento esperado (você vai usá-lo pra validar que a reescrita não mudou o
   resultado observável, mas sim performance e latência)
3. Reescreve a estrutura interna seguindo os requisitos abaixo
4. Ao final, os dois devem produzir o **mesmo resultado** pros mesmos inputs a
   reescrita muda a estrutura de dados, não a lógica de negócio

## Visão geral do problema

Um order book implementado com `container/heap` (ou qualquer heap binário padrão) mais
um `map[preço][]*Order` tem duas fraquezas estruturais que só aparecem sob volume real:

1. **Heap não suporta remoção arbitrária.** Ele só sabe tirar o topo. Cancelar o
   último pedido de um nível de preço obriga reconstruir o heap inteiro (O(n log n))
   ou aceitar entradas "fantasma" que precisam ser filtradas depois.
2. **Slice não suporta remoção O(1) do meio.** Cancelar uma ordem específica dentro
   de um nível de preço, mesmo que o nível não esvazie, exige varrer o slice inteiro
   pra reconstruí-lo sem aquele elemento O(tamanho do nível), toda vez. Isso é um grande problema
    em orderbooks (matching engines) em sistemas em produção, onde em média 70% - 80% são ordens de cancelamento.

Dado que cancelamento é tipicamente a operação mais frequente num order book real
(mais comum que fills), esses dois pontos tendem a dominar o custo do sistema sob
carga, mesmo que o matching em si seja rápido.

## Requisitos obrigatórios

### Frente A estrutura dos price levels
- Substitui heap + map por uma **skip list** ordenada por preço (bids em ordem
  decrescente melhor preço primeiro; asks em ordem crescente)
- Inserção, remoção e busca do melhor preço devem ser O(log n) esperado
- A skip list deve suportar **iteração ordenada a partir do topo** (necessário pra
  qualquer requisito futuro de profundidade de book ver Bonus)

### Frente B estrutura interna de cada price level
- Cada `PriceLevel` deixa de guardar `Orders []*Order` e passa a guardar uma
  **lista duplamente encadeada** (`container/list` do Go, ou implementação própria)
- O `tracker` (lookup de ordem por ID) deixa de guardar só `*Order` passa a
  guardar também o nó da lista (`*list.Element` ou equivalente), permitindo
  remover uma ordem específica em **O(1)**, independente de onde ela esteja
  dentro do nível

### Preservação de comportamento
- `AddOrder`, `Match`, `Cancel`, `BidDepth`, `AskDepth` continuam com a mesma
  assinatura pública e o mesmo comportamento observável do order book original
- Price-time priority (FIFO dentro de cada nível de preço) precisa continuar
  correta a lista encadeada preserva ordem de inserção naturalmente, só
  confirma que sua implementação não inverte isso em nenhum ponto
- Thread-safety mantida (mesmo nível de proteção contra concorrência que o
  original já tinha)

## Decisões de design registre suas escolhas e o porquê

Diferente da maioria dos challenges anteriores, este não tem uma única resposta
"certa" pra algumas decisões estruturais. Documenta no seu README de solução:

**1. Skip list própria vs adaptação de uma já existente.** Se você já tem uma skip
list genérica de outro challenge, ela suporta comparador customizável (ordem
crescente pra asks, decrescente pra bids)? Vale adaptar ou reescrever uma versão
mais enxuta específica pra esse caso?

**2. Locking em camadas.** Se sua skip list tem lock próprio (thread-safe por si
só) e o `OrderBook` também tem seu mutex protegendo tudo, você tem dois locks
empilhados por operação redundante e potencial fonte de contenção
desnecessária. Decisão a registrar: skip list com lock interno (mais segura
isoladamente, mais lenta em conjunto) ou skip list sem lock, confiando inteiramente
no mutex do `OrderBook` (mais rápida, mas só correta se **todo** acesso realmente
passar pelo `OrderBook`)?

**3. Quantidade agregada por nível.** Vale a pena o `PriceLevel` manter um
`TotalQty` incremental (somado/subtraído a cada insert/cancel/match), transformando
`BidDepth`/`AskDepth` de "somar todas as ordens individuais" pra "somar os poucos
price levels"? Isso não muda a assinatura pública, só a complexidade interna do
cálculo decisão de custo/benefício de implementação vs performance.

## Comparação de referência: Skip List vs Heap Indexado

Caso você decida ir por heap indexado em vez de skip list (variação válida do
challenge ajuste os requisitos de acordo se escolher esse caminho):

| Critério | Skip List | Heap Indexado |
|---|---|---|
| Melhor preço (peek) | O(1) | O(1) |
| Insert/remove nível (com referência em mãos) | O(log n) esperado | O(log n) garantido |
| Top-N níveis ordenados (profundidade de book) | O(k), natural | Ruim exige pop destrutivo ou sort O(n log n) |
| Garantia de pior caso | Probabilística | Determinística |
| Overhead de memória | Maior (múltiplos ponteiros por nó) | Menor |
| Risco de bug de implementação | Baixo-médio | Médio-alto (índice precisa ficar sincronizado a cada swap) |

O fator que mais pesa pra maioria dos casos: se o seu order book vai eventualmente
precisar publicar profundidade de mercado (top-N níveis, não só o melhor preço)
requisito comum em qualquer sistema real de trading skip list se paga por
suportar isso nativamente. Heap indexado é competitivo se a única operação que
importa é sempre o melhor preço isolado.

## Bonus (se sobrar tempo)

- Método `TopLevels(n int) []PriceLevel` usa a iteração ordenada da skip list pra
  devolver os N melhores níveis de cada lado, de graça graças à Frente A
- Benchmark comparando a implementação original (heap) vs a nova (skip list),
  especificamente no cenário de **muitos cancelamentos intercalados com poucas
  ordens novas** esse é o cenário que mais expõe a diferença
- `go test -race` nos dois lados pra confirmar que a reescrita não introduziu
  nenhuma condição de corrida nova

## O que será observado

1. Se as duas frentes (A e B) foram resolvidas de forma **independente e
   testável** cada uma, ou se ficaram acopladas de um jeito que dificulta saber
   qual delas introduziu um bug, caso apareça um
2. Se o comportamento observável (resultado de matches, profundidade, ordem de
   execução) permanece idêntico ao order book original nos mesmos cenários de
   teste
3. Se as decisões de design (locking, estrutura própria vs reaproveitada,
   agregação de quantidade) foram tomadas conscientemente e documentadas, não só
   "o que compilou primeiro"

## Sequência sugerida (se dividir em 2 sessões, ~1h cada)

**Sessão 1 Frente B isolada.** Troca só a estrutura interna de cada
`PriceLevel` (slice → lista encadeada), mantendo heap+map exatamente como no
original por enquanto. Testável isoladamente: cancelamento no meio de um nível
já fica O(1), sem mexer em mais nada.

**Sessão 2 Frente A por cima da Sessão 1 já estável.** Com o `PriceLevel` já
funcionando, troca heap+map pela skip list. A mudança fica isolada na camada de
"como encontro/insiro/removo um nível de preço", sem misturar com a lógica interna
do nível, que já foi validada na sessão anterior.

Ao final: roda o mesmo cenário de exemplo do order book original (as ordens de
teste do `main.go` de referência, incluindo pelo menos um match e um cancel) nos
dois e confirma que o resultado é idêntico.

## Load Test Stage 1: Implementação Single Node, Concurrency em Contenção

Simula um fluxo de ordens realista (70% cancel, 20% add, 10% match) em
níveis crescentes de goroutines concorrentes, medindo throughput e
latência por operação individual.

Rodar com `go test -run TestLoad -v ./...`.

| Workers | Total de ops | Throughput | p50 | p99 |
|---|---|---|---|---|
| 1 | 500 | 2.301.867 ops/s | 248ns | 2,5µs |
| 10 | 5.000 | 1.207.425 ops/s | 384ns | 134µs |
| 50 | 25.000 | 1.055.594 ops/s | 476ns | 1,1ms |
| 100 | 50.000 | 1.137.860 ops/s | 438ns | 2,2ms |

### O que está acontecendo de verdade: contenção de lock

O `OrderBook` usa um único `sync.RWMutex` protegendo `AddOrder`, `Cancel` e
`Match`. Toda operação, por menor que seja, precisa esperar sua vez de
conseguir esse lock antes de fazer qualquer trabalho de verdade.

Com 1 worker, não tem ninguém pra esperar, então a latência fica baixa e
consistente. Assim que mais goroutines passam a disputar o mesmo lock,
duas coisas acontecem ao mesmo tempo:

- **O throughput cai na hora** (1 → 10 workers: de 2,3M pra 1,2M ops/s),
  porque tempo de CPU que antes ia pra trabalho útil agora vai pra
  goroutines paradas, esperando o lock
- **O p99 cresce muito mais rápido que o p50** (2,5µs → 2,2ms, quase
  900x, enquanto o p50 só sai de 248ns pra 438ns). A maioria das operações
  continua rápida assim que consegue o lock. Mas uma fatia crescente delas
  fica presa esperando atrás de todo mundo na fila, e são essas que
  inflam a cauda de latência

Esse é o formato clássico que a contenção de lock assume: a mediana quase
não se mexe, enquanto a cauda piora drasticamente. Um dashboard mostrando
só a latência média perderia esse padrão completamente.

### Por que o throughput estabiliza em vez de continuar subindo

Depois de 10 workers, o throughput fica relativamente estável (~1,0-1,1M
ops/s), em vez de continuar caindo ou subindo. Isso é o mutex já
totalmente saturado: o sistema já está gastando praticamente todo o tempo
que vai gastar serializando o acesso ao lock. Adicionar mais goroutines
depois desse ponto não ajuda nem atrapalha muito — só deixa a fila atrás
do lock mais comprida, que é exatamente o que o p99 crescente mostra.

### Uma ressalva que vale ser honesto

`p50=248ns` pra uma operação que inclui busca numa skip list e mutação de
lista encadeada parece rápido demais. Explicação provável: com 70% das
operações sendo `Cancel` num ID sorteado aleatoriamente entre só 1.000
ordens pré-populadas, uma fração relevante dessas chamadas bate direto em
`if !exists { return false }`, sem fazer trabalho de verdade, porque outra
goroutine já cancelou aquele ID antes. Isso não invalida o padrão de
contenção (a fila do lock é real de qualquer forma), mas significa que o
número bruto de throughput está um pouco inflado. Um benchmark mais fiel
garantiria que cada cancel mira numa ordem que ainda está realmente ativa.

### Conclusão

A reescrita da estrutura de dados (skip list + lista duplamente encadeada)
fez o trabalho dela: operações individuais são rápidas. O teto atual não é
mais complexidade algorítmica, é o mutex único serializando todo o acesso.
Esse é o argumento concreto pra evolução planejada, seja particionando o
lock (ex: um lock por price level) ou indo pra um design estilo actor
(uma única goroutine dona do estado do book, todo mundo mais falando com
ela via channel, sem lock nenhum) conforme isso evolui pro matching engine
distribuído.