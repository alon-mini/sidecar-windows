package event

import (
	"sync"
	"testing"
	"time"
)

func TestDispatcher_SubscribePublish(t *testing.T) {
	d := New()
	defer d.Close()

	ch := d.Subscribe("test")
	e := NewEvent(TypeFileChanged, "test", "data")

	d.Publish("test", e)

	select {
	case received := <-ch:
		if received.Type != e.Type {
			t.Errorf("got type %s, want %s", received.Type, e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestDispatcher_MultipleSubscribers(t *testing.T) {
	d := New()
	defer d.Close()

	ch1 := d.Subscribe("test")
	ch2 := d.Subscribe("test")
	e := NewEvent(TypeTDUpdate, "test", nil)

	d.Publish("test", e)

	for i, ch := range []<-chan Event{ch1, ch2} {
		select {
		case received := <-ch:
			if received.Type != e.Type {
				t.Errorf("sub %d: got type %s, want %s", i, received.Type, e.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("sub %d: timeout waiting for event", i)
		}
	}
}

func TestDispatcher_NonBlocking(t *testing.T) {
	d := New()
	defer d.Close()

	ch := d.Subscribe("test")
	e := NewEvent(TypeError, "test", nil)

	// Fill buffer
	for i := 0; i < defaultBufferSize; i++ {
		d.Publish("test", e)
	}

	// This should not block
	done := make(chan bool)
	go func() {
		d.Publish("test", e) // Should drop
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("Publish blocked with full buffer")
	}

	// Drain and verify we got buffer size events
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			if count != defaultBufferSize {
				t.Errorf("got %d events, want %d", count, defaultBufferSize)
			}
			return
		}
	}
}

func TestDispatcher_Close(t *testing.T) {
	d := New()
	ch := d.Subscribe("test")

	d.Close()

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout - channel not closed")
	}

	// Publish after close should not panic
	d.Publish("test", NewEvent(TypeError, "test", nil))
}

func TestDispatcher_Concurrent(t *testing.T) {
	d := New()

	const numSubscribers = 5
	const numPublishers = 5
	const numEvents = 20

	var pubWg sync.WaitGroup
	pubWg.Add(numPublishers)

	// Subscribe before publishing
	channels := make([]<-chan Event, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		channels[i] = d.Subscribe("concurrent")
	}

	// Publishers
	for i := 0; i < numPublishers; i++ {
		go func() {
			defer pubWg.Done()
			for j := 0; j < numEvents; j++ {
				d.Publish("concurrent", NewEvent(TypeRefreshNeeded, "concurrent", j))
			}
		}()
	}

	// Wait for publishers to finish
	pubWg.Wait()

	// Close dispatcher - this closes all channels
	d.Close()

	// Drain channels - verify no panic and channels are closed
	for _, ch := range channels {
		for range ch {
			// Drain
		}
	}
}
