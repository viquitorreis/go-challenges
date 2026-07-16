# Worker Pool com Priority Queue

🇺🇸 [English version](README.md)

**Categoria:** Concorrência / Estruturas de Dados
**Tempo estimado:** ~1h30

## O que é

Um sistema de processamento de jobs onde um pool fixo de workers sempre pega o job de maior prioridade disponível primeiro, no mesmo espírito de Sidekiq ou Celery, com múltiplos produtores submetendo jobs e múltiplos workers consumindo concorrentemente.

## O que você aprende

- Envolver o `container/heap` do Go (que não é thread-safe sozinho) com um mutex sem transformar cada operação num gargalo.
- Usar `sync.Cond` pra que workers ociosos bloqueiem de forma eficiente numa fila vazia, em vez de busy-loop ou polling com `time.Sleep`.
- Backpressure: o que fazer quando jobs chegam mais rápido do que os workers conseguem processar, em vez de deixar a fila crescer sem limite.

## O que foi implementado

- `JobHeap` implementando `heap.Interface` (`Len`, `Less`, `Swap`, `Push`, `Pop`), ordenado pra que o menor número de prioridade seja retirado primeiro.
- `PriorityQueue` envolvendo a heap com um mutex: `Enqueue`, `Dequeue`, `Len`, `Shutdown`.
- `WorkerPool` com número configurável de workers e um callback `processor func(*Job) error`: `Start`, `Submit`, `Shutdown`, `Stats`.
- `Enqueue` retorna um erro quando a fila limitada está cheia em vez de crescer sem limite, dando ao chamador um backpressure explícito.
- Os testes cobrem ordenação básica da heap, thread safety sob enqueue/dequeue concorrente, a fila limitada rejeitando quando cheia, processamento de jobs, e ordenação por prioridade sob carga.

## Decisões de design

- A fila impõe um tamanho máximo (`maxSize`) e rejeita novos jobs além desse ponto em vez de permitir crescimento ilimitado de memória - uma escolha explícita de backpressure em vez de descartar silenciosamente ou bloquear pra sempre.
- `sync.Cond` acorda as goroutines de worker exatamente quando um job fica disponível ou o shutdown é disparado, evitando tanto busy-waiting quanto a latência de polling.

## Como rodar

```bash
go run .
go test ./...
```
