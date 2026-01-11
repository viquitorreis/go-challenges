package main

import (
	"fmt"
	"sync"
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

	// Publisher envia eventos
	time.Sleep(100 * time.Millisecond) // dar tempo pros subscribers iniciarem

	bus.Publish("user.login")
	bus.Publish("order.created")
	bus.Publish("user.logout")

	time.Sleep(1100 * time.Millisecond) // dar tempo para processar

	bus.Close()
	time.Sleep(100 * time.Millisecond) // dar tempo para fechar

	fmt.Println("Event bus closed")
}

type EventBus interface {
	Subscribe(name string) <-chan string // registra um subscribe -> retorna um channel READ ONLY
	Publish(event string)                // envia evento para todos os subscribers
	Close()                              // fecha o bus e todos os channels de subscribers
}

type eventBus struct {
	channelsSubs map[string]chan string
	queue        []string
	mu           sync.RWMutex
}

func NewEventBus() EventBus {
	return &eventBus{
		channelsSubs: make(map[string]chan string),
	}
}

func (e *eventBus) Subscribe(name string) <-chan string {
	// TODO:
	// 1. Criar um channel para este subscriber
	// 2. Guardar o channel internamente
	// 3. Retornar o channel (read-only)
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. podiamos usar waitgroup, que iria acrescentar ordem ao código, mas diminuir a performance (sistema bancário -> mais consistency)
	// 2. também podiamos criar buffered channels (sistemas de logging e etc -> availability)
	//		TRADEOFFS: se subscriber for lento o buffer enche e bloqueia
	//			qual o tamanho ideal do buffer? nesse caso é simples, mas em produção as vezes é dificil saber
	// 3. Podiamos criar um select statement (availability)
	// 		a. evenot recebido, envia com sucesso
	// 		b. no default subscriber estaria cheio/lento entao pula
	// 4. Podiamos criar uma fila para cada subscriber

	ch := make(chan string, 10)
	e.channelsSubs[name] = ch
	return ch
}

func (e *eventBus) Publish(event string) {
	// TODO:
	// 1. Iterar por todos os subscribers
	// 2. Enviar o evento para cada channel
	// 3. Usar goroutines para não bloquear OU algum mecanismo especifico de criação de subscribers que não bloqueie o envio (Qual? Tradeoffs?)
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, ch := range e.channelsSubs {
		// vai processar rapido, em teoria - temos buffer
		// TRADEOFF:
		//		Se o evento demorasse para ler da fila (consumer) e o buffer enchesse, teriamos problema -> consumers esperando eventos
		//		Porém, nesse exemplo Não é problema, vamos usar poucos eventos, e rápidos
		ch <- event
	}
}

func (e *eventBus) Close() {
	// TODO:
	// 1. Fechar todos os channels
	// 2. Limpar subscribers
	e.mu.Lock()
	defer e.mu.Unlock()

	// nao precisamos de goroutines para fechar os maps, é instantaneo, não bloqueia
	for _, ch := range e.channelsSubs {
		close(ch)
	}

	// para limpar todos os maps nao precisamos de delete nesse caso, pois queremos limpar todos
	e.channelsSubs = make(map[string]chan string)
}
