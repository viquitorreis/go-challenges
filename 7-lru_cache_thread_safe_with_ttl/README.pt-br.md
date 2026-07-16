# LRU Cache com TTL

🇺🇸 [English version](README.md)

**Categoria:** Estruturas de Dados / Concorrência
**Tempo estimado:** ~1h30

## O que é

Um cache em memória no estilo produção que combina eviction LRU (least recently used) com expiração por TTL (time to live) por item, seguro pra leituras e escritas concorrentes.

## O que você aprende

- Combinar um hashmap com uma lista duplamente encadeada pra obter `Get`/`Set`/eviction O(1), a implementação clássica de LRU.
- O trade-off entre expiração de TTL preguiçosa (checada no acesso) e limpeza ativa (uma goroutine de background varrendo periodicamente) - e por que essa implementação faz as duas.
- Escolher entre um `sync.Mutex` único, `sync.RWMutex`, ou sharding pra um cache sob carga concorrente.

## O que foi implementado

- `NewLRUCache(capacity int, ttl time.Duration) *LRUCache`.
- `Get(key string) (any, bool)` - retorna o valor se presente e não expirado, e move pra frente da lista (mais recentemente usado).
- `Set(key string, value any)` - insere ou atualiza, removendo o item menos recentemente usado quando a capacidade é excedida.
- `Delete(key string)` e `Size() int`.
- Helpers internos da lista duplamente encadeada: `addToFront`, `remove`, `moveToFront`.
- `cleanup()` rodando em background pra purgar ativamente itens expirados, mais `Close()` pra pará-lo.
- Os testes cobrem operações básicas, eviction, ordenação LRU, updates, expiração de TTL, refresh de TTL no acesso, acesso concorrente, limpeza em background, e `-race`.

## Decisões de design

- A lista encadeada rastreia a ordem de uso em O(1) por operação; o hashmap dá lookup de chave O(1) - nenhum dos dois sozinho é suficiente pra LRU O(1) de verdade.
- A limpeza de TTL é preguiçosa (checada no `Get`) e ativa (goroutine de background) ao mesmo tempo, então entradas expiradas não ficam presas em memória indefinidamente mesmo que ninguém as leia.

## Como rodar

```bash
go run .
go test -race ./...
```
