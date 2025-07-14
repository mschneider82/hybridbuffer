package storage

import "io"

// Backend defines the minimal interface for external storage backends
type Backend interface {
	// Create creates a new storage location and returns a writer
	Create() (io.WriteCloser, error)

	// Open opens an existing storage location for reading
	Open() (io.ReadCloser, error)

	// Remove removes the storage location
	Remove() error
}