# Challenge - Log File Analyzer Concorrente (35-40min)

## Task

Processar um arquivo de log grande, contando erros por categoria, usando leitura sequencial + processamento concorrente.

### Formato do arquivo (access.log)

```
2026-06-17 10:00:01 [INFO] request completed
2026-06-17 10:00:02 [ERROR] database timeout
2026-06-17 10:00:03 [WARN] high memory usage
2026-06-17 10:00:04 [ERROR] connection refused
2026-06-17 10:00:05 [INFO] request completed
```

### Requisitos

```
go
type LineResult struct {
    Level string // INFO, WARN, ERROR
    Count int
}

func AnalyzeLog(filename string, numWorkers int) (map[string]int, error)
```

- 1 goroutine lê o arquivo sequencialmente (bufio.Scanner, linha por linha - nunca carrega tudo na memória)
- Cada linha lida é mandada pra um channel de jobs
- N workers consomem desse channel, fazem o parsing (extrair o nível: INFO/WARN/ERROR) e mandam o resultado pra um channel de resultados
- 1 goroutine agregadora consome os resultados e popula o map[string]int final (contagem por nível)
- Retorna o mapa final: {"INFO": 2, "WARN": 1, "ERROR": 2}

### Por que isso é interessante (responde sua pergunta)

- A leitura é inerentemente sequencial (você não pode "pular" no meio de um arquivo texto sem saber onde as linhas começam)
- O paralelismo ganha valor quando o parsing/processamento de cada linha é caro (regex complexo, chamada de API, agregação pesada) - não na leitura em si, que já é rápida e é gargalo de I/O, não de CPU
- Pra esse exercício, o processamento é simples (extrair string entre [ e ]), mas a estrutura é a mesma que você usaria se fosse algo pesado

### Regras

- Crie um arquivo de teste com ~20-30 linhas (pode gerar via código ou escrever um .log manualmen­te)
- Se uma linha não tiver o formato esperado, ignora (não quebra o programa)
map[string]int final precisa ser thread-safe na agregação (você só tem 1 goroutine agregando, mas pense em por que isso evita precisar de mutex)

### O que será avaliado

- Separação correta: 1 reader sequencial, N workers paralelos, 1 aggregator
- Fechamento de channels na ordem certa (reader fecha jobsCh quando termina; workers fecham resultsCh via WaitGroup quando terminam)
- Sem necessidade de mutex no map final (porque só 1 goroutine escreve nele - pensa em por que essa escolha de design evita lock)
- Tratamento de linha malformada sem crash