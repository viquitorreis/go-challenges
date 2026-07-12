# Worker Pool Health Checker

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~30 minutos

## O que é

Um worker pool de tamanho fixo que checa a "saúde" de uma lista de URLs concorrentemente, aplicando um timeout por URL mesmo a função de checagem não aceitando `context.Context`.

## O que você aprende

- Aplicar um timeout de fora de uma função sem suporte a `context`: envolver a chamada e disputá-la contra `time.After` com `select`, em vez de alterar a função em si.
- Coletar resultados de N workers numa única slice com segurança, sem exigir que a ordem dos resultados bata com a ordem de entrada.
- Dimensionar um worker pool independentemente do número de jobs, e garantir que todo job apareça na saída.

## O que foi implementado

- `CheckURLs(urls []string, numWorkers int, timeout time.Duration) []CheckResult` como ponto de entrada público.
- `doWork(ctx context.Context, started time.Time, jobsCh chan string, resCh chan CheckResult)` rodado por cada uma das `numWorkers` goroutines.
- `mockCheck(url string) (bool, error)` simulando uma checagem HTTP com delay aleatório e uma falha determinística pra qualquer URL contendo `"bad"`.
- O timeout por URL é aplicado externamente, já que `mockCheck` não aceita context; uma URL que passa do `timeout` volta como `Healthy: false` com um erro de deadline excedido.

## Decisões de design

- `sync.WaitGroup` mais um channel de resultado é usado em vez de errgroup ou lib de terceiros, seguindo a própria restrição do challenge de ficar só na stdlib.
- A aplicação do timeout fica no ponto de chamada (`doWork`), não dentro de `mockCheck`, espelhando a situação real de envolver um timeout em cima de código que você não controla.

## Como rodar

```bash
go run .
```
