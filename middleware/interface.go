package middleware

import "io"

// Middleware wraps Reader and Writer for processing data during storage operations
type Middleware interface {
	// Writer wraps an io.Writer to apply middleware (e.g., encryption, compression)
	Writer(io.Writer) io.Writer
	
	// Reader wraps an io.Reader to reverse middleware (e.g., decryption, decompression)
	Reader(io.Reader) io.Reader
}