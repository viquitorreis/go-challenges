package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

func main() {
	agg := NewLogAggregator()
	agg.Start()

	// Fan-in Pattern: N producers -> 1 aggregator
	// simulating 3 different services generating logs
	services := []string{"api", "database", "cache"}

	go func() {
		for _, service := range services {
			logChan := make(chan LogEntry, 10)
			agg.Register(logChan)

			// each service runs on its own goroutine
			go func(name string, ch chan LogEntry) {
				rand := rand.Intn(200)
				for i := range rand {
					ch <- LogEntry{
						Timestamp: time.Now(),
						Source:    name,
						Level:     "INFO",
						Message:   fmt.Sprintf("Log %d from %s", i+1, name),
					}

					time.Sleep(time.Millisecond * 10) // simulating work
				}

				slog.Info("Finished producing logs", "name", name)

				// we need to close the channel after we finish
				close(ch)
			}(service, logChan)
		}
	}()

	time.Sleep(time.Millisecond * 200)

	slog.Info("Stopping aggregator...")
	logs := agg.Stop()

	slog.Info("Collected", "logs", len(logs))

	// group by source
	bySource := make(map[string]int)
	for _, log := range logs {
		bySource[log.Source]++
		slog.Info("Logs", "source", log.Source, "level", log.Level, "message", log.Message)
	}

	log.Println("Logs per service:")
	for service, count := range bySource {
		slog.Info("Source", "service", service, "count", count)
	}
}

type LogEntry struct {
	Timestamp time.Time
	Source    string // "api", "database", "cache", etc
	Level     string // "DEBUG", "INFO", "WARN", "ERROR"
	Message   string
}

type LogAggregator struct {
	logs         []LogEntry
	bridge       chan LogEntry // async read without block will be bridged by here
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

// Register adds a new producer to the aggregator
// the aggregator only receives logs and aggregates them. Thats why its a receive-only channel
func (la *LogAggregator) Register(logChan <-chan LogEntry) {
	// read goroutines: each Register creates a specific channel
	la.wg.Add(1)
	go func(logEntry <-chan LogEntry) {
		defer la.wg.Done()

		for log := range logEntry {
			la.bridge <- log
		}
	}(logChan)
}

// start inits the log processing
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
