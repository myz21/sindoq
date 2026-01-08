package sindoq

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrSandboxNotFound", ErrSandboxNotFound},
		{"ErrSandboxStopped", ErrSandboxStopped},
		{"ErrExecutionTimeout", ErrExecutionTimeout},
		{"ErrProviderUnavailable", ErrProviderUnavailable},
		{"ErrLanguageNotSupported", ErrLanguageNotSupported},
		{"ErrLanguageDetectionFailed", ErrLanguageDetectionFailed},
		{"ErrResourceExhausted", ErrResourceExhausted},
		{"ErrPermissionDenied", ErrPermissionDenied},
		{"ErrInvalidConfiguration", ErrInvalidConfiguration},
		{"ErrProviderNotRegistered", ErrProviderNotRegistered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("sentinel error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("sentinel error should have a message")
			}
		})
	}
}

func TestSandboxError(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		err := NewError("execute", "docker", "abc123", ErrExecutionTimeout)

		if err.Op != "execute" {
			t.Errorf("Op = %q, want %q", err.Op, "execute")
		}
		if err.Provider != "docker" {
			t.Errorf("Provider = %q, want %q", err.Provider, "docker")
		}
		if err.SandboxID != "abc123" {
			t.Errorf("SandboxID = %q, want %q", err.SandboxID, "abc123")
		}

		msg := err.Error()
		if msg != "execute [docker/abc123]: execution timeout" {
			t.Errorf("Error() = %q, want format with all fields", msg)
		}
	})

	t.Run("without sandbox ID", func(t *testing.T) {
		err := NewError("create", "vercel", "", ErrProviderUnavailable)

		msg := err.Error()
		if msg != "create [vercel]: provider unavailable" {
			t.Errorf("Error() = %q, want format without sandbox ID", msg)
		}
	})

	t.Run("without provider", func(t *testing.T) {
		err := NewError("init", "", "", ErrInvalidConfiguration)

		msg := err.Error()
		if msg != "init: invalid configuration" {
			t.Errorf("Error() = %q, want format without provider", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		underlying := ErrSandboxStopped
		err := NewError("execute", "docker", "abc", underlying)

		if !errors.Is(err, underlying) {
			t.Error("errors.Is should match underlying error")
		}

		unwrapped := err.Unwrap()
		if unwrapped != underlying {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
		}
	})

	t.Run("Is method", func(t *testing.T) {
		err := NewError("execute", "docker", "abc", ErrExecutionTimeout)

		if !err.Is(ErrExecutionTimeout) {
			t.Error("Is() should return true for matching error")
		}
		if err.Is(ErrSandboxStopped) {
			t.Error("Is() should return false for non-matching error")
		}
	})
}

func TestExecutionError(t *testing.T) {
	t.Run("with stderr", func(t *testing.T) {
		err := NewExecutionError(1, "output", "error message", errors.New("exec failed"))

		if err.ExitCode != 1 {
			t.Errorf("ExitCode = %d, want 1", err.ExitCode)
		}
		if err.Stdout != "output" {
			t.Errorf("Stdout = %q, want %q", err.Stdout, "output")
		}
		if err.Stderr != "error message" {
			t.Errorf("Stderr = %q, want %q", err.Stderr, "error message")
		}

		msg := err.Error()
		if msg == "" {
			t.Error("Error() should not be empty")
		}
		if !contains(msg, "exit code 1") {
			t.Errorf("Error() should contain exit code: %s", msg)
		}
		if !contains(msg, "error message") {
			t.Errorf("Error() should contain stderr: %s", msg)
		}
	})

	t.Run("without stderr", func(t *testing.T) {
		err := NewExecutionError(2, "output", "", errors.New("exec failed"))

		msg := err.Error()
		if contains(msg, "stderr") {
			t.Errorf("Error() should not contain stderr label: %s", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewExecutionError(1, "", "", underlying)

		if !errors.Is(err, underlying) {
			t.Error("errors.Is should match underlying error")
		}

		unwrapped := err.Unwrap()
		if unwrapped != underlying {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
