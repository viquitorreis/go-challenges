# Concurrent Log File Analyzer

🇺🇸 [English version](README.md)

**Categoria:** Concorrência / I/O
**Tempo estimado:** ~35-40 minutos

## O que é

Um analisador de arquivo de log que lê um arquivo (potencialmente grande) linha por linha e conta entradas por nível (INFO/WARN/ERROR), usando um leitor sequencial, N workers de parsing, e uma goroutine agregadora.

## O que você aprende

- Por que ler um arquivo texto é inerentemente sequencial (você não consegue pular pro meio de um arquivo sem já saber onde as linhas começam), enquanto o parsing por linha é exatamente onde a concorrência agrega valor quando o trabalho por linha fica caro.
- O formato leitor -> N workers -> 1 agregador como um modelo geral pra processamento de log/dados orientado a linha sem carregar o arquivo inteiro na memória.
- Por que uma única goroutine agregadora escrevendo num `map[string]int` evita precisar de mutex nesse map, mesmo com múltiplos workers produzindo resultados concorrentemente.

## O que foi implementado

- `AnalyzeLog(filename string, numWorkers int) (map[string]int, error)` como ponto de entrada público.
- Uma única goroutine lendo o arquivo com `bufio.Scanner`, mandando cada linha pra um channel de jobs.
- `extractLevel(bt []byte) string` extraindo o nível de uma linha (o texto entre `[` e `]`).
- `numWorkers` workers consumindo linhas e fazendo o parsing concorrentemente.
- Uma única goroutine agregadora coletando os resultados parseados no `map[string]int` final.
- `createLogs()` pra gerar um arquivo `access.log` de exemplo pra testar.
- Linhas malformadas (sem o formato `[LEVEL]` esperado) são ignoradas em vez de quebrar o programa.

## Decisões de design

- Só uma goroutine escreve no map de resultado, e é isso que torna seguro sem mutex, a concorrência fica no estágio de parsing, não no de agregação.
- Os channels são fechados numa ordem específica: o leitor fecha o channel de jobs quando o arquivo termina, e os workers fecham o channel de resultados (via `WaitGroup`) só depois que todos terminaram, não antes.

## Como rodar

```bash
go run .
```
