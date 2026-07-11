package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func createLogs() {
	f, err := os.Create("access.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	levels := []string{"INFO", "WARN", "ERROR"}
	messages := map[string][]string{
		"INFO":  {"request completed", "user logged in", "cache hit", "health check ok"},
		"WARN":  {"high memory usage", "slow query detected", "retry attempt"},
		"ERROR": {"database timeout", "connection refused", "nil pointer", "request failed"},
	}

	start := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	for i := 0; i < 10000; i++ {
		level := levels[rand.Intn(len(levels))]
		msgs := messages[level]
		msg := msgs[rand.Intn(len(msgs))]
		ts := start.Add(time.Duration(i) * time.Second)
		fmt.Fprintf(w, "%s [%s] %s\n", ts.Format("2006-01-02 15:04:05"), level, msg)
	}
}
