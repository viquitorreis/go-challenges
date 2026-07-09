# Challenge - Graceful Shutdown com Worker Pool (45min)

Worker pool com graceful shutdown - sem http.Server.Shutdown() que faz tudo por você. Você implementa o drain manualmente.

# Task

Um sistema que processa jobs de uma fila continuamente. Quando recebe SIGTERM/SIGINT:

- Para de aceitar jobs novos
- Termina os jobs em andamento (não mata no meio)
- Loga o progresso do shutdown
- Sai limpo

```go
type Job struct {
    ID      int
    Payload string
}

type WorkerPool struct {
    // você decide
}

func NewWorkerPool(numWorkers int) *WorkerPool
func (wp *WorkerPool) Submit(job Job) error // retorna erro se shutdown iniciado
func (wp *WorkerPool) Shutdown()             // drain + espera workers terminarem
func (wp *WorkerPool) Start(ctx context.Context)
```

## Comportamento esperado

```
2026/06/24 13:40:00 worker pool started with 3 workers
2026/06/24 13:40:01 job 1 started (worker 2)
2026/06/24 13:40:01 job 2 started (worker 1)
2026/06/24 13:40:01 job 3 started (worker 3)
^C
2026/06/24 13:40:02 shutdown signal received
2026/06/24 13:40:02 no new jobs accepted
2026/06/24 13:40:03 job 1 finished
2026/06/24 13:40:03 job 3 finished
2026/06/24 13:40:03 job 2 finished
2026/06/24 13:40:03 shutdown complete
```

## Main

```go
func main() {
    pool := NewWorkerPool(3)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    pool.Start(ctx)

    // simula jobs chegando continuamente
    go func() {
        for i := 1; i <= 20; i++ {
            err := pool.Submit(Job{ID: i, Payload: fmt.Sprintf("task-%d", i)})
            if err != nil {
                log.Printf("job %d rejected: %v", i, err)
                break
            }
            time.Sleep(300 * time.Millisecond)
        }
    }()

    // graceful shutdown no sinal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    <-sigCh
    log.Println("shutdown signal received")
    pool.Shutdown()
    log.Println("shutdown complete")
}
```

## Requisitos

- Submit retorna errors.New("pool is shutting down") se Shutdown já foi chamado
- Shutdown fecha o jobs channel (para de aceitar novos), espera todos os workers terminarem os jobs em andamento via WaitGroup
- Workers processam jobs com time.Sleep(1 * time.Second) (mock de trabalho real)
- Sem panic de "send on closed channel" se Submit e Shutdown rodarem concorrentemente ctx propagado pros workers - se o context cancelar, workers param de pegar novos jobs mas terminam o job atual antes de sair

## O que será avaliado

- Submit thread-safe contra Shutdown concorrente (como evitar send em channel fechado?)
- Drain correto: close(jobsCh) + wg.Wait()
- Workers que terminam o job atual antes de responder ao cancelamento
- sync.Once pra garantir que Shutdown só executa uma vez mesmo se chamado múltiplas vezes