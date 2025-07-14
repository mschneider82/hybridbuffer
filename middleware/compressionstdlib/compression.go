// Package compressionstdlib provides compression middleware for HybridBuffer using stdlib
package compressionstdlib

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/pkg/errors"
	"schneider.vip/hybridbuffer/middleware"
)

// Algorithm represents the compression algorithm to use
type Algorithm int

const (
	// Gzip compression using compress/gzip
	Gzip Algorithm = iota
	// Zlib compression using compress/zlib
	Zlib
)

// Middleware implements compression/decompression
type Middleware struct {
	algorithm Algorithm
	level     int
}

// Ensure Middleware implements middleware.Middleware interface
var _ middleware.Middleware = (*Middleware)(nil)

// Option configures compression middleware
type Option func(*Middleware)

// WithLevel sets the compression level (1-9, where 9 is best compression)
func WithLevel(level int) Option {
	return func(m *Middleware) {
		if level >= 1 && level <= 9 {
			m.level = level
		}
	}
}

// New creates a new compression middleware with the given algorithm
func New(algorithm Algorithm, opts ...Option) *Middleware {
	m := &Middleware{
		algorithm: algorithm,
		level:     6, // Default compression level
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Writer wraps an io.Writer with compression
func (m *Middleware) Writer(w io.Writer) io.Writer {
	switch m.algorithm {
	case Gzip:
		gzipWriter, err := gzip.NewWriterLevel(w, m.level)
		if err != nil {
			panic("failed to create gzip writer: " + err.Error())
		}
		return &gzipWriteCloser{gzipWriter}
	case Zlib:
		zlibWriter, err := zlib.NewWriterLevel(w, m.level)
		if err != nil {
			panic("failed to create zlib writer: " + err.Error())
		}
		return &zlibWriteCloser{zlibWriter}
	default:
		panic("unsupported compression algorithm")
	}
}

// Reader wraps an io.Reader with decompression
func (m *Middleware) Reader(r io.Reader) io.Reader {
	switch m.algorithm {
	case Gzip:
		gzipReader, err := gzip.NewReader(r)
		if err != nil {
			panic("failed to create gzip reader: " + err.Error())
		}
		return gzipReader
	case Zlib:
		zlibReader, err := zlib.NewReader(r)
		if err != nil {
			panic("failed to create zlib reader: " + err.Error())
		}
		return &zlibReadCloser{zlibReader}
	default:
		panic("unsupported compression algorithm")
	}
}

// gzipWriteCloser wraps gzip.Writer to ensure proper closing
type gzipWriteCloser struct {
	*gzip.Writer
}

func (w *gzipWriteCloser) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)
}

func (w *gzipWriteCloser) Close() error {
	if err := w.Writer.Close(); err != nil {
		return errors.Wrap(err, "failed to close gzip writer")
	}
	return nil
}

// zlibWriteCloser wraps zlib.Writer to ensure proper closing
type zlibWriteCloser struct {
	*zlib.Writer
}

func (w *zlibWriteCloser) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)
}

func (w *zlibWriteCloser) Close() error {
	if err := w.Writer.Close(); err != nil {
		return errors.Wrap(err, "failed to close zlib writer")
	}
	return nil
}

// zlibReadCloser wraps zlib reader to implement io.ReadCloser
type zlibReadCloser struct {
	io.ReadCloser
}

func (r *zlibReadCloser) Read(p []byte) (n int, err error) {
	return r.ReadCloser.Read(p)
}

func (r *zlibReadCloser) Close() error {
	return r.ReadCloser.Close()
}