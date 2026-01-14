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

- Publicar evento quando não há subscribers -> evento é perdido (ok)
- Subscribers recebem eventos na ordem publicada
- Se um subscriber está lento, não bloqueia outros subscribers
- Fechar o bus -> fecha todos os channels de subscribers

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

## Log Aggregator (Fan-in)

### O Desafio

Você tem múltiplos serviços (API, Database, Cache) gerando logs concorrentemente. Precisa de um agregador central que coleta todos esses logs e garante que nenhum se perde.

Aprendizados desse challenge:

✅ Fan-in pattern (N -> 1) - múltiplos producers enviando para um único channel intermediário (bridge) que é consumido por uma goroutine
✅ for range em channels - o padrão idiomático para processar todos os valores até o channel fechar, bloqueia automaticamente quando não há dados
✅ Select com default cria busy-waiting - evitar usar default quando você quer bloquear esperando dados, channels já dormem eficientemente
✅ break dentro de select - só quebra o select, não o for externo (use return para sair da goroutine inteira)
✅ WaitGroup regra crítica: Add antes de Wait - nunca chamar Wait() antes que todos os Add() tenham sido executados, causa race condition no contador interno
✅ WaitGroup para coordenação - Add(1) antes de criar goroutine, defer wg.Done() na goroutine, Wait() para bloquear até todas terminarem
✅ Onde criar a goroutine auxiliar importa - se criar no Start() e ela chamar Wait() imediatamente, vai retornar antes dos Register() adicionarem ao WaitGroup; melhor chamar Wait() no Stop()
✅ Variable shadowing com := - usar bridge := make(chan) cria variável local que esconde la.bridge, deixando o campo do struct como nil; sempre usar la.bridge = make(chan) para inicializar campos
✅ Inicializar bridge no lugar certo - criar no Start() ao invés do construtor, senão a goroutine auxiliar pode fechar o channel antes de qualquer Register() acontecer
✅ Done channel para sincronizar término - usar chan struct{} que a goroutine consumidora fecha quando termina, permitindo Stop() esperar sem race condition em la.logs
✅ defer dentro de loop acumula - não executa a cada iteração, só no final da função (causa deadlock com mutex em loops)
✅ Single writer pattern - quando só uma goroutine escreve em uma estrutura, não precisa de mutex (sem race condition)
✅ Ownership de channels - quem cria o channel deve fechá-lo, não tente fechar channels de terceiros (causa panic)
✅ Graceful shutdown em camadas - producers fecham channels -> leitoras terminam e Done() -> Stop() fecha bridge -> consumidora termina e fecha done channel -> Stop() retorna logs
✅ Channel buffering - buffer no bridge permite leitoras continuarem enviando mesmo se consumidora estiver ocupada (trade-off latência vs throughput)
✅ Send on closed channel panic - tentar enviar para channel fechado causa panic; garantir que bridge só é fechado quando todas as goroutines leitoras já terminaram

### Os 3 Problemas Principais

**1. Fan-in (N -> 1)**: Como ler de múltiplos channels ao mesmo tempo?
- Não dá pra fazer `for` em cada channel separado (seria sequencial)
- Precisa de algo que espera em TODOS simultaneamente

**2. Graceful Shutdown**: Como saber quando TODOS os producers terminaram?
- Um producer fecha seu channel... mas ainda tem outros rodando
- Você só pode parar quando o ÚLTIMO fechar

**3. Coordenação**: Como avisar o aggregator que pode parar?
- Quem vai fechar o channel central?
- Como sincronizar isso com os producers?

### De forma simples:

**Fan-in**

- Vários produtores.
- Um consimudor.

**Produtores:** criam logs, api, database, etc
**Consumidor:** aggregator.

Aggregator deve registrar os channels para ler depois. Estrutura simples.

**O que é Fan-in de verdade?**

Fan-in significa que você precisa estar lendo de TODOS os channels ao mesmo tempo.
Quando qualquer um deles tiver um log pronto, processa. É concorrente, não sequencial.

### Requisitos Funcionais

1. **LogAggregator deve permitir:**
   - `Register(logChan <-chan LogEntry)` - registra um channel de producer para agregar
   - `Start()` - inicia o processamento de logs de todos os producers
   - `Stop() []LogEntry` - para gracefully e retorna todos os logs coletados

