package main

import (
	"testing"
	"time"
)

func TestLogAggregator_SingleProducer(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	logsChan := make(chan LogEntry, 10)

	// Register and send logs
	agg.Register(logsChan)

	go func() {
		for i := 0; i < 5; i++ {
			logsChan <- LogEntry{
				Timestamp: time.Now(),
				Source:    "api",
				Level:     "INFO",
				Message:   "test message",
			}
		}
		close(logsChan)
	}()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	logs := agg.Stop()

	if len(logs) != 5 {
		t.Errorf("expected 5 logs, got %d", len(logs))
	}
}

func TestLogAggregator_MultipleProducers(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	producers := []string{"api", "database", "cache"}
	logsPerProducer := 10

	for _, name := range producers {
		ch := make(chan LogEntry, logsPerProducer)
		agg.Register(ch)

		go func(source string, logCh chan LogEntry) {
			for i := 0; i < logsPerProducer; i++ {
				logCh <- LogEntry{
					Timestamp: time.Now(),
					Source:    source,
					Level:     "INFO",
					Message:   "concurrent log",
				}
				time.Sleep(time.Millisecond) // Simulate work
			}
			close(logCh)
		}(name, ch)
	}

	// Wait for all producers to finish
	time.Sleep(200 * time.Millisecond)

	logs := agg.Stop()

	expectedTotal := len(producers) * logsPerProducer
	if len(logs) != expectedTotal {
		t.Errorf("expected %d logs, got %d", expectedTotal, len(logs))
	}

	// Verify all sources are present
	sources := make(map[string]int)
	for _, log := range logs {
		sources[log.Source]++
	}

	for _, name := range producers {
		if sources[name] != logsPerProducer {
			t.Errorf("expected %d logs from %s, got %d",
				logsPerProducer, name, sources[name])
		}
	}
}

func TestLogAggregator_GracefulShutdown(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	// Fast producer
	fastCh := make(chan LogEntry, 5)
	agg.Register(fastCh)

	// Slow producer
	slowCh := make(chan LogEntry, 5)
	agg.Register(slowCh)

	// Fast producer finishes immediately
	go func() {
		for i := 0; i < 3; i++ {
			fastCh <- LogEntry{
				Timestamp: time.Now(),
				Source:    "fast",
				Level:     "INFO",
				Message:   "quick",
			}
		}
		close(fastCh)
	}()

	// Slow producer takes time
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(50 * time.Millisecond)
			slowCh <- LogEntry{
				Timestamp: time.Now(),
				Source:    "slow",
				Level:     "INFO",
				Message:   "delayed",
			}
		}
		close(slowCh)
	}()

	// Should wait for BOTH producers to finish
	logs := agg.Stop()

	if len(logs) != 6 {
		t.Errorf("graceful shutdown failed: expected 6 logs, got %d", len(logs))
	}

	fastCount := 0
	slowCount := 0
	for _, log := range logs {
		if log.Source == "fast" {
			fastCount++
		}
		if log.Source == "slow" {
			slowCount++
		}
	}

	if fastCount != 3 || slowCount != 3 {
		t.Errorf("expected 3 fast and 3 slow logs, got fast=%d slow=%d",
			fastCount, slowCount)
	}
}

func TestLogAggregator_EmptyProducers(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	// Register producers that send nothing
	for i := 0; i < 3; i++ {
		ch := make(chan LogEntry)
		agg.Register(ch)
		close(ch) // Immediately close
	}

	time.Sleep(50 * time.Millisecond)
	logs := agg.Stop()

	if len(logs) != 0 {
		t.Errorf("expected 0 logs from empty producers, got %d", len(logs))
	}
}

func TestLogAggregator_DifferentLogLevels(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	ch := make(chan LogEntry, 10)
	agg.Register(ch)

	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	go func() {
		for _, level := range levels {
			ch <- LogEntry{
				Timestamp: time.Now(),
				Source:    "test",
				Level:     level,
				Message:   "message",
			}
		}
		close(ch)
	}()

	time.Sleep(50 * time.Millisecond)
	logs := agg.Stop()

	if len(logs) != len(levels) {
		t.Errorf("expected %d logs, got %d", len(levels), len(logs))
	}

	// Verify all levels are present
	levelCount := make(map[string]int)
	for _, log := range logs {
		levelCount[log.Level]++
	}

	for _, level := range levels {
		if levelCount[level] != 1 {
			t.Errorf("expected 1 log with level %s, got %d", level, levelCount[level])
		}
	}
}

// This test verifies no data races occur (run with -race flag)
func TestLogAggregator_ConcurrencyStress(t *testing.T) {
	agg := NewLogAggregator()
	agg.Start()

	numProducers := 20
	logsPerProducer := 50

	for i := 0; i < numProducers; i++ {
		ch := make(chan LogEntry, logsPerProducer)
		agg.Register(ch)

		go func(id int, logCh chan LogEntry) {
			for j := 0; j < logsPerProducer; j++ {
				logCh <- LogEntry{
					Timestamp: time.Now(),
					Source:    "stress",
					Level:     "INFO",
					Message:   "concurrent stress test",
				}
			}
			close(logCh)
		}(i, ch)
	}

	logs := agg.Stop()

	expectedTotal := numProducers * logsPerProducer
	if len(logs) != expectedTotal {
		t.Errorf("stress test: expected %d logs, got %d", expectedTotal, len(logs))
	}
}
