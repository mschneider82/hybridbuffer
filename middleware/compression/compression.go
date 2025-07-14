// Package compression provides advanced compression middleware for HybridBuffer using klauspost/compress
package compression

import (
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/flate"
	"schneider.vip/hybridbuffer/middleware"
)

// Algorithm represents the compression algorithm to use
type Algorithm int

const (
	// Gzip compression using klauspost/compress/gzip (faster than stdlib)
	Gzip Algorithm = iota
	// Zstd compression - excellent compression ratio and speed
	Zstd
	// S2 compression - very fast compression/decompression
	S2
	// Snappy compression - very fast, moderate compression
	Snappy
	// Zlib compression using klauspost/compress/zlib
	Zlib
	// Flate compression (raw deflate)
	Flate
)

// Level represents compression level
type Level int

const (
	// Fastest compression with least CPU usage
	Fastest Level = iota
	// Default balanced compression
	Default
	// Better compression with more CPU usage
	Better
	// Best compression with most CPU usage
	Best
)

// Middleware implements compression/decompression
type Middleware struct {
	algorithm Algorithm
	level     Level
}

// Ensure Middleware implements middleware.Middleware interface
var _ middleware.Middleware = (*Middleware)(nil)

// Option configures compression middleware
type Option func(*Middleware)

// WithLevel sets the compression level
func WithLevel(level Level) Option {
	return func(m *Middleware) {
		m.level = level
	}
}

// New creates a new compression middleware with the given algorithm
func New(algorithm Algorithm, opts ...Option) *Middleware {
	m := &Middleware{
		algorithm: algorithm,
		level:     Default, // Default compression level
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
		return m.createGzipWriter(w)
	case Zstd:
		return m.createZstdWriter(w)
	case S2:
		return m.createS2Writer(w)
	case Snappy:
		return m.createSnappyWriter(w)
	case Zlib:
		return m.createZlibWriter(w)
	case Flate:
		return m.createFlateWriter(w)
	default:
		panic("unsupported compression algorithm")
	}
}

// Reader wraps an io.Reader with decompression
func (m *Middleware) Reader(r io.Reader) io.Reader {
	switch m.algorithm {
	case Gzip:
		return m.createGzipReader(r)
	case Zstd:
		return m.createZstdReader(r)
	case S2:
		return m.createS2Reader(r)
	case Snappy:
		return m.createSnappyReader(r)
	case Zlib:
		return m.createZlibReader(r)
	case Flate:
		return m.createFlateReader(r)
	default:
		panic("unsupported compression algorithm")
	}
}

// Gzip compression methods
func (m *Middleware) createGzipWriter(w io.Writer) io.Writer {
	var level int
	switch m.level {
	case Fastest:
		level = gzip.BestSpeed
	case Default:
		level = gzip.DefaultCompression
	case Better:
		level = gzip.BestCompression - 1
	case Best:
		level = gzip.BestCompression
	}
	
	gzipWriter, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		panic("failed to create gzip writer: " + err.Error())
	}
	return &gzipWriteCloser{gzipWriter}
}

func (m *Middleware) createGzipReader(r io.Reader) io.Reader {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		panic("failed to create gzip reader: " + err.Error())
	}
	return gzipReader
}

// Zstd compression methods
func (m *Middleware) createZstdWriter(w io.Writer) io.Writer {
	var level zstd.EncoderLevel
	switch m.level {
	case Fastest:
		level = zstd.SpeedFastest
	case Default:
		level = zstd.SpeedDefault
	case Better:
		level = zstd.SpeedBetterCompression
	case Best:
		level = zstd.SpeedBestCompression
	}
	
	zstdWriter, err := zstd.NewWriter(w, zstd.WithEncoderLevel(level))
	if err != nil {
		panic("failed to create zstd writer: " + err.Error())
	}
	return &zstdWriteCloser{zstdWriter}
}

func (m *Middleware) createZstdReader(r io.Reader) io.Reader {
	zstdReader, err := zstd.NewReader(r)
	if err != nil {
		panic("failed to create zstd reader: " + err.Error())
	}
	return &zstdReadCloser{zstdReader}
}

// S2 compression methods
func (m *Middleware) createS2Writer(w io.Writer) io.Writer {
	return s2.NewWriter(w)
}

func (m *Middleware) createS2Reader(r io.Reader) io.Reader {
	return s2.NewReader(r)
}

// Snappy compression methods
func (m *Middleware) createSnappyWriter(w io.Writer) io.Writer {
	return snappy.NewBufferedWriter(w)
}

func (m *Middleware) createSnappyReader(r io.Reader) io.Reader {
	return snappy.NewReader(r)
}

// Zlib compression methods
func (m *Middleware) createZlibWriter(w io.Writer) io.Writer {
	var level int
	switch m.level {
	case Fastest:
		level = zlib.BestSpeed
	case Default:
		level = zlib.DefaultCompression
	case Better:
		level = zlib.BestCompression - 1
	case Best:
		level = zlib.BestCompression
	}
	
	zlibWriter, err := zlib.NewWriterLevel(w, level)
	if err != nil {
		panic("failed to create zlib writer: " + err.Error())
	}
	return &zlibWriteCloser{zlibWriter}
}

func (m *Middleware) createZlibReader(r io.Reader) io.Reader {
	zlibReader, err := zlib.NewReader(r)
	if err != nil {
		panic("failed to create zlib reader: " + err.Error())
	}
	return &zlibReadCloser{zlibReader}
}

// Flate compression methods
func (m *Middleware) createFlateWriter(w io.Writer) io.Writer {
	var level int
	switch m.level {
	case Fastest:
		level = flate.BestSpeed
	case Default:
		level = flate.DefaultCompression
	case Better:
		level = flate.BestCompression - 1
	case Best:
		level = flate.BestCompression
	}
	
	flateWriter, err := flate.NewWriter(w, level)
	if err != nil {
		panic("failed to create flate writer: " + err.Error())
	}
	return &flateWriteCloser{flateWriter}
}

func (m *Middleware) createFlateReader(r io.Reader) io.Reader {
	flateReader := flate.NewReader(r)
	return &flateReadCloser{flateReader}
}

// Wrapper types for proper io.WriteCloser implementation

type gzipWriteCloser struct {
	*gzip.Writer
}

func (w *gzipWriteCloser) Close() error {
	return w.Writer.Close()
}

type zstdWriteCloser struct {
	*zstd.Encoder
}

func (w *zstdWriteCloser) Close() error {
	return w.Encoder.Close()
}

type zstdReadCloser struct {
	*zstd.Decoder
}

func (r *zstdReadCloser) Close() error {
	r.Decoder.Close()
	return nil
}

type zlibWriteCloser struct {
	*zlib.Writer
}

func (w *zlibWriteCloser) Close() error {
	return w.Writer.Close()
}

type zlibReadCloser struct {
	io.ReadCloser
}

func (r *zlibReadCloser) Close() error {
	return r.ReadCloser.Close()
}

type flateWriteCloser struct {
	*flate.Writer
}

func (w *flateWriteCloser) Close() error {
	return w.Writer.Close()
}

type flateReadCloser struct {
	io.ReadCloser
}

func (r *flateReadCloser) Close() error {
	return r.ReadCloser.Close()
}