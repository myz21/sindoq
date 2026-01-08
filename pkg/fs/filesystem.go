// Package fs provides file system abstraction for sandbox environments.
package fs

import (
	"context"
	"io"
	"time"
)

// FileSystem provides file operations within a sandbox.
type FileSystem interface {
	// Read reads file contents.
	Read(ctx context.Context, path string) ([]byte, error)

	// Write writes data to a file.
	Write(ctx context.Context, path string, data []byte) error

	// Delete removes a file or directory.
	Delete(ctx context.Context, path string) error

	// List lists files in a directory.
	List(ctx context.Context, path string) ([]FileInfo, error)

	// Exists checks if a path exists.
	Exists(ctx context.Context, path string) (bool, error)

	// Stat returns file information.
	Stat(ctx context.Context, path string) (*FileInfo, error)

	// Upload uploads a local file to the sandbox.
	Upload(ctx context.Context, localPath, remotePath string) error

	// UploadReader uploads content from a reader to the sandbox.
	UploadReader(ctx context.Context, reader io.Reader, remotePath string) error

	// Download downloads a file from the sandbox.
	Download(ctx context.Context, remotePath string, writer io.Writer) error

	// MkDir creates a directory (including parents).
	MkDir(ctx context.Context, path string) error

	// Copy copies a file within the sandbox.
	Copy(ctx context.Context, src, dst string) error

	// Move moves/renames a file within the sandbox.
	Move(ctx context.Context, src, dst string) error
}

// FileInfo contains file metadata.
type FileInfo struct {
	// Name is the base name of the file.
	Name string

	// Path is the full path within the sandbox.
	Path string

	// Size is the file size in bytes.
	Size int64

	// IsDir indicates if this is a directory.
	IsDir bool

	// ModTime is the modification time.
	ModTime time.Time

	// Mode is the file mode/permissions.
	Mode uint32

	// MIMEType is the detected content type (optional).
	MIMEType string
}

// FileReader provides streaming read access to a file.
type FileReader interface {
	io.ReadCloser
	FileInfo() *FileInfo
}

// FileWriter provides streaming write access to a file.
type FileWriter interface {
	io.WriteCloser
	Path() string
}

// WatchEvent represents a file system change event.
type WatchEvent struct {
	// Type is the event type (create, modify, delete, rename).
	Type WatchEventType

	// Path is the affected file path.
	Path string

	// OldPath is set for rename events.
	OldPath string

	// Timestamp when the event occurred.
	Timestamp time.Time
}

// WatchEventType categorizes watch events.
type WatchEventType string

const (
	WatchCreate WatchEventType = "create"
	WatchModify WatchEventType = "modify"
	WatchDelete WatchEventType = "delete"
	WatchRename WatchEventType = "rename"
)

// Watcher provides file system watching capabilities.
type Watcher interface {
	// Watch starts watching a path for changes.
	// Returns a channel that receives events and an unwatch function.
	Watch(ctx context.Context, path string) (<-chan *WatchEvent, func(), error)
}

// WatchableFileSystem extends FileSystem with watching capabilities.
type WatchableFileSystem interface {
	FileSystem
	Watcher
}
