# Event Bus (Fan-out)

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~1 hora

## O que é

Um event bus pub/sub em memória: publishers enviam eventos (strings), e cada subscriber registrado ganha seu próprio channel read-only e recebe todos os eventos, de forma concorrente e independente dos outros.

## O que você aprende

- **Fan-out**: um evento distribuído pra N subscribers independentes.
- O clássico bug de **closure sobre variável de loop** em goroutines e como evitar.
- Proteger um map de subscribers compartilhado com `sync.RWMutex`.
- Channels bufferizados vs. não-bufferizados e o trade-off de throughput/backpressure que cada um traz.

## O que foi implementado

- `Subscribe(name string) <-chan string` - registra um subscriber e retorna um channel dedicado, read-only e bufferizado.
- `Publish(event string)` - envia um evento pro channel de cada subscriber registrado.
- `Close()` - fecha todos os channels de subscribers e limpa o map interno.
- Se não houver subscribers quando `Publish` roda, o evento é simplesmente descartado (sem erro, sem bloqueio) - esse é o comportamento documentado como aceitável pro challenge.

## Decisões de design

- Cada subscriber ganha seu **próprio channel bufferizado** (capacidade 10) em vez de compartilhar um só - um consumidor lento afeta só ele mesmo, não os outros.
- O bus é exposto como uma interface `EventBus`, com a struct `eventBus` privada implementando os detalhes - mantém a concorrência fora da API pública.
- Um único `sync.RWMutex` protege o map de subscribers; leituras (`Publish` iterando subscribers) poderiam usar `RLock`, escritas (`Subscribe`/`Close`) usam `Lock`.

## Como rodar

```bash
go run .
go test ./...
```
