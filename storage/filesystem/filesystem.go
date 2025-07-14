// Package filesystem provides file system storage backend for HybridBuffer
package filesystem

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"schneider.vip/hybridbuffer/storage"
)

// Backend implements StorageBackend for local file system
type Backend struct {
	filename string
	tempDir  string
	prefix   string
}

// Option configures file storage backend
type Option func(*Backend)

// WithTempDir sets the temporary directory for file storage
func WithTempDir(tempDir string) Option {
	return func(fs *Backend) {
		fs.tempDir = tempDir
	}
}

// WithPrefix sets the file prefix for temporary files
func WithPrefix(prefix string) Option {
	return func(fs *Backend) {
		fs.prefix = prefix
	}
}

// newBackend creates a new file-based storage backend with options
func newBackend(opts ...Option) (*Backend, error) {
	fs := &Backend{
		prefix: "hybridbuffer", // default prefix
	}

	// Apply options
	for _, opt := range opts {
		opt(fs)
	}

	return fs, nil
}

// Create implements StorageBackend
func (fs *Backend) Create() (io.WriteCloser, error) {
	// Generate temporary file with custom prefix and .tmp extension
	pattern := fs.prefix + "-*.tmp"
	file, err := os.CreateTemp(fs.tempDir, pattern)
	if err != nil {
		return nil, err
	}

	fs.filename = file.Name()
	return file, nil
}

// Open implements StorageBackend
func (fs *Backend) Open() (io.ReadCloser, error) {
	if fs.filename == "" {
		return nil, errors.New("no file created yet")
	}
	return os.Open(fs.filename)
}

// Remove implements StorageBackend
func (fs *Backend) Remove() error {
	if fs.filename == "" {
		return nil
	}
	return os.Remove(fs.filename)
}

// New creates a new file storage backend provider function
func New(opts ...Option) func() storage.Backend {
	return func() storage.Backend {
		backend, _ := newBackend(opts...)
		return backend
	}
}
