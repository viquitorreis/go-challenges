package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
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

// LogEntry represents a log for each service
type LogEntry struct {
	Timestamp time.Time
	Source    string // "api", "database", "cache", etc
	Level     string // "DEBUG", "INFO", "WARN", "ERROR"
	Message   string
}

// LogAggregator is the heat of this Fan-in pattern
type LogAggregator struct {
	// TODO: add all necessary fields to:
	// - Store channels for producers
	// - Coordinate shutdown for all producers
	// - Collect processed logs
	// - You need a sync mechanism here, which?
}

func NewLogAggregator() *LogAggregator {
	return &LogAggregator{
		// TODO: initialize all fields
	}
}

// Register adds a new producer to the aggregator
// the aggregator only receives logs and aggregates them. Thats why its a receive-only channel
func (la *LogAggregator) Register(logChan <-chan LogEntry) {
	// TODO: store the producer's channel
	// Hint: you will need to read ALL of those channels later on
}

// Start inits the log processing
func (la *LogAggregator) Start() {
	// TODO: start a goroutine that will:
	// 1. Read from ALL channels registered (Fan-In)
	// 2. Process each received log
	// 3. For when all producers finish

	// Hint: How to read from multiple channels at the same time?
	// Hint: how to know when ALL channels were closed? You will probably need a mechanism for that.
}

// Stop para o aggregator e retorna todos os logs coletados
func (la *LogAggregator) Stop() []LogEntry {
	// TODO: implements a graceful shutdown
	// - waits for all producers to finish
	// - return collected logs

	// Hint: você precisa sinalizar que quer parar E esperar
	// que o processamento realmente termine

	// Hint: you will need to signal you want to stop and wait
	//	in order for the processing to actually finish

	return nil // TODO: retorn logs
}
