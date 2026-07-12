# Rate Limiter (Token Bucket)

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~1 hora

## O que é

Um rate limiter thread-safe baseado no algoritmo token bucket: o bucket comporta até N tokens, se reabastece numa taxa fixa, e cada requisição consome um token ou é rejeitada se não sobrar nenhum.

## O que você aprende

- O algoritmo token bucket e como ele difere do leaky bucket (token bucket permite bursts até a capacidade máxima; leaky bucket mantém uma taxa de drenagem constante).
- Modelar o estado do bucket sem guardar um array de tokens: só rastreando uma contagem e o timestamp do último refill.
- Refill preguiçoso: calcular quantos tokens deveriam existir com base no tempo decorrido, em vez de rodar uma goroutine de background num timer a cada tick.
- Onde colocar o mutex pra que chamadas concorrentes de `Allow()` continuem corretas sem virar um gargalo.

## O que foi implementado

- `NewTokenBucket(capacity uint64, refillRate int, refillInterval time.Duration) *TokenBucket`.
- `Allow() bool` - tenta consumir um token, reabastecendo antes se tempo suficiente passou.
- `Stats() (available, capacity uint64)` pra introspecção.
- Os testes cobrem contagem inicial de tokens, consumir todos os tokens, comportamento de refill ao longo do tempo, acesso concorrente, correção do rate limiting sob carga, e checagem com `-race`.

## Decisões de design

- O refill é calculado de forma preguiçosa dentro de `Allow()`/`refill()`, com base no tempo decorrido desde o último refill, evitando uma goroutine de timer dedicada por bucket.
- Um único mutex protege tanto a contagem de tokens quanto o timestamp do último refill, já que os dois são lidos e atualizados juntos em toda chamada de `Allow()`.

## Como rodar

```bash
go run .
go test -race ./...
```
