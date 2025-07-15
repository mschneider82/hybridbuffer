package hybridbuffer

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/pkg/errors"
	"schneider.vip/hybridbuffer/middleware"
	"schneider.vip/hybridbuffer/storage"
	"schneider.vip/hybridbuffer/storage/filesystem"
)

// Buffer defines the interface for hybrid memory/disk buffers
type Buffer interface {
	io.ReadWriter
	io.ReaderFrom
	io.WriterTo
	io.ByteReader
	io.ByteWriter
	io.StringWriter

	// bytes.Buffer compatible methods
	ReadBytes(delim byte) ([]byte, error)
	ReadString(delim byte) (string, error)
	ReadRune() (r rune, size int, err error)
	WriteRune(r rune) (n int, err error)
	Next(n int) []byte

	// Data access (WARNING: Unlike bytes.Buffer, these consume the buffer content!)
	Bytes() []byte
	String() string

	// Size and capacity
	Len() int
	Cap() int
	Available() int
	Size() int64

	// Buffer management
	Reset()
	Truncate(n int)
	Grow(n int)
	Close() error
}

// hybridBuffer implements Buffer interface
type hybridBuffer struct {
	threshold       int
	size            int
	offset          int
	memoryBuffer    bytes.Buffer
	storageBackend  storage.Backend
	storageProvider func() storage.Backend
	writeStream     io.WriteCloser
	readStream      io.ReadCloser
	middlewares     []middleware.Middleware
	usingStorage    bool
	preAllocSize    int // Size to pre-allocate in memory buffer
}

// New creates a new hybrid buffer with the given options
func New(opts ...Option) Buffer {
	buf := &hybridBuffer{
		threshold: 2 << 20, // 2MB default
		// Will be set by default WithFilesystemStorage() option below
		middlewares: []middleware.Middleware{}, // No middlewares by default
	}

	// Apply default filesystem storage if none specified
	WithStorage(filesystem.New())(buf)

	// Apply options
	for _, opt := range opts {
		opt(buf)
	}

	// Set default pre-allocation size if not specified
	if buf.preAllocSize == 0 {
		buf.preAllocSize = buf.threshold / 2
	}

	// Pre-allocate memory buffer
	buf.memoryBuffer.Grow(buf.preAllocSize)

	return buf
}

// NewFromBytes creates a buffer with initial data
func NewFromBytes(data []byte, opts ...Option) Buffer {
	buf := New(opts...)
	buf.Write(data)
	return buf
}

// NewFromString creates a buffer with initial string data
func NewFromString(s string, opts ...Option) Buffer {
	return NewFromBytes([]byte(s), opts...)
}

// Write implements io.Writer
func (b *hybridBuffer) Write(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	// Check if we need to switch to storage
	if !b.usingStorage && b.memoryBuffer.Len()+len(data) > b.threshold {
		if err = b.flushToStorage(); err != nil {
			return 0, errors.Wrap(err, "failed to flush to storage")
		}
	}

	if b.usingStorage {
		// Write to storage
		if b.writeStream == nil {
			if err = b.openWriteStream(); err != nil {
				return 0, errors.Wrap(err, "failed to open write stream")
			}
		}
		n, err = b.writeStream.Write(data)
	} else {
		// Write to memory
		n, err = b.memoryBuffer.Write(data)
	}

	if err == nil {
		b.size += n
	}
	return n, err
}

// Read implements io.Reader
func (b *hybridBuffer) Read(data []byte) (n int, err error) {
	if b.offset >= b.size {
		return 0, io.EOF
	}

	// Ensure write stream is closed before reading (critical for encryption)
	if b.writeStream != nil {
		b.writeStream.Close()
		b.writeStream = nil
	}

	// Read from current offset
	bytesToRead := len(data)
	available := b.size - b.offset
	if bytesToRead > available {
		bytesToRead = available
	}

	if b.usingStorage {
		// Read from storage
		if b.readStream == nil {
			if err = b.openReadStream(); err != nil {
				return 0, errors.Wrap(err, "failed to open read stream")
			}
		}
		n, err = b.readStream.Read(data[:bytesToRead])
	} else {
		// Read from memory buffer
		memData := b.memoryBuffer.Bytes()
		if b.offset < len(memData) {
			endPos := b.offset + bytesToRead
			if endPos > len(memData) {
				endPos = len(memData)
			}
			n = copy(data, memData[b.offset:endPos])
		}
	}

	b.offset += n
	return n, err
}

// WriteTo implements io.WriterTo
func (b *hybridBuffer) WriteTo(w io.Writer) (int64, error) {
	var n int64
	data := make([]byte, 512)
	for {
		rN, rErr := b.Read(data)
		if rErr != nil && rErr != io.EOF {
			return n, rErr
		}

		if rN > 0 {
			wN, wErr := w.Write(data[:rN])
			n += int64(wN)
			if wErr != nil {
				return n, wErr
			}
		}

		if rErr == io.EOF {
			return n, nil
		}
	}
}

