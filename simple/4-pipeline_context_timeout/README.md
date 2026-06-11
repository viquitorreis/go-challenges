
Pipeline com context e timeout:

Implemente uma pipeline que:

1. Gera números de 1 a 10 (producer)
2. Filtra só os pares (filter)
3. Dobra o valor (transform)
4. Para tudo se demorar mais de 200ms

```go
func run(ctx context.Context) []int
```

Requisitos:

- Cada etapa é uma goroutine separada comunicando por channel
- Se o context cancelar, todas as etapas param limpo — sem goroutine leak
- Use select com ctx.Done()

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    result := run(ctx)
    fmt.Println(result)
}
```

**15 minutos.**