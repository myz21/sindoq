//go:build linux

package gvisor

import (
	"archive/tar"
	"io"
	"path/filepath"
	"time"
)

// tarWriter wraps tar.Writer with helper methods.
type tarWriter struct {
	tw *tar.Writer
}

// newTarWriter creates a new tarWriter.
func newTarWriter(w io.Writer) *tarWriter {
	return &tarWriter{
		tw: tar.NewWriter(w),
	}
}

// WriteFile adds a file to the tar archive.
func (t *tarWriter) WriteFile(name string, content []byte) error {
	filename := filepath.Base(name)

	header := &tar.Header{
		Name:    filename,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	if err := t.tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := t.tw.Write(content)
	return err
}

// WriteDir adds a directory entry to the tar archive.
func (t *tarWriter) WriteDir(name string) error {
	header := &tar.Header{
		Name:     name + "/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
		ModTime:  time.Now(),
	}

	return t.tw.WriteHeader(header)
}

// Close closes the tar writer.
func (t *tarWriter) Close() error {
	return t.tw.Close()
}
