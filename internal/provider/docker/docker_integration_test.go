//go:build integration

package docker

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/happyhackingspace/sindoq/internal/provider"
	"github.com/happyhackingspace/sindoq/pkg/executor"
)

func TestDockerProviderIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p, err := New(nil)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer p.Close()

	t.Run("validate", func(t *testing.T) {
		if err := p.Validate(ctx); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("capabilities", func(t *testing.T) {
		caps := p.Capabilities()
		if !caps.SupportsStreaming {
			t.Error("SupportsStreaming should be true")
		}
		if len(caps.SupportedLanguages) == 0 {
			t.Error("SupportedLanguages should not be empty")
		}
	})

	t.Run("execute python", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		result, err := instance.Execute(ctx, `print("Hello from Python!")`, &executor.ExecutionOptions{
			Language: "Python",
			Timeout:  30 * time.Second,
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0\nStderr: %s", result.ExitCode, result.Stderr)
		}
		if !strings.Contains(result.Stdout, "Hello from Python!") {
			t.Errorf("Stdout = %q, want to contain %q", result.Stdout, "Hello from Python!")
		}
	})

	t.Run("execute javascript", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "JavaScript",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		result, err := instance.Execute(ctx, `console.log("Hello from JavaScript!")`, &executor.ExecutionOptions{
			Language: "JavaScript",
			Timeout:  30 * time.Second,
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0\nStderr: %s", result.ExitCode, result.Stderr)
		}
		if !strings.Contains(result.Stdout, "Hello from JavaScript!") {
			t.Errorf("Stdout = %q, want to contain %q", result.Stdout, "Hello from JavaScript!")
		}
	})

	t.Run("execute with error", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		result, err := instance.Execute(ctx, `raise Exception("test error")`, &executor.ExecutionOptions{
			Language: "Python",
			Timeout:  30 * time.Second,
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.ExitCode == 0 {
			t.Error("ExitCode should be non-zero for error")
		}
		if result.Stderr == "" {
			t.Error("Stderr should contain error message")
		}
	})

	t.Run("execute stream", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		var output strings.Builder
		err = instance.ExecuteStream(ctx, `
for i in range(3):
    print(f"Line {i}")
`, &executor.ExecutionOptions{
			Language: "Python",
			Timeout:  30 * time.Second,
		}, func(e *executor.StreamEvent) error {
			if e.Type == executor.StreamStdout {
				output.WriteString(e.Data)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("ExecuteStream() error = %v", err)
		}

		out := output.String()
		if !strings.Contains(out, "Line 0") || !strings.Contains(out, "Line 1") || !strings.Contains(out, "Line 2") {
			t.Errorf("output = %q, want lines 0-2", out)
		}
	})

	t.Run("run command", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		result, err := instance.RunCommand(ctx, "echo", []string{"hello"})
		if err != nil {
			t.Fatalf("RunCommand() error = %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
		if !strings.Contains(result.Stdout, "hello") {
			t.Errorf("Stdout = %q, want to contain %q", result.Stdout, "hello")
		}
	})

	t.Run("status", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		status, err := instance.Status(ctx)
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if status != provider.StatusRunning {
			t.Errorf("Status() = %v, want %v", status, provider.StatusRunning)
		}

		instance.Stop(ctx)

		status, err = instance.Status(ctx)
		if err != nil {
			t.Fatalf("Status() after stop error = %v", err)
		}
		if status != provider.StatusStopped {
			t.Errorf("Status() after stop = %v, want %v", status, provider.StatusStopped)
		}
	})

	t.Run("filesystem", func(t *testing.T) {
		instance, err := p.Create(ctx, &provider.CreateOptions{
			Runtime: "Python",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		defer instance.Stop(ctx)

		fs := instance.FileSystem()
		if fs == nil {
			t.Skip("FileSystem not implemented")
		}

		// Write file to /tmp which exists in all containers
		err = fs.Write(ctx, "/tmp/test.txt", []byte("hello world"))
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Read file
		data, err := fs.Read(ctx, "/tmp/test.txt")
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if string(data) != "hello world" {
			t.Errorf("Read() = %q, want %q", data, "hello world")
		}

		// Check exists
		exists, err := fs.Exists(ctx, "/tmp/test.txt")
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("file should exist")
		}

		// Delete
		err = fs.Delete(ctx, "/tmp/test.txt")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		exists, err = fs.Exists(ctx, "/tmp/test.txt")
		if err != nil {
			t.Fatalf("Exists() after delete error = %v", err)
		}
		if exists {
			t.Error("file should not exist after delete")
		}
	})
}

func TestDockerProviderTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p, err := New(nil)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer p.Close()

	instance, err := p.Create(ctx, &provider.CreateOptions{
		Runtime: "Python",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer instance.Stop(ctx)

	// Execute with very short timeout
	_, err = instance.Execute(ctx, `
import time
time.sleep(60)
`, &executor.ExecutionOptions{
		Language: "Python",
		Timeout:  1 * time.Second,
	})

	if err == nil {
		t.Error("Execute() should fail with timeout")
	}
}

func TestDockerProviderWithEnv(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p, err := New(nil)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer p.Close()

	instance, err := p.Create(ctx, &provider.CreateOptions{
		Runtime: "Python",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer instance.Stop(ctx)

	result, err := instance.Execute(ctx, `
import os
print(os.environ.get("MY_VAR", "not set"))
`, &executor.ExecutionOptions{
		Language: "Python",
		Timeout:  30 * time.Second,
		Env: map[string]string{
			"MY_VAR": "test_value",
		},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Stdout, "test_value") {
		t.Errorf("Stdout = %q, want to contain %q", result.Stdout, "test_value")
	}
}

func TestDockerProviderWithStdin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p, err := New(nil)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer p.Close()

	instance, err := p.Create(ctx, &provider.CreateOptions{
		Runtime: "Python",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer instance.Stop(ctx)

	result, err := instance.Execute(ctx, `
import sys
data = sys.stdin.read()
print(f"Got: {data}")
`, &executor.ExecutionOptions{
		Language: "Python",
		Timeout:  30 * time.Second,
		Stdin:    "hello from stdin",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Stdout, "hello from stdin") {
		t.Errorf("Stdout = %q, want to contain stdin data", result.Stdout)
	}
}
