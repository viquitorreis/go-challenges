# Worker Pool + Pipeline - 40min

## Task: Log Processing Pipeline with Rate-Limited Output

Simula um sistema que processa logs de múltiplas fontes, filtra erros, e envia para um "sink" rate-limited (ex: API externa com limite de requests/seg).

### Pipeline (3 stages)

Generator -> Filter (workers) -> Rate-Limited Sink

### Tipos

```go
type LogEntry struct {
    Source    string
    Level     string // "info", "warn", "error"
    Message   string
    Timestamp time.Time
}

type SinkResult struct {
    Entry    LogEntry
    SentAt   time.Time
    Err      error
}
```

### Requisitos

**Stage 1 - Generator**

```go
func generateLogs(ctx context.Context, n int) <-chan LogEntry
```

- Gera 'n' logs mockados, niveis aleatorios (info/warn/error), em uma goroutine
- Respeita `ctx.Done()`

**Stage 2** - Filter (worker pool)

```go
func filterErrors(ctx context.Context, in <-chan LogEntry, numWorkers int) <-chan LogEntry
```

**Stage 3** - Rate-Limited Sink

```go
func sink(ctx context.Context, in <-chan LogEntry, ratePerSecond int) <-chan SinkResult
```

- Processa no máximo ratePerSecond itens por segundo (use time.Ticker ou time.Sleep)
- Mock de "envio": time.Sleep(10*time.Millisecond), sempre sucesso
- Retorna resultados em outro channel

**main**

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    logsCh := generateLogs(ctx, 50)
    errorsCh := filterErrors(ctx, logsCh, 4)
    resultsCh := sink(ctx, errorsCh, 5) // 5/sec

    count := 0
    for res := range resultsCh {
        fmt.Println(res)
        count++
    }
    fmt.Println("total processed:", count)
}
```

**O que será avaliado**

- Fan-in correto (múltiplos workers -> 1 channel, sem fechar prematuramente)
- Rate limiting funcional (medir tempo total: ~deve bater com errors_count / ratePerSecond)
- Propagação de ctx em todos os 3 estágios
- Fechamento correto de channels em cadeia (cada stage fecha seu próprio output quando termina)
- Sem deadlock, sem goroutine leak

**Twist proposital**: o fan-in do stage 2 é o ponto mais fácil de errar **múltiplos workers escrevendo no mesmo channel**, e alguém precisa fechar esse channel só depois que TODOS os workers terminarem (já viu esse padrão antes, hoje mesmo).