2. **Comportamento esperado:**
   - Múltiplos producers podem enviar logs simultaneamente
   - Nenhum log pode ser perdido (todos devem ser coletados)
   - Se um producer termina antes dos outros, seus logs já devem estar agregados
   - `Stop()` só retorna quando TODOS os producers terminaram e todos os logs foram processados
   - Producers mais lentos não bloqueiam producers mais rápidos

### Constraints

- Não usar libs externas (só stdlib)
- Tempo: ~1 hora
- Código: 80-160 linhas
- Usar `sync` package conforme necessário
- Implementar graceful shutdown sem perder logs

### Aprendizados desse challenge:

### Template

```go

func main() {
	fmt.Println("=== Log Aggregator Challenge ===")
	fmt.Println("Fan-in Pattern: N producers -> 1 aggregator")

	agg := NewLogAggregator()

	// simulando 3 serviços diferentes gerando logs
	services := []string{"api", "database", "cache"}

	for _, service := range services {
		logChan := make(chan LogEntry, 10)
		agg.Register(logChan)

		// Cada serviço roda em sua própria goroutine
		go func(name string, ch chan LogEntry) {
			for i := 0; i < 5; i++ {
				ch <- LogEntry{
					Timestamp: time.Now(),
					Source:    name,
					Level:     "INFO",
					Message:   fmt.Sprintf("Log %d from %s", i+1, name),
				}
				time.Sleep(time.Millisecond * 10) // simulando trabalho
			}
			fmt.Printf("[%s] Finished producing logs\n", name)
			close(ch) // PRECIOSAMOS fechar o channel depois de terminar
		}(service, logChan)
	}

	agg.Start()

	// aguardar logs serem gerados
	time.Sleep(time.Millisecond * 200)

	fmt.Println("\nStopping aggregator...")
	logs := agg.Stop()

	fmt.Printf("\n=== Collected %d logs ===\n", len(logs))

	// agrupa por source
	bySource := make(map[string]int)
	for _, log := range logs {
		bySource[log.Source]++
		// fmt.Printf("[%s] %s: %s\n", log.Source, log.Level, log.Message)
	}

	fmt.Println("\nLogs per service:")
	for service, count := range bySource {
		fmt.Printf("  %s: %d logs\n", service, count)
	}

	fmt.Println("\nRun 'go test -v' to verify your implementation")
	fmt.Println("Run 'go test -race' to check for race conditions")
}


// LogEntry representa um log de qualquer serviço
type LogEntry struct {
	Timestamp time.Time
	Source    string // "api", "database", "cache", etc
	Level     string // "DEBUG", "INFO", "WARN", "ERROR"
	Message   string
}

// LogProducer simula um serviço gerando logs
type LogProducer struct {
	id       string
	logsChan chan LogEntry
}

// LogAggregator é o coração do Fan-in pattern
type LogAggregator struct {
	// TODO: adicione campos necessários para:
	// - Armazenar channels dos producers
	// - Coordenar o shutdown de todos producers
	// - Coletar logs processados
}

func NewLogAggregator() *LogAggregator {
	return &LogAggregator{
		// TODO: inicialize seus campos
	}
}

// Register adiciona um novo producer ao aggregator
func (la *LogAggregator) Register(logChan <-chan LogEntry) {
	// TODO: armazene o channel do producer
	// Hint: você vai precisar ler de TODOS esses channels depois
}

// Start inicia o processamento de logs
func (la *LogAggregator) Start() {
	// TODO: inicie uma goroutine que:
	// 1. Lê de TODOS os channels registrados (Fan-in!)
	// 2. Processa cada log recebido
	// 3. Para quando todos os producers terminarem

	// Hint: como ler de múltiplos channels ao mesmo tempo?
	// Hint: como saber quando TODOS os channels foram fechados?
}

// Stop para o aggregator e retorna todos os logs coletados
func (la *LogAggregator) Stop() []LogEntry {
	// TODO: implemente graceful shutdown
	// - Espere todos os producers terminarem
	// - Retorne os logs coletados

	// Hint: você precisa sinalizar que quer parar E esperar
	// que o processamento realmente termine

	return nil // TODO: retorne os logs
}

```