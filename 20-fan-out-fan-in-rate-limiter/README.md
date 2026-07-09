# Challenge - Fan-out/Fan-in com Rate Limiter (30min)

## Task: Concurrent API Request Simulator

Você tem 20 "requests" pra processar. Precisa: (1) processar com fan-out de N workers, (2) aplicar um rate limiter de R requests/segundo globalmente (não por worker - global), (3) fan-in os resultados num único channel.

```
go
type Request struct {
    ID int
}

type Result struct {
    RequestID int
    Success   bool
    Duration  time.Duration
}

func ProcessRequests(requests []Request, numWorkers int, ratePerSecond int) []Result
```

### Regras

- numWorkers workers em paralelo (fan-out)
- Rate limiter global - mesmo com 5 workers, não pode passar de ratePerSecond requests/segundo no total (use time.Ticker compartilhado entre os workers)
- Mock de processamento: time.Sleep(50 * time.Millisecond), sempre sucesso
- Todos os resultados convergem num único channel (fan-in), depois coletados num slice
- Sem deadlock, sem goroutine leak, todos os 20 requests processados

**O que será avaliado*8

- Ticker compartilhado entre múltiplos workers sem corrida de dados
- Fan-out + fan-in combinados corretamente
- Fechar channels na ordem certa