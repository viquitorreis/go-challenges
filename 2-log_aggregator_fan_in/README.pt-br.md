# Log Aggregator (Fan-in)

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~1 hora

## O que é

Um agregador de logs que coleta entradas de múltiplos produtores independentes (simulando serviços como `api`, `database`, `cache`) num resultado único, usando o padrão fan-in (N produtores para 1 consumidor).

## O que você aprende

- **Fan-in (N para 1)**: ler de múltiplos channels concorrentemente através de um channel "bridge" compartilhado, em vez de fazer polling sequencial em cada um.
- Por que `select` com `default` vira busy-waiting, e por que um `for range` simples num channel é a forma idiomática de bloquear de forma eficiente.
- A regra crítica do `sync.WaitGroup`: todo `Add` precisa acontecer antes do `Wait` correspondente, senão o contador interno tem race condition.
- Graceful shutdown em camadas: produtores fecham seus channels primeiro, depois a bridge fecha, depois a goroutine agregadora retorna.
- Single-writer pattern: quando só uma goroutine escreve numa slice, não precisa de mutex.

## O que foi implementado

- `Register(logChan <-chan LogEntry)` - registra o channel de um produtor pra ser agregado.
- `Start()` - inicia a goroutine interna de fan-in que junta todos os channels registrados num só.
- `Stop() []LogEntry` - espera todos os produtores e o agregador terminarem, depois retorna tudo que foi coletado.
- Os testes cobrem produtor único, múltiplos produtores concorrentes, graceful shutdown, zero produtores, níveis de log misturados, e um teste de stress de concorrência.

## Decisões de design

- Um channel `done` sinaliza que a goroutine consumidora drenou completamente a bridge antes do `Stop()` retornar, evitando race conditions na slice de logs interna.
- O channel bridge é criado dentro do `Start()`, não no construtor, pra que a goroutine de fan-in não consiga fechá-lo antes de qualquer produtor se registrar.
- A propriedade dos channels é respeitada: cada produtor fecha seu próprio channel; o agregador nunca fecha um channel que não criou.

## Como rodar

```bash
go run .
go test ./...
```