// ReadFrom implements io.ReaderFrom
func (b *hybridBuffer) ReadFrom(r io.Reader) (int64, error) {
	var n int64
	data := make([]byte, 512)
	for {
		rN, rErr := r.Read(data)
		if rErr != nil && rErr != io.EOF {
			return n, rErr
		}

		if rN > 0 {
			wN, wErr := b.Write(data[:rN])
			n += int64(wN)
			if wErr != nil {
				return n, wErr
			}
		}

		if rErr == io.EOF {
			return n, nil
		}
	}
}

// WriteByte implements io.ByteWriter
func (b *hybridBuffer) WriteByte(c byte) error {
	_, err := b.Write([]byte{c})
	return err
}

// WriteRune writes a rune (compatible with bytes.Buffer)
func (b *hybridBuffer) WriteRune(r rune) (n int, err error) {
	var buf [utf8.UTFMax]byte
	n = utf8.EncodeRune(buf[:], r)
	return b.Write(buf[:n])
}

// WriteString implements io.StringWriter
func (b *hybridBuffer) WriteString(s string) (n int, err error) {
	return b.Write([]byte(s))
}

// ReadByte implements io.ByteReader
func (b *hybridBuffer) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := b.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	return buf[0], nil
}

// ReadBytes reads until delimiter (compatible with bytes.Buffer)
func (b *hybridBuffer) ReadBytes(delim byte) ([]byte, error) {
	var result []byte
	for {
		c, err := b.ReadByte()
		if err != nil {
			return result, err
		}

		result = append(result, c)
		if c == delim {
			return result, nil
		}
	}
}

// ReadString reads until delimiter (compatible with bytes.Buffer)
func (b *hybridBuffer) ReadString(delim byte) (string, error) {
	bytes, err := b.ReadBytes(delim)
	return string(bytes), err
}

// ReadRune reads a rune (compatible with bytes.Buffer)
func (b *hybridBuffer) ReadRune() (r rune, size int, err error) {
	var buf [utf8.UTFMax]byte
	var n int

	for n < utf8.UTFMax {
		c, err := b.ReadByte()
		if err != nil {
			if n == 0 {
				return 0, 0, err
			}
			break
		}

		buf[n] = c
		n++

		if utf8.FullRune(buf[:n]) {
			r, size = utf8.DecodeRune(buf[:n])
			return r, size, nil
		}
	}

	// If we get here, we have an incomplete rune
	r, size = utf8.DecodeRune(buf[:n])
	return r, size, nil
}

// Next returns the next n bytes (compatible with bytes.Buffer)
func (b *hybridBuffer) Next(n int) []byte {
	if n <= 0 {
		return nil
	}

	// Read up to n bytes from current position
	available := b.Len()
	if n > available {
		n = available
	}

	if n == 0 {
		return nil
	}

	buf := make([]byte, n)
	readBytes, err := b.Read(buf)
	if err != nil && err != io.EOF {
		panic(err) // bytes.Buffer.Next() panics on error
	}

	return buf[:readBytes]
}

// Len returns the number of unread bytes (compatible with bytes.Buffer)
func (b *hybridBuffer) Len() int {
	return b.size - b.offset
}

// Cap returns the capacity (equal to Len for compatibility)
func (b *hybridBuffer) Cap() int {
	return b.Len()
}

// Available returns available capacity in the buffer (compatible with bytes.Buffer)
func (b *hybridBuffer) Available() int {
	if b.usingStorage {
		return 0
	}
	return b.threshold - b.memoryBuffer.Len()
}

// Size returns the total size of data written
func (b *hybridBuffer) Size() int64 {
	return int64(b.size)
}

// Reset resets the buffer to initial state (compatible with bytes.Buffer)
func (b *hybridBuffer) Reset() {
	// Close streams
	if b.writeStream != nil {
		b.writeStream.Close()
		b.writeStream = nil
	}
	if b.readStream != nil {
		b.readStream.Close()
		b.readStream = nil
	}

	// Remove storage
	if b.storageBackend != nil {
		b.storageBackend.Remove()
		b.storageBackend = nil
	}

	// Reset state
	b.memoryBuffer.Reset()
	b.size = 0
	b.offset = 0
	b.usingStorage = false
}

// Close closes the buffer and cleans up resources
func (b *hybridBuffer) Close() error {
	var lastErr error

	// Close streams
	if b.writeStream != nil {
		if err := b.writeStream.Close(); err != nil {
			lastErr = err
		}
		b.writeStream = nil
	}
	if b.readStream != nil {
		if err := b.readStream.Close(); err != nil {
			lastErr = err
		}
		b.readStream = nil
	}

	// Remove storage
	if b.storageBackend != nil {
		if err := b.storageBackend.Remove(); err != nil {
			lastErr = err
		}
		b.storageBackend = nil
	}

	return lastErr
}

