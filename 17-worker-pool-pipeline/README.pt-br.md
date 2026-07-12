# Worker Pool + Pipeline (Log Processing)

🇺🇸 [English version](README.md)

**Categoria:** Concorrência / Pipeline
**Tempo estimado:** ~40 minutos

## O que é

Um pipeline de 3 estágios simulando processamento de logs: gera entradas de log mockadas, filtra só os erros usando um worker pool, e empurra o resultado por um sink com rate limit (tipo uma API externa com limite de requests/segundo).

## O que você aprende

- Combinar um **pipeline** (estágios sequenciais) com um **worker pool dentro de um estágio** (fan-out pra filtrar, fan-in de volta pra um único channel).
- A parte mais complicada de um estágio de fan-in: múltiplos workers escrevendo no mesmo channel de saída, garantindo que ele só seja fechado depois que TODOS os workers terminarem, não depois do primeiro.
- Aplicar rate limit num estágio de saída com `time.Ticker`/`time.Sleep`, ainda respeitando cancelamento por `context`.

## O que foi implementado

- `generateLogs(ctx context.Context, n int) <-chan LogEntry` gerando `n` entradas mockadas com níveis aleatórios, respeitando `ctx.Done()`.
- `filterErrors(ctx context.Context, in <-chan LogEntry, numWorkers int) <-chan LogEntry` rodando `numWorkers` goroutines de filtro que fazem fan-in da saída de volta pra um único channel.
- `sink(ctx context.Context, in <-chan LogEntry, ratePerSecond int) <-chan SinkResult` processando no máximo `ratePerSecond` itens por segundo e retornando os resultados em outro channel.

## Decisões de design

- O channel de fan-in do `filterErrors` é fechado por uma goroutine coordenadora que espera num `sync.WaitGroup` por todos os workers de filtro, não por nenhum worker individual, evitando um panic de "send on closed channel".
- O `context` é passado pelas três funções de estágio, então o cancelamento para o pipeline em qualquer ponto, não só no generator.

## Como rodar

```bash
go run .
```
