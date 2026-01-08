package executor

import (
	"errors"
	"testing"
	"time"
)

func TestExecutionResultSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   ExecutionResult
		expected bool
	}{
		{
			name:     "success with zero exit code",
			result:   ExecutionResult{ExitCode: 0, Error: nil},
			expected: true,
		},
		{
			name:     "failure with non-zero exit code",
			result:   ExecutionResult{ExitCode: 1, Error: nil},
			expected: false,
		},
		{
			name:     "failure with error",
			result:   ExecutionResult{ExitCode: 0, Error: errors.New("exec failed")},
			expected: false,
		},
		{
			name:     "failure with both non-zero and error",
			result:   ExecutionResult{ExitCode: 1, Error: errors.New("exec failed")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExecutionResultFields(t *testing.T) {
	result := ExecutionResult{
		ExitCode:  0,
		Stdout:    "output",
		Stderr:    "errors",
		Duration:  500 * time.Millisecond,
		Language:  "Python",
		Artifacts: []Artifact{{Name: "output.txt", Path: "/tmp/output.txt"}},
		Metadata:  map[string]any{"key": "value"},
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "output" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "output")
	}
	if result.Stderr != "errors" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "errors")
	}
	if result.Duration != 500*time.Millisecond {
		t.Errorf("Duration = %v, want %v", result.Duration, 500*time.Millisecond)
	}
	if result.Language != "Python" {
		t.Errorf("Language = %q, want %q", result.Language, "Python")
	}
	if len(result.Artifacts) != 1 {
		t.Errorf("Artifacts length = %d, want 1", len(result.Artifacts))
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want %q", result.Metadata["key"], "value")
	}
}

func TestCommandResultSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   CommandResult
		expected bool
	}{
		{
			name:     "success",
			result:   CommandResult{ExitCode: 0},
			expected: true,
		},
		{
			name:     "failure",
			result:   CommandResult{ExitCode: 127},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArtifact(t *testing.T) {
	artifact := Artifact{
		Name:     "report.pdf",
		Path:     "/workspace/report.pdf",
		MIMEType: "application/pdf",
		Size:     1024,
		Data:     []byte("PDF content"),
	}

	if artifact.Name != "report.pdf" {
		t.Errorf("Name = %q, want %q", artifact.Name, "report.pdf")
	}
	if artifact.Path != "/workspace/report.pdf" {
		t.Errorf("Path = %q, want %q", artifact.Path, "/workspace/report.pdf")
	}
	if artifact.MIMEType != "application/pdf" {
		t.Errorf("MIMEType = %q, want %q", artifact.MIMEType, "application/pdf")
	}
	if artifact.Size != 1024 {
		t.Errorf("Size = %d, want 1024", artifact.Size)
	}
	if string(artifact.Data) != "PDF content" {
		t.Errorf("Data = %q, want %q", artifact.Data, "PDF content")
	}
}

func TestDefaultExecutionOptions(t *testing.T) {
	opts := DefaultExecutionOptions()

	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}
	if opts.WorkDir != "/workspace" {
		t.Errorf("WorkDir = %q, want %q", opts.WorkDir, "/workspace")
	}
	if opts.Env == nil {
		t.Error("Env should not be nil")
	}
	if opts.Files == nil {
		t.Error("Files should not be nil")
	}
}

func TestExecutionOptionsMerge(t *testing.T) {
	defaults := DefaultExecutionOptions()

	t.Run("nil options returns defaults", func(t *testing.T) {
		var opts *ExecutionOptions
		result := opts.Merge(defaults)

		if result != defaults {
			t.Error("Merge(nil) should return defaults")
		}
	})

	t.Run("merge fills unset values", func(t *testing.T) {
		opts := &ExecutionOptions{
			Language: "Python",
		}
		result := opts.Merge(defaults)

		if result.Language != "Python" {
			t.Errorf("Language = %q, want %q", result.Language, "Python")
		}
		if result.Timeout != defaults.Timeout {
			t.Errorf("Timeout = %v, want %v", result.Timeout, defaults.Timeout)
		}
		if result.WorkDir != defaults.WorkDir {
			t.Errorf("WorkDir = %q, want %q", result.WorkDir, defaults.WorkDir)
		}
	})

	t.Run("merge preserves set values", func(t *testing.T) {
		opts := &ExecutionOptions{
			Language: "Go",
			Timeout:  1 * time.Minute,
			WorkDir:  "/app",
			Env:      map[string]string{"FOO": "bar"},
			Files:    map[string][]byte{"test.txt": []byte("test")},
		}
		result := opts.Merge(defaults)

		if result.Timeout != 1*time.Minute {
			t.Errorf("Timeout = %v, want %v", result.Timeout, 1*time.Minute)
		}
		if result.WorkDir != "/app" {
			t.Errorf("WorkDir = %q, want %q", result.WorkDir, "/app")
		}
		if result.Env["FOO"] != "bar" {
			t.Errorf("Env[FOO] = %q, want %q", result.Env["FOO"], "bar")
		}
	})
}

func TestExecutionOptionsFields(t *testing.T) {
	opts := &ExecutionOptions{
		Language:      "Python",
		Filename:      "main.py",
		Timeout:       1 * time.Minute,
		Env:           map[string]string{"DEBUG": "true"},
		WorkDir:       "/app",
		Stdin:         "input data",
		Files:         map[string][]byte{"config.json": []byte("{}")},
		KeepArtifacts: true,
	}

	if opts.Language != "Python" {
		t.Errorf("Language = %q, want %q", opts.Language, "Python")
	}
	if opts.Filename != "main.py" {
		t.Errorf("Filename = %q, want %q", opts.Filename, "main.py")
	}
	if opts.Timeout != 1*time.Minute {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 1*time.Minute)
	}
	if opts.Env["DEBUG"] != "true" {
		t.Errorf("Env[DEBUG] = %q, want %q", opts.Env["DEBUG"], "true")
	}
	if opts.WorkDir != "/app" {
		t.Errorf("WorkDir = %q, want %q", opts.WorkDir, "/app")
	}
	if opts.Stdin != "input data" {
		t.Errorf("Stdin = %q, want %q", opts.Stdin, "input data")
	}
	if string(opts.Files["config.json"]) != "{}" {
		t.Errorf("Files[config.json] = %q, want %q", opts.Files["config.json"], "{}")
	}
	if !opts.KeepArtifacts {
		t.Error("KeepArtifacts should be true")
	}
}