// Bytes returns the contents as a byte slice
//
// IMPORTANT DIFFERENCE from bytes.Buffer:
// Unlike bytes.Buffer.Bytes(), this method CONSUMES the buffer content and advances
// the read position. Subsequent calls will return different/empty results.
// This method is primarily intended for testing and final data retrieval.
//
// WARNING: This loads ALL remaining data into memory! Use with caution for large buffers.
func (b *hybridBuffer) Bytes() []byte {
	// Ensure write stream is closed before reading
	if b.writeStream != nil {
		b.writeStream.Close()
		b.writeStream = nil
	}

	// Read all remaining data from current position
	remaining := b.Len()
	if remaining == 0 {
		return nil
	}

	result := make([]byte, remaining)
	n, err := b.Read(result)
	if err != nil && err != io.EOF {
		// If read fails, return what we got
		return result[:n]
	}

	return result[:n]
}

// String returns the contents as a string
//
// IMPORTANT DIFFERENCE from bytes.Buffer:
// Unlike bytes.Buffer.String(), this method CONSUMES the buffer content and advances
// the read position. Subsequent calls will return different/empty results.
// This method is primarily intended for testing and final data retrieval.
//
// WARNING: This loads ALL remaining data into memory! Use with caution for large buffers.
func (b *hybridBuffer) String() string {
	return string(b.Bytes())
}

// Grow grows the buffer's capacity (compatible with bytes.Buffer)
func (b *hybridBuffer) Grow(n int) {
	// Only grow if we're still in memory phase
	if !b.usingStorage {
		b.memoryBuffer.Grow(n)
	}
}

// Truncate truncates the buffer (compatible with bytes.Buffer)
func (b *hybridBuffer) Truncate(n int) {
	if n < 0 || n > b.size {
		panic("hybridbuffer: truncation out of range")
	}

	if n == 0 {
		b.Reset()
		return
	}

	// For simplicity, if we're truncating, we reset and re-write the first n bytes
	// Save current data up to n bytes
	oldOffset := b.offset
	b.offset = 0

	// Reset read stream to start from beginning
	if b.readStream != nil {
		b.readStream.Close()
		b.readStream = nil
	}

	data := make([]byte, n)
	actualRead, _ := b.Read(data)

	// Reset and write back
	b.Reset()
	b.Write(data[:actualRead])

	// Restore offset if it was within the truncated range
	if oldOffset < actualRead {
		b.offset = oldOffset
	}
}

// flushToStorage moves all memory data to storage
func (b *hybridBuffer) flushToStorage() error {
	if b.usingStorage {
		return nil // Already using storage
	}

	// Create storage backend
	b.storageBackend = b.storageProvider()

	// Open write stream
	if err := b.openWriteStream(); err != nil {
		return err
	}

	// Write memory buffer to storage
	memData := b.memoryBuffer.Bytes()
	if len(memData) > 0 {
		if _, err := b.writeStream.Write(memData); err != nil {
			return errors.Wrap(err, "failed to write memory data to storage")
		}
	}

	// Switch to storage mode
	b.usingStorage = true
	return nil
}

// openWriteStream opens a write stream for storage
func (b *hybridBuffer) openWriteStream() error {
	if b.writeStream != nil {
		return nil // Already open
	}

	writeStream, err := b.storageBackend.Create()
	if err != nil {
		return errors.Wrap(err, "failed to create storage write stream")
	}

	// Apply middleware pipeline in forward order (first middleware first)
	writer := io.Writer(writeStream)
	for _, middleware := range b.middlewares {
		writer = middleware.Writer(writer)
	}

	// Convert back to WriteCloser
	if wc, ok := writer.(io.WriteCloser); ok {
		b.writeStream = wc
	} else {
		b.writeStream = &writeCloserWrapper{
			Writer:     writer,
			underlying: writeStream,
		}
	}

	return nil
}

// openReadStream opens a read stream for storage
func (b *hybridBuffer) openReadStream() error {
	if b.readStream != nil {
		return nil // Already open
	}

	readStream, err := b.storageBackend.Open()
	if err != nil {
		return errors.Wrap(err, "failed to open storage read stream")
	}

	// Apply middleware pipeline in reverse order (last middleware first)
	reader := io.Reader(readStream)
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		reader = b.middlewares[i].Reader(reader)
	}

	// Convert back to ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		b.readStream = rc
	} else {
		b.readStream = &readCloserWrapper{
			Reader:     reader,
			underlying: readStream,
		}
	}

	return nil
}

// Wrapper types for middleware pipeline
type writeCloserWrapper struct {
	io.Writer
	underlying io.WriteCloser
}

func (w *writeCloserWrapper) Close() error {
	// Close the middleware writer first if it implements io.Closer
	if closer, ok := w.Writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			w.underlying.Close() // Still close underlying
			return err
		}
	}
	return w.underlying.Close()
}

type readCloserWrapper struct {
	io.Reader
	underlying io.ReadCloser
}

func (r *readCloserWrapper) Close() error {
	// Close the middleware reader first if it implements io.Closer
	if closer, ok := r.Reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			r.underlying.Close() // Still close underlying
			return err
		}
	}
	return r.underlying.Close()
}
