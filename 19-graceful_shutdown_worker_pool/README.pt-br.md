# Graceful Shutdown Worker Pool

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~45 minutos

## O que é

Um worker pool que processa jobs de uma fila continuamente e, ao receber `SIGTERM`/`SIGINT`, para de aceitar jobs novos, termina os jobs já em andamento, e sai de forma limpa, tudo sem depender de um framework tipo `http.Server.Shutdown()` fazendo o drain por você.

## O que você aprende

- Implementar graceful shutdown na mão: parar de aceitar trabalho novo, deixar o trabalho em andamento terminar, e só então sair, nessa ordem específica.
- Coordenar o shutdown entre múltiplas goroutines de worker com cancelamento por `context` e um `sync.WaitGroup`.

## O que foi implementado

- `NewWorkerPool(numWorkers int) *WorkerPool`.
- `Start(ctx context.Context)` iniciando as goroutines de worker.
- `Submit(job Job) error` pra enfileirar novos jobs.
- `Shutdown()` drenando jobs em andamento antes de retornar.

## Decisões de design

- O shutdown é manual (não tem atalho de `http.Server.Shutdown()` disponível aqui, já que isso não é um servidor HTTP): `Shutdown()` para o intake explicitamente primeiro e só depois espera os workers drenarem o job atual deles.

## Como rodar

```bash
go run ./1
```

Nota: o código-fonte está na subpasta `1/` em vez de direto na raiz deste challenge; o comando de execução foi ajustado de acordo.
