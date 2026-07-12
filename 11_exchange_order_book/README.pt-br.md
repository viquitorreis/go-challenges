# Exchange Order Book

🇺🇸 [English version](README.md)

**Categoria:** Estruturas de Dados
**Tempo estimado:** ~1h30

## O que é

Um motor de matching com price-time priority pra um único ticker: bids e asks ficam em heaps separadas, e ordens casam sempre que o melhor preço de bid é pelo menos igual ao melhor preço de ask.

## O que você aprende

- Modelar um order book com duas heaps: uma max-heap pra bids, uma min-heap pra asks, ambas ordenadas por preço e, em caso de empate, por horário de chegada (price-time priority).
- A condição de matching `max(bids) >= min(asks)` e como funcionam os fills parciais quando as quantidades das ordens não batem exatamente.
- Implementar `heap.Interface` duas vezes pra duas ordenações diferentes sobre dados estruturalmente parecidos.

## O que foi implementado

- `bidsHeap` e `asksHeap`, ambas implementando `heap.Interface` com ordenação de preço oposta.
- `NewOrderBook(symbol string) *OrderBook`, `NewOrder(id, userID string, side Side, price, qty int) *Order`.
- `AddOrder(o *Order)` inserindo na heap do lado correto.
- `Match() []Trade` executando todos os matches possíveis no momento e retornando os trades resultantes.
- `Cancel(orderID string) bool`, `BidDepth() int`, `AskDepth() int`.
- Os testes cobrem adicionar e casar ordens, cenários sem match, cancelamento, ordenação por price-time priority, e submissão concorrente de ordens.

## Decisões de design

- Bids e asks usam dois tipos de heap separados em vez de uma heap genérica com checagem de lado em runtime, mantendo o comparador de ordenação simples e correto pra cada lado.
- O matching é uma chamada explícita e separada, `Match()`, em vez de casar ordens inline a cada `AddOrder`, o que mantém o caminho de inserção barato e facilita testar o comportamento de matching isoladamente.

## Como rodar

```bash
make execute
# ou
go run .
go test ./...
```
