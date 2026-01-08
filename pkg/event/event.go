// Package event provides the event system for the sindoq SDK using the Observer pattern.
package event

import (
	"time"
)

// EventType categorizes sandbox events.
type EventType string

const (
	// Sandbox lifecycle events
	EventSandboxCreated EventType = "sandbox.created"
	EventSandboxStarted EventType = "sandbox.started"
	EventSandboxStopped EventType = "sandbox.stopped"
	EventSandboxError   EventType = "sandbox.error"

	// Execution events
	EventExecutionStarted  EventType = "execution.started"
	EventExecutionComplete EventType = "execution.complete"
	EventExecutionError    EventType = "execution.error"
	EventExecutionTimeout  EventType = "execution.timeout"

	// Output events
	EventOutputStdout EventType = "output.stdout"
	EventOutputStderr EventType = "output.stderr"

	// File events
	EventFileWritten  EventType = "file.written"
	EventFileRead     EventType = "file.read"
	EventFileDeleted  EventType = "file.deleted"
	EventFileUploaded EventType = "file.uploaded"

	// Network events
	EventPortPublished   EventType = "port.published"
	EventPortUnpublished EventType = "port.unpublished"
)

// Event represents a sandbox event.
type Event struct {
	// Type categorizes the event.
	Type EventType

	// SandboxID is the sandbox that generated the event.
	SandboxID string

	// Timestamp when the event occurred.
	Timestamp time.Time

	// Data contains event-specific payload.
	Data any

	// Error is set for error events.
	Error error

	// Metadata contains additional context.
	Metadata map[string]any
}

// NewEvent creates a new event.
func NewEvent(eventType EventType, sandboxID string, data any) *Event {
	return &Event{
		Type:      eventType,
		SandboxID: sandboxID,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]any),
	}
}

// NewErrorEvent creates an error event.
func NewErrorEvent(eventType EventType, sandboxID string, err error) *Event {
	return &Event{
		Type:      eventType,
		SandboxID: sandboxID,
		Timestamp: time.Now(),
		Error:     err,
		Metadata:  make(map[string]any),
	}
}

// WithMetadata adds metadata to the event and returns it for chaining.
func (e *Event) WithMetadata(key string, value any) *Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// EventHandler processes events.
type EventHandler func(event *Event)

// Emitter publishes events to subscribers.
type Emitter interface {
	// Emit sends an event to all relevant subscribers.
	Emit(event *Event)

	// Subscribe registers a handler for a specific event type.
	// Returns an unsubscribe function.
	Subscribe(eventType EventType, handler EventHandler) func()

	// SubscribeAll registers a handler for all events.
	// Returns an unsubscribe function.
	SubscribeAll(handler EventHandler) func()
}

// ExecutionStartedData contains data for execution.started events.
type ExecutionStartedData struct {
	Language string
	CodeSize int
}

// ExecutionCompleteData contains data for execution.complete events.
type ExecutionCompleteData struct {
	ExitCode int
	Duration time.Duration
	Language string
}

// OutputData contains data for output events.
type OutputData struct {
	Content string
	Line    int
}

// FileEventData contains data for file events.
type FileEventData struct {
	Path string
	Size int64
}

// PortEventData contains data for port events.
type PortEventData struct {
	Port      int
	PublicURL string
}
