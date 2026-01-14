package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Log Aggregator Challenge ===")
	fmt.Println("Fan-in Pattern: N producers → 1 aggregator")

	agg := NewLogAggregator()
	agg.Start()

	// simulando 3 serviços diferentes gerando logs
	services := []string{"api", "database", "cache"}

	go func() {
		for _, service := range services {
			logChan := make(chan LogEntry, 10)
			agg.Register(logChan)

			// Cada serviço roda em sua própria goroutine
			go func(name string, ch chan LogEntry) {
				rand := rand.Intn(200)
				for i := 0; i < rand; i++ {
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
	}()

	time.Sleep(time.Millisecond * 200)

	fmt.Println("\nStopping aggregator...")
	logs := agg.Stop()

	fmt.Printf("\n=== Collected %d logs ===\n", len(logs))

	// agrupa por source
	bySource := make(map[string]int)
	for _, log := range logs {
		bySource[log.Source]++
		fmt.Printf("[%s] %s: %s\n", log.Source, log.Level, log.Message)
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

// LogAggregator é o coração do Fan-in pattern
type LogAggregator struct {
	logs         []LogEntry
	bridge       chan LogEntry // leitura async sem block será intermediada por aqui
	wg           sync.WaitGroup
	consumerDone chan struct{}
}

func NewLogAggregator() *LogAggregator {
	return &LogAggregator{
		logs: make([]LogEntry, 0),
		// idealmente, nao crie os bridges aqui.
		// quando o Start() é chamado, ele cria a goroutine auxiliar
		// se nao tiver nada no WaitGroup ele termina o channel e nao vamos conseguir enviar a ele
	}
}

// Register adiciona um novo producer ao aggregator
// o agregador só recebe os logs e agrega eles, por isso receive-only channel
func (la *LogAggregator) Register(logChan <-chan LogEntry) {
	// goroutines de leitura: cada Register cria uma o channel especifico
	la.wg.Add(1)
	go func(logEntry <-chan LogEntry) {
		defer la.wg.Done()

		for log := range logEntry {
			la.bridge <- log
		}
	}(logChan)
}

// Start inicia o processamento de logs
func (la *LogAggregator) Start() {
	la.bridge = make(chan LogEntry, 100)
	la.consumerDone = make(chan struct{})

	// Goroutine consumidora
	// ela vai consumir de forma async, nao precisa de nenhum mutex
	// NAO adiciona ao WaitGroup porque queremos esperar só as goroutines que leem no Stop()
	// essa é só a bridge de processamento
	go func() {
		for log := range la.bridge {
			la.logs = append(la.logs, log)
		}
		close(la.consumerDone)
	}()
}

func (la *LogAggregator) Stop() []LogEntry {
	la.wg.Wait() // espera goroutines de leitura terminar (producers)
	close(la.bridge)
	<-la.consumerDone // espera goroutine consumidora terminar (consumer)
	return la.logs
}
