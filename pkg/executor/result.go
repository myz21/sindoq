// Package executor provides core execution abstractions for the sindoq SDK.
package executor

import (
	"time"
)

// ExecutionResult contains the outcome of code execution.
type ExecutionResult struct {
	// ExitCode is the process exit code (0 = success).
	ExitCode int

	// Stdout contains standard output.
	Stdout string

	// Stderr contains standard error output.
	Stderr string

	// Duration is the execution time.
	Duration time.Duration

	// Language is the detected programming language.
	Language string

	// Artifacts contains any generated files or outputs.
	Artifacts []Artifact

	// Error contains any execution error.
	Error error

	// Metadata contains provider-specific additional data.
	Metadata map[string]any
}

// Success returns true if the execution completed successfully.
func (r *ExecutionResult) Success() bool {
	return r.ExitCode == 0 && r.Error == nil
}

// CommandResult contains the outcome of command execution.
type CommandResult struct {
	// ExitCode is the process exit code (0 = success).
	ExitCode int

	// Stdout contains standard output.
	Stdout string

	// Stderr contains standard error output.
	Stderr string

	// Duration is the execution time.
	Duration time.Duration
}

// Success returns true if the command completed successfully.
func (r *CommandResult) Success() bool {
	return r.ExitCode == 0
}

// Artifact represents a generated file or output.
type Artifact struct {
	// Name is the artifact name.
	Name string

	// Path is the path within the sandbox.
	Path string

	// MIMEType is the content type.
	MIMEType string

	// Size is the file size in bytes.
	Size int64

	// Data contains the artifact content (may be nil for large files).
	Data []byte
}

// ExecutionOptions configures execution behavior.
type ExecutionOptions struct {
	// Language overrides auto-detection (optional).
	Language string

	// Filename provides a hint for language detection.
	Filename string

	// Timeout for execution (default: 30s).
	Timeout time.Duration

	// Env provides environment variables.
	Env map[string]string

	// WorkDir sets the working directory.
	WorkDir string

	// Stdin provides input to the program.
	Stdin string

	// Files to create before execution.
	Files map[string][]byte

	// KeepArtifacts preserves generated files after execution.
	KeepArtifacts bool
}

// DefaultExecutionOptions returns sensible defaults.
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		Timeout: 30 * time.Second,
		WorkDir: "/workspace",
		Env:     make(map[string]string),
		Files:   make(map[string][]byte),
	}
}

// Merge combines options with defaults for unset values.
func (o *ExecutionOptions) Merge(defaults *ExecutionOptions) *ExecutionOptions {
	if o == nil {
		return defaults
	}

	result := *o

	if result.Timeout == 0 {
		result.Timeout = defaults.Timeout
	}
	if result.WorkDir == "" {
		result.WorkDir = defaults.WorkDir
	}
	if result.Env == nil {
		result.Env = defaults.Env
	}
	if result.Files == nil {
		result.Files = defaults.Files
	}

	return &result
}
