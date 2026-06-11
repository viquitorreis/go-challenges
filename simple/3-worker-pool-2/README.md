Implemente uma função que processa uma lista de jobs concorrentemente com um número fixo de workers:

```go
func processJobs(jobs []int, numWorkers int) []int
// Cada job é um número
// Cada worker eleva ao quadrado: job * job
// Retorna os resultados (ordem não importa)
```

- Requisitos:

numWorkers goroutines no máximo rodando ao mesmo tempo
Use channels
Sem goroutine por job

jobs = [1, 2, 3, 4, 5], numWorkers = 2

Worker 1 pega job 1 -> resultado 1
Worker 2 pega job 2 -> resultado 4
Worker 1 pega job 3 -> resultado 9
...
resultado final: [1, 4, 9, 16, 25] (qualquer ordem)

Estrutura que você precisa:

```
jobsCh   -> channel de entrada (você envia os jobs)
resultCh -> channel de saída (workers enviam resultados)
```

- numWorkers goroutines, cada uma lê de jobsCh até fechar
- Main envia todos os jobs em jobsCh e fecha
- Coleta todos os resultados de resultCh

O desafio: quando fechar resultCh? Pensa nisso antes de codar.