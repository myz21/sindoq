package testutil

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/happyhackingspace/sindoq/pkg/fs"
)

// MockFileSystem is a configurable mock implementation of fs.FileSystem.
type MockFileSystem struct {
	files map[string][]byte
	mu    sync.RWMutex

	// Hooks for testing
	OnRead   func(ctx context.Context, path string) ([]byte, error)
	OnWrite  func(ctx context.Context, path string, data []byte) error
	OnDelete func(ctx context.Context, path string) error
	OnList   func(ctx context.Context, path string) ([]fs.FileInfo, error)
	OnExists func(ctx context.Context, path string) (bool, error)
}

// NewMockFileSystem creates a new mock filesystem.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

// Read reads a file from the mock filesystem.
func (f *MockFileSystem) Read(ctx context.Context, path string) ([]byte, error) {
	if f.OnRead != nil {
		return f.OnRead(ctx, path)
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

// Write writes a file to the mock filesystem.
func (f *MockFileSystem) Write(ctx context.Context, path string, data []byte) error {
	if f.OnWrite != nil {
		return f.OnWrite(ctx, path, data)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.files[path] = data
	return nil
}

// Delete removes a file from the mock filesystem.
func (f *MockFileSystem) Delete(ctx context.Context, path string) error {
	if f.OnDelete != nil {
		return f.OnDelete(ctx, path)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.files, path)
	return nil
}

// List lists files in the mock filesystem.
func (f *MockFileSystem) List(ctx context.Context, path string) ([]fs.FileInfo, error) {
	if f.OnList != nil {
		return f.OnList(ctx, path)
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	var result []fs.FileInfo
	for name, data := range f.files {
		result = append(result, fs.FileInfo{
			Name:    name,
			Size:    int64(len(data)),
			IsDir:   false,
			ModTime: time.Now(),
		})
	}
	return result, nil
}

// Exists checks if a file exists.
func (f *MockFileSystem) Exists(ctx context.Context, path string) (bool, error) {
	if f.OnExists != nil {
		return f.OnExists(ctx, path)
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	_, ok := f.files[path]
	return ok, nil
}

// Stat returns file info.
func (f *MockFileSystem) Stat(ctx context.Context, path string) (*fs.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return &fs.FileInfo{
		Name:    path,
		Size:    int64(len(data)),
		IsDir:   false,
		ModTime: time.Now(),
	}, nil
}

// Upload uploads a file (same as Write for mock).
func (f *MockFileSystem) Upload(ctx context.Context, localPath, remotePath string) error {
	return fmt.Errorf("upload not supported in mock - use Write instead")
}

// UploadReader uploads content from a reader.
func (f *MockFileSystem) UploadReader(ctx context.Context, reader io.Reader, remotePath string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return f.Write(ctx, remotePath, data)
}

// Download downloads a file to a writer.
func (f *MockFileSystem) Download(ctx context.Context, remotePath string, writer io.Writer) error {
	data, err := f.Read(ctx, remotePath)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// MkDir creates a directory (no-op in simple mock).
func (f *MockFileSystem) MkDir(ctx context.Context, path string) error {
	return nil
}

// Copy copies a file.
func (f *MockFileSystem) Copy(ctx context.Context, src, dst string) error {
	data, err := f.Read(ctx, src)
	if err != nil {
		return err
	}
	return f.Write(ctx, dst, data)
}

// Move moves a file.
func (f *MockFileSystem) Move(ctx context.Context, src, dst string) error {
	if err := f.Copy(ctx, src, dst); err != nil {
		return err
	}
	return f.Delete(ctx, src)
}

// SetFile adds a file to the mock filesystem (for test setup).
func (f *MockFileSystem) SetFile(path string, content []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[path] = content
}

// GetFile retrieves a file from the mock filesystem (for assertions).
func (f *MockFileSystem) GetFile(path string) ([]byte, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	data, ok := f.files[path]
	return data, ok
}

// Clear removes all files from the mock filesystem.
func (f *MockFileSystem) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files = make(map[string][]byte)
}

var _ fs.FileSystem = (*MockFileSystem)(nil)
