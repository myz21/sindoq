package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewBus(t *testing.T) {
	bus := NewBus()
	if bus == nil {
		t.Fatal("NewBus() returned nil")
	}
}

func TestBus_Subscribe(t *testing.T) {
	bus := NewBus()

	var received *Event
	unsubscribe := bus.Subscribe(EventOutputStdout, func(e *Event) {
		received = e
	})

	if unsubscribe == nil {
		t.Fatal("Subscribe() returned nil unsubscribe function")
	}

	// Emit an event synchronously
	bus.EmitSync(&Event{
		Type: EventOutputStdout,
		Data: "test output",
	})

	if received == nil {
		t.Fatal("Event was not received")
	}
	if received.Data != "test output" {
		t.Errorf("Expected data 'test output', got %v", received.Data)
	}

	// Unsubscribe
	unsubscribe()

	// Emit again - should not be received
	received = nil
	bus.EmitSync(&Event{
		Type: EventOutputStdout,
		Data: "after unsubscribe",
	})

	if received != nil {
		t.Error("Should not receive events after unsubscribe")
	}
}

func TestBus_SubscribeAll(t *testing.T) {
	bus := NewBus()

	var receivedEvents []*Event
	var mu sync.Mutex

	unsubscribe := bus.SubscribeAll(func(e *Event) {
		mu.Lock()
		receivedEvents = append(receivedEvents, e)
		mu.Unlock()
	})

	// Emit different event types synchronously
	bus.EmitSync(&Event{Type: EventOutputStdout, Data: "stdout"})
	bus.EmitSync(&Event{Type: EventOutputStderr, Data: "stderr"})
	bus.EmitSync(&Event{Type: EventExecutionComplete})

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count != 3 {
		t.Errorf("Expected 3 events, got %d", count)
	}

	unsubscribe()
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := NewBus()

	var count1, count2 int

	unsub1 := bus.Subscribe(EventOutputStdout, func(e *Event) {
		count1++
	})
	unsub2 := bus.Subscribe(EventOutputStdout, func(e *Event) {
		count2++
	})

	bus.EmitSync(&Event{Type: EventOutputStdout})

	if count1 != 1 || count2 != 1 {
		t.Errorf("Both subscribers should receive event: count1=%d, count2=%d", count1, count2)
	}

	unsub1()
	unsub2()
}

func TestBus_EventFiltering(t *testing.T) {
	bus := NewBus()

	var stdoutCount, stderrCount int

	bus.Subscribe(EventOutputStdout, func(e *Event) {
		stdoutCount++
	})
	bus.Subscribe(EventOutputStderr, func(e *Event) {
		stderrCount++
	})

	// Emit stdout event
	bus.EmitSync(&Event{Type: EventOutputStdout})

	if stdoutCount != 1 {
		t.Errorf("Stdout subscriber should receive 1 event, got %d", stdoutCount)
	}
	if stderrCount != 0 {
		t.Errorf("Stderr subscriber should receive 0 events, got %d", stderrCount)
	}
}

func TestBus_Clear(t *testing.T) {
	bus := NewBus()

	var received bool
	bus.Subscribe(EventOutputStdout, func(e *Event) {
		received = true
	})

	bus.Clear()

	// Emit after clear
	bus.EmitSync(&Event{Type: EventOutputStdout})

	if received {
		t.Error("Should not receive events after Clear")
	}
}

func TestBus_SubscriberCount(t *testing.T) {
	bus := NewBus()

	if bus.SubscriberCount() != 0 {
		t.Error("Initial subscriber count should be 0")
	}

	unsub1 := bus.Subscribe(EventOutputStdout, func(e *Event) {})
	if bus.SubscriberCount() != 1 {
		t.Error("Subscriber count should be 1")
	}

	unsub2 := bus.SubscribeAll(func(e *Event) {})
	if bus.SubscriberCount() != 2 {
		t.Error("Subscriber count should be 2")
	}

	unsub1()
	if bus.SubscriberCount() != 1 {
		t.Error("Subscriber count should be 1 after unsub")
	}

	unsub2()
	if bus.SubscriberCount() != 0 {
		t.Error("Subscriber count should be 0 after all unsubs")
	}
}

func TestBus_Emit_Async(t *testing.T) {
	bus := NewBus()

	var received atomic.Bool
	bus.Subscribe(EventOutputStdout, func(e *Event) {
		received.Store(true)
	})

	// Emit asynchronously
	bus.Emit(&Event{Type: EventOutputStdout})

	// Give async publish time to complete
	time.Sleep(50 * time.Millisecond)

	if !received.Load() {
		t.Error("Async emit should deliver event")
	}
}

func TestBus_SubscribeMultiple(t *testing.T) {
	bus := NewBus()

	var count atomic.Int32
	unsub := bus.SubscribeMultiple([]EventType{EventOutputStdout, EventOutputStderr}, func(e *Event) {
		count.Add(1)
	})

	bus.EmitSync(&Event{Type: EventOutputStdout})
	bus.EmitSync(&Event{Type: EventOutputStderr})
	bus.EmitSync(&Event{Type: EventExecutionComplete}) // Should not be received

	if count.Load() != 2 {
		t.Errorf("Expected 2 events, got %d", count.Load())
	}

	unsub()
}

func TestBus_ConcurrentEmit(t *testing.T) {
	bus := NewBus()

	var count atomic.Int32
	bus.Subscribe(EventOutputStdout, func(e *Event) {
		count.Add(1)
	})

	// Emit many events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Emit(&Event{Type: EventOutputStdout})
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 100 {
		t.Errorf("Expected 100 events, got %d", count.Load())
	}
}

func TestNewEvent(t *testing.T) {
	e := NewEvent(EventOutputStdout, "sandbox-123", "test data")

	if e.Type != EventOutputStdout {
		t.Errorf("Type = %v, want %v", e.Type, EventOutputStdout)
	}
	if e.SandboxID != "sandbox-123" {
		t.Errorf("SandboxID = %v, want %v", e.SandboxID, "sandbox-123")
	}
	if e.Data != "test data" {
		t.Errorf("Data = %v, want %v", e.Data, "test data")
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestNewErrorEvent(t *testing.T) {
	err := &testError{msg: "test error"}
	e := NewErrorEvent(EventExecutionError, "sandbox-123", err)

	if e.Type != EventExecutionError {
		t.Errorf("Type = %v, want %v", e.Type, EventExecutionError)
	}
	if e.Error != err {
		t.Errorf("Error = %v, want %v", e.Error, err)
	}
}

func TestEvent_WithMetadata(t *testing.T) {
	e := NewEvent(EventOutputStdout, "sandbox-123", nil)
	e.WithMetadata("key1", "value1").WithMetadata("key2", 42)

	if e.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want %v", e.Metadata["key1"], "value1")
	}
	if e.Metadata["key2"] != 42 {
		t.Errorf("Metadata[key2] = %v, want %v", e.Metadata["key2"], 42)
	}
}

func TestEventTypes(t *testing.T) {
	types := []EventType{
		EventSandboxCreated,
		EventSandboxStarted,
		EventSandboxStopped,
		EventSandboxError,
		EventExecutionStarted,
		EventExecutionComplete,
		EventExecutionError,
		EventExecutionTimeout,
		EventOutputStdout,
		EventOutputStderr,
		EventFileWritten,
		EventFileRead,
		EventFileDeleted,
		EventFileUploaded,
		EventPortPublished,
		EventPortUnpublished,
	}

	for _, typ := range types {
		if typ == "" {
			t.Error("Event type should not be empty")
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
