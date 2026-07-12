# Skip List

🇺🇸 [English version](README.md)

**Categoria:** Estruturas de Dados / Concorrência
**Tempo estimado:** ~1h30

## O que é

Uma skip list probabilística thread-safe, a estrutura por trás dos sorted sets do Redis (`ZADD`, `ZRANGE`, `ZRANK`) e dos storage engines do LevelDB/RocksDB: múltiplas linked lists empilhadas onde cada node "sobe de nível" aleatoriamente pra dar busca, insert e delete O(log n) amortizado.

## O que você aprende

- Por que o balanceamento probabilístico (nível decidido por cara-ou-coroa) é mais simples de implementar e de tornar concorrente do que uma BST balanceada deterministicamente.
- O **update array pattern**: percorrer cada nível de cima pra baixo, guardando o último node "à esquerda" do ponto de inserção em cada nível, e depois reconectar os ponteiros usando esses predecessores.
- Locking horizontal (através dos níveis da lista) como uma forma de concorrência diferente do locking vertical (por node, pai-pra-filho) da trie.

## O que foi implementado

- `NewSkipList(maxLevel int, p float64, rng *rand.Rand) *SkipList`.
- `Insert(score int, value any)`, `Search(score int) (any, bool)`, `Delete(score int) bool`.
- `RangeSearch(min, max int) []any` e `Size() int`.
- `randomLevel()` implementando a atribuição de nível por cara-ou-coroa com probabilidade `p`.
- Os testes cobrem insert/search, atualizar um score existente, delete, range search (incluindo lista vazia), ordem de inserção, e inserts/leituras-e-escritas/deletes concorrentes.

## Decisões de design

- O update array é calculado uma vez por insert/delete varrendo do nível mais alto pro mais baixo, evitando travessias completas repetidas por nível.
- `p = 0.5` é usado como probabilidade de nível, a escolha padrão que mantém aproximadamente metade dos nodes em cada nível sucessivo.

## Como rodar

```bash
make execute
# ou
go run .
go test ./...
```
