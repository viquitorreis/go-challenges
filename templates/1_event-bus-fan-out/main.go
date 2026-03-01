package main

import (
	"fmt"
	"time"
)

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
		for range counter {
			eventCount++
			fmt.Printf("[COUNTER] Total events: %d\n", eventCount)
		}
	}()

	// some time to enable subscribers to start
	time.Sleep(100 * time.Millisecond)

	// publisher sends events
	bus.Publish("user.login")
	bus.Publish("order.created")
	bus.Publish("user.logout")

	// wait time to process
	time.Sleep(500 * time.Millisecond)

	bus.Close()
	time.Sleep(100 * time.Millisecond)

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
