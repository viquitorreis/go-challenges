package main

import (
	"sync"
	"testing"
)

func TestEventBus(t *testing.T) {
	t.Run("single subscriber receives events", func(t *testing.T) {
		bus := NewEventBus()
		sub := bus.Subscribe("test")

		go bus.Publish("event1")

		received := <-sub
		if received != "event1" {
			t.Errorf("expected 'event1', got '%s'", received)
		}

		bus.Close()
	})

	t.Run("multiple subscribers receive same event", func(t *testing.T) {
		bus := NewEventBus()
		sub1 := bus.Subscribe("sub1")
		sub2 := bus.Subscribe("sub2")

		var wg sync.WaitGroup
		wg.Add(2)

		received1 := ""
		received2 := ""

		go func() {
			received1 = <-sub1
			wg.Done()
		}()

		go func() {
			received2 = <-sub2
			wg.Done()
		}()

		bus.Publish("shared-event")
		wg.Wait()

		if received1 != "shared-event" || received2 != "shared-event" {
			t.Error("both subscribers should receive the event")
		}

		bus.Close()
	})

	t.Run("events are received in order", func(t *testing.T) {
		bus := NewEventBus()
		sub := bus.Subscribe("ordered")

		events := []string{}
		done := make(chan bool)

		go func() {
			for event := range sub {
				events = append(events, event)
				if len(events) == 3 {
					done <- true
				}
			}
		}()

		bus.Publish("first")
		bus.Publish("second")
		bus.Publish("third")

		<-done

		if events[0] != "first" || events[1] != "second" || events[2] != "third" {
			t.Errorf("events out of order: %v", events)
		}

		bus.Close()
	})
}
