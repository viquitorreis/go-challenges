# Idempotent Payment Processing (PostgreSQL)

🇺🇸 [English version](README.md)

**Categoria:** Sistemas Distribuídos
**Tempo estimado:** ~2 horas

## O que é

Um serviço de processamento de pagamentos que garante que a mesma chave de idempotência nunca resulte numa cobrança duplicada, mesmo sob retries concorrentes, shardeando requisições por chave e processando cada shard sequencialmente.

## O que você aprende

- Por que idempotência importa pra APIs de pagamento: timeouts de rede e crashes de servidor significam que o cliente vai tentar de novo, e o retry precisa ser seguro.
- Sharding por chave de idempotência como alternativa a um advisory lock global único, pra que chaves diferentes processem em paralelo enquanto chaves iguais são sempre serializadas.
- Usar `ON CONFLICT DO NOTHING` na camada do banco como a rede de segurança final por baixo do sharding em memória.

## O que foi implementado

- `NewDB()` e `createTables(db *sql.DB) error` configurando o schema do PostgreSQL.
- `NewPaymentService(db *DB, numShards int) *PaymentService`.
- `ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error)` como o ponto de entrada público.
- `hashKey(key string, shards int) int64` roteando uma requisição pra um shard fixo com base na chave de idempotência.
- `Shard.worker(db *DB)`: uma goroutine por shard, processando as requisições daquele shard uma de cada vez.
- `processStripeCharge(...)` como uma chamada de cobrança mockada (sem integração real com Stripe).
- Os testes cobrem idempotência pra uma única chave, requisições concorrentes com a mesma chave, chaves diferentes processando em paralelo, e múltiplas chaves distintas.

## Decisões de design

- Cada shard tem exatamente uma goroutine worker, então requisições com a mesma chave de idempotência são serializadas naturalmente, sem precisar de um lock por chave.
- `hashKey` é determinístico, então a mesma chave sempre cai no mesmo shard entre retries.
- A constraint `ON CONFLICT DO NOTHING` do banco é a última linha de defesa caso duas requisições com a mesma chave, por algum motivo, cheguem ao banco concorrentemente.

## Como rodar

Precisa de uma instância PostgreSQL rodando (detalhes de conexão em `db.go`).

```bash
make execute
# ou
go run .
go test ./...
```
