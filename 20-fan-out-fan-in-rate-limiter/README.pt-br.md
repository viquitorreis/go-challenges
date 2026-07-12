# Fan-out/Fan-in Rate Limiter

🇺🇸 [English version](README.md)

**Categoria:** Concorrência
**Tempo estimado:** ~30 minutos

## O que é

Um simulador concorrente de processamento de requests de API: N goroutines worker processam requests em paralelo (fan-out), mas todas compartilham um único rate limit global, e os resultados delas convergem de volta pra um único channel (fan-in).

## O que você aprende

- A diferença entre um rate limit **por worker** e um rate limit **global** compartilhado entre todos os workers, e por que um `time.Ticker` compartilhado é o que torna o limite realmente global.
- Combinar fan-out (múltiplos workers puxando de uma fonte de jobs compartilhada) com fan-in (todos os resultados convergindo pra um channel) numa única função.

## O que foi implementado

- `ProcessRequests(requests []Request, numWorkers int, ratePerSecond int) []Result` como único ponto de entrada público.
- `numWorkers` workers processando requests concorrentemente, controlados por um único rate limiter compartilhado (não um limiter por worker).
- Processamento mockado (`time.Sleep(50ms)`, sempre com sucesso), pra manter o foco no padrão de concorrência em vez de I/O real.

## Decisões de design

- O rate limiter é um único `time.Ticker` compartilhado (ou equivalente) que todo worker lê antes de processar um request, em vez de cada worker ter seu próprio ticker, o que é o que torna o limite global em vez de `numWorkers * ratePerSecond`.
- Fan-in e fan-out são combinados numa única função em vez de separados em helpers exportados distintos, já que o escopo do challenge é exatamente esse padrão combinado.

## Como rodar

```bash
go run .
```
