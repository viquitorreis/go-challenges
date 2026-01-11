# Challenges Customizados

## Event Bus

### Aprendizados desse challenge:

✅ Fan-out pattern - um evento distribuído para múltiplos subscribers
✅ Closure bug em goroutines - passar parâmetros vs capturar variáveis do loop
✅ Thread-safety - usar sync.RWMutex com maps compartilhados
✅ Buffered vs unbuffered channels - trade-offs de performance
✅ Ordem vs concorrência - quando goroutines ajudam e quando atrapalham
✅ Trade-offs de sistemas distribuídos - consistency vs availability

Implemente um sistema simples de Event Bus onde:

- Publishers enviam eventos (strings como "user.login", "order.created")
- Subscribers se registram para receber todos os eventos
- Quando um evento é publicado, todos os subscribers recebem concorrentemente
- Cada subscriber processa eventos de forma independente

### Requisitos Funcionais

1. EventBus deve permitir

- Subscribe(name string) <-chan string - registra um subscriber, retorna channel read-only
- Publish(event string) - envia evento para todos subscribers
- Close() - fecha o bus e todos os channels de subscribers

2. Comportamento esperado

- Publicar evento quando não há subscribers → evento é perdido (ok)
- Subscribers recebem eventos na ordem publicada
- Se um subscriber está lento, não bloqueia outros subscribers
- Fechar o bus → fecha todos os channels de subscribers

**Constraints**

- Não usar libs externas (só stdlib)
- Implemente uma fila para cada subscriber
- Tempo: ~1 hora
- Código: 80-120 linhas
- Usar sync package conforme necessário

### Func main() para testes:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Exemplo de uso
func main() {
    bus := NewEventBus()
    
    // Subscriber 1: Logger
    logger := bus.Subscribe("logger")
    go func() {
        for event := range logger {
            fmt.Printf("[LOGGER] Received: %s\n", event)
        }
    }()
    
    // Subscriber 2: Counter
    counter := bus.Subscribe("counter")
    eventCount := 0
    go func() {
        for event := range counter {
            eventCount++
            fmt.Printf("[COUNTER] Total events: %d\n", eventCount)
        }
    }()
    
    // Publisher envia eventos
    time.Sleep(100 * time.Millisecond) // dar tempo pros subscribers iniciarem
    
    bus.Publish("user.login")
    bus.Publish("order.created")
    bus.Publish("user.logout")
    
    time.Sleep(500 * time.Millisecond) // dar tempo para processar
    
    bus.Close()
    time.Sleep(100 * time.Millisecond) // dar tempo para fechar
    
    fmt.Println("Event bus closed")
}

// EventBus gerencia publishers e subscribers
type EventBus struct {
    // TODO: adicione campos necessários
    // Dica: você precisa guardar os channels dos subscribers
    // Dica: você precisa de um mutex para thread-safety
}

// NewEventBus cria um novo event bus
func NewEventBus() *EventBus {
    return &EventBus{
        // TODO: inicialize campos
    }
}

// Subscribe registra um novo subscriber e retorna um channel para receber eventos
func (eb *EventBus) Subscribe(name string) <-chan string {
    // TODO: 
    // 1. Criar um channel para este subscriber
    // 2. Guardar o channel internamente
    // 3. Retornar o channel (read-only)
}

// Publish envia um evento para todos os subscribers
func (eb *EventBus) Publish(event string) {
    // TODO:
    // 1. Iterar por todos os subscribers
    // 2. Enviar o evento para cada channel
    // 3. Usar goroutines para não bloquear OU algum mecanismo especifico de criação de subscribers que não bloqueie o envio (Qual? Tradeoffs?)
}

// Close fecha o event bus e todos os channels de subscribers
func (eb *EventBus) Close() {
    // TODO:
    // 1. Fechar todos os channels
    // 2. Limpar subscribers
}
```