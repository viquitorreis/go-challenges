# Trie Thread-Safe

🇺🇸 [English version](README.md)

**Categoria:** Estruturas de Dados / Concorrência
**Tempo estimado:** ~1 hora

## O que é

Uma trie no estilo autocomplete (do tipo que roda por trás da busca de repositórios do GitHub enquanto você digita) que suporta insert, busca exata, busca por prefixo e autocomplete concorrentes, sem um lock global transformando cada operação num gargalo.

## O que você aprende

- Locking granular por node ("hand-over-hand"): adquirir e liberar o mutex de cada node conforme você desce na árvore, em vez de travar a estrutura inteira.
- Por que isso importa pra lookups de prefixo concorrentes de alto throughput (o mesmo problema que roteadores HTTP como Gin/Echo, servidores DNS, e tabelas de roteamento de IP resolvem).
- Tirar um snapshot dos filhos de um node sob lock antes de recursar, pra que a travessia não segure o lock do pai enquanto percorre a subárvore.

## O que foi implementado

- `NewTrie() *Trie`.
- `Insert(word string)`, `Search(word string) bool`, `StartsWith(prefix string) bool`.
- `AutoComplete(prefix string, limit int) []string`, retornando até `limit` resultados (ou todos, quando `limit` é 0).
- Os testes cobrem insert/search básico, busca por prefixo, autocomplete (com e sem limite), prefixos inexistentes, trie vazia, palavras de um caractere, palavras sobrepostas, inserts/reads/operações mistas concorrentes, caracteres unicode, um dataset maior, e benchmarks de insert/search/autocomplete/inserts concorrentes.

## Decisões de design

- Cada `TrieNode` tem seu próprio `sync.RWMutex`, então buscas em prefixos não relacionados (ex: "react" vs "golang") nunca disputam o mesmo lock.
- `collectWords` tira um snapshot dos filhos de um node sob um lock breve antes de recursar em cada filho, em vez de segurar o lock do pai durante toda a travessia recursiva.

## Como rodar

```bash
go run .
go test ./...
go test -bench=. ./...
```
