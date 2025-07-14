package hybridbuffer

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"schneider.vip/hybridbuffer/storage"
	"schneider.vip/hybridbuffer/storage/filesystem"
)

func TestHybridBuffer_BasicOperations(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Test write
	data := []byte("Hello, World!")
	n, err := buf.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to write %d bytes, got %d", len(data), n)
	}

	// Test size
	if buf.Size() != int64(len(data)) {
		t.Fatalf("Expected size %d, got %d", len(data), buf.Size())
	}

	// Test read - start from beginning
	readData := make([]byte, len(data))
	n, err = buf.Read(readData)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, readData) {
		t.Fatalf("Expected %q, got %q", string(data), string(readData))
	}
}

func TestHybridBuffer_LargeData(t *testing.T) {
	buf := New(WithThreshold(1024)) // 1KB threshold
	defer buf.Close()

	// Write 2KB of data to trigger storage
	data := make([]byte, 2048)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := buf.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to write %d bytes, got %d", len(data), n)
	}

	// Test sequential reading
	readData := make([]byte, len(data))
	n, err = buf.Read(readData)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, readData) {
		t.Fatalf("Data mismatch after storage transition")
	}
}

func TestHybridBuffer_BytesBufferCompatibility(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Test WriteByte
	err := buf.WriteByte('A')
	if err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}

	// Test WriteRune
	n, err := buf.WriteRune('ñ')
	if err != nil {
		t.Fatalf("WriteRune failed: %v", err)
	}
	if n != 2 { // 'ñ' is 2 bytes in UTF-8
		t.Fatalf("Expected WriteRune to return 2, got %d", n)
	}

	// Test WriteString
	n, err = buf.WriteString("Hello")
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if n != 5 {
		t.Fatalf("Expected WriteString to return 5, got %d", n)
	}

	// Test ReadByte
	b, err := buf.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}
	if b != 'A' {
		t.Fatalf("Expected 'A', got %c", b)
	}

	// Test ReadRune
	r, size, err := buf.ReadRune()
	if err != nil {
		t.Fatalf("ReadRune failed: %v", err)
	}
	if r != 'ñ' {
		t.Fatalf("Expected 'ñ', got %c", r)
	}
	if size != 2 {
		t.Fatalf("Expected size 2, got %d", size)
	}

	// Test ReadString
	s, err := buf.ReadString('o')
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if s != "Hello" {
		t.Fatalf("Expected 'Hello', got %q", s)
	}
}

func TestHybridBuffer_Next(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("0123456789")
	buf.Write(data)

	// Test Next
	next := buf.Next(3)
	if !bytes.Equal(next, []byte("012")) {
		t.Fatalf("Expected '012', got %q", string(next))
	}

	// Test remaining data
	remaining := buf.Next(100) // Request more than available
	if !bytes.Equal(remaining, []byte("3456789")) {
		t.Fatalf("Expected '3456789', got %q", string(remaining))
	}
}

func TestHybridBuffer_String(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := "Hello, World!"
	buf.WriteString(data)

	str := buf.String()
	if str != data {
		t.Fatalf("String() mismatch: expected %q, got %q", data, str)
	}
}

func TestHybridBuffer_Truncate(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("0123456789")
	buf.Write(data)

	// Truncate to 5 bytes
	buf.Truncate(5)
	if buf.Size() != 5 {
		t.Fatalf("Expected size 5 after truncate, got %d", buf.Size())
	}

	// Test remaining data
	remaining := make([]byte, 10)
	n, err := buf.Read(remaining)
	if err != nil && err != io.EOF {
		t.Fatalf("Read after truncate failed: %v", err)
	}
	if n != 5 {
		t.Fatalf("Expected to read 5 bytes after truncate, got %d", n)
	}
	if !bytes.Equal(remaining[:n], []byte("01234")) {
		t.Fatalf("Expected '01234', got %q", string(remaining[:n]))
	}
}

func TestHybridBuffer_Reset(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("Hello, World!")
	buf.Write(data)

	// Reset
	buf.Reset()
	if buf.Size() != 0 {
		t.Fatalf("Expected size 0 after reset, got %d", buf.Size())
	}
	if buf.Len() != 0 {
		t.Fatalf("Expected len 0 after reset, got %d", buf.Len())
	}

	// Test that we can write after reset
	buf.Write([]byte("New data"))
	if buf.Size() != 8 {
		t.Fatalf("Expected size 8 after write, got %d", buf.Size())
	}
}

func TestHybridBuffer_WriteTo(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("Hello, World!")
	buf.Write(data)

	// Write to another buffer
	var target bytes.Buffer
	n, err := buf.WriteTo(&target)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("Expected to write %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, target.Bytes()) {
		t.Fatalf("WriteTo data mismatch")
	}
}

func TestHybridBuffer_ReadFrom(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("Hello, World!")
	source := bytes.NewReader(data)

	n, err := buf.ReadFrom(source)
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if buf.Size() != int64(len(data)) {
		t.Fatalf("Expected size %d, got %d", len(data), buf.Size())
	}
}

func TestHybridBuffer_ReadBytes(t *testing.T) {
	buf := New()
	defer buf.Close()

	buf.WriteString("Hello\nWorld\n")

	// Read first line
	line1, err := buf.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if string(line1) != "Hello\n" {
		t.Fatalf("Expected 'Hello\\n', got %q", string(line1))
	}

	// Read second line
	line2, err := buf.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if string(line2) != "World\n" {
		t.Fatalf("Expected 'World\\n', got %q", string(line2))
	}
}

func TestHybridBuffer_Bytes(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("Hello, World!")
	buf.Write(data)

	result := buf.Bytes()
	if !bytes.Equal(data, result) {
		t.Fatalf("Bytes() mismatch: expected %q, got %q", string(data), string(result))
	}

	// Test that Bytes() consumes the buffer (different from bytes.Buffer!)
	result2 := buf.Bytes()
	if len(result2) != 0 {
		t.Fatalf("Second Bytes() call should return empty slice, got %q", string(result2))
	}
}

func TestHybridBuffer_Available(t *testing.T) {
	threshold := 1024
	buf := New(WithThreshold(threshold))
	defer buf.Close()

	// Initially should have full threshold available
	if buf.Available() != threshold {
		t.Fatalf("Expected available %d, got %d", threshold, buf.Available())
	}

	// Write some data
	data := make([]byte, 500)
	buf.Write(data)

	expected := threshold - 500
	if buf.Available() != expected {
		t.Fatalf("Expected available %d, got %d", expected, buf.Available())
	}

	// Write more data to exceed threshold and switch to storage
	moreData := make([]byte, 600)
	buf.Write(moreData)

	// Should now show 0 available since using storage
	if buf.Available() != 0 {
		t.Fatalf("Expected available 0 after storage switch, got %d", buf.Available())
	}
}

func TestHybridBuffer_WithFilePrefix(t *testing.T) {
	buf := New(
		WithThreshold(10), // Small threshold to trigger storage
		WithStorage(filesystem.New(filesystem.WithPrefix("myapp"))),
	)
	defer buf.Close()

	// Write data to trigger storage
	data := []byte("This is test data that exceeds the small threshold")
	buf.Write(data)

	// Test reading back
	readData := make([]byte, len(data))
	n, err := buf.Read(readData)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, readData) {
		t.Fatalf("Data mismatch: expected %q, got %q", string(data), string(readData))
	}

	// Note: We can't easily test the exact filename since it's generated by CreateTemp,
	// but we can verify the functionality works correctly
}

func TestHybridBuffer_CombinedStorageOptions(t *testing.T) {
	buf := New(
		WithThreshold(10), // Small threshold to trigger storage
		WithStorage(filesystem.New(
			filesystem.WithTempDir("/tmp"),
			filesystem.WithPrefix("testapp"),
		)),
	)
	defer buf.Close()

	// Write data to trigger storage
	data := []byte("This is test data that exceeds the small threshold")
	buf.Write(data)

	// Test reading back
	readData := make([]byte, len(data))
	n, err := buf.Read(readData)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, readData) {
		t.Fatalf("Data mismatch: expected %q, got %q", string(data), string(readData))
	}
}

func TestHybridBuffer_WithCustomStorage(t *testing.T) {
	// Create custom storage with options
	buf := New(
		WithThreshold(5), // Small threshold
		WithStorage(filesystem.New(
			filesystem.WithPrefix("custom"),
			filesystem.WithTempDir("/tmp"),
		)),
	)
	defer buf.Close()

	// Write data to trigger storage
	data := []byte("This is a test with custom storage factory")
	buf.Write(data)

	// Test reading back
	readData := make([]byte, len(data))
	n, err := buf.Read(readData)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(data, readData) {
		t.Fatalf("Data mismatch: expected %q, got %q", string(data), string(readData))
	}
}

func TestHybridBuffer_NewFromBytes(t *testing.T) {
	data := []byte("Initial data from bytes")

	// Test without options
	buf := NewFromBytes(data)
	defer buf.Close()

	// Verify initial data
	if buf.Size() != int64(len(data)) {
		t.Fatalf("Expected size %d, got %d", len(data), buf.Size())
	}

	result := buf.String()
	if result != string(data) {
		t.Fatalf("Expected %q, got %q", string(data), result)
	}

	// Test with options (use larger threshold to avoid test issues)
	buf2 := NewFromBytes(data, WithThreshold(100))
	defer buf2.Close()

	result2 := buf2.String()
	if result2 != string(data) {
		t.Fatalf("Expected %q, got %q", string(data), result2)
	}
}

func TestHybridBuffer_NewFromString(t *testing.T) {
	text := "Initial string data"

	// Test without options
	buf := NewFromString(text)
	defer buf.Close()

	// Verify initial data
	if buf.Size() != int64(len(text)) {
		t.Fatalf("Expected size %d, got %d", len(text), buf.Size())
	}

	result := buf.String()
	if result != text {
		t.Fatalf("Expected %q, got %q", text, result)
	}

	// Test with options
	buf2 := NewFromString(text, WithThreshold(3), WithStorage(filesystem.New(filesystem.WithPrefix("test"))))
	defer buf2.Close()

	result2 := buf2.String()
	if result2 != text {
		t.Fatalf("Expected %q, got %q", text, result2)
	}
}

func TestHybridBuffer_Grow(t *testing.T) {
	buf := New(WithThreshold(100))
	defer buf.Close()

	// Test Grow when still in memory
	buf.Grow(50)

	// Write some data
	data := []byte("Test data for grow")
	buf.Write(data)

	// Verify data
	result := buf.String()
	if result != string(data) {
		t.Fatalf("Expected %q, got %q", string(data), result)
	}

	// Test Grow after switching to storage (should be no-op)
	buf2 := New(WithThreshold(5))
	defer buf2.Close()

	largeData := []byte("This data exceeds threshold")
	buf2.Write(largeData) // Triggers storage

	buf2.Grow(100) // Should be ignored since using storage

	result2 := buf2.String()
	if result2 != string(largeData) {
		t.Fatalf("Expected %q, got %q", string(largeData), result2)
	}
}

func TestHybridBuffer_WithPreAlloc(t *testing.T) {
	// Test with custom pre-allocation size
	buf := New(
		WithThreshold(1024),
		WithPreAlloc(512),
	)
	defer buf.Close()

	// Write data that fits in pre-allocated space
	data := make([]byte, 400)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := buf.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to write %d bytes, got %d", len(data), n)
	}

	// Verify data
	result := buf.String()
	if len(result) != len(data) {
		t.Fatalf("Expected %d bytes, got %d", len(data), len(result))
	}

	// Test with zero pre-allocation (should use default)
	buf2 := New(
		WithThreshold(1024),
		WithPreAlloc(0), // Invalid, should be ignored
	)
	defer buf2.Close()

	buf2.Write(data)
	result2 := buf2.String()
	if len(result2) != len(data) {
		t.Fatalf("Expected %d bytes with default pre-alloc, got %d", len(data), len(result2))
	}

	// Test with negative pre-allocation (should use default)
	buf3 := New(
		WithThreshold(1024),
		WithPreAlloc(-100), // Invalid, should be ignored
	)
	defer buf3.Close()

	buf3.Write(data)
	result3 := buf3.String()
	if len(result3) != len(data) {
		t.Fatalf("Expected %d bytes with default pre-alloc, got %d", len(data), len(result3))
	}

	// Test with large pre-allocation
	buf4 := New(
		WithThreshold(2048),
		WithPreAlloc(1024),
	)
	defer buf4.Close()

	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	buf4.Write(largeData)
	result4 := buf4.String()
	if len(result4) != len(largeData) {
		t.Fatalf("Expected %d bytes with large pre-alloc, got %d", len(largeData), len(result4))
	}
}

func TestHybridBuffer_CustomStorageBackend(t *testing.T) {
	// Create a mock storage backend for testing
	mockBackend := &mockStorageBackend{
		data: make([]byte, 0),
	}

	buf := New(
		WithThreshold(5),
		WithStorage(func() storage.Backend { return mockBackend }),
	)
	defer buf.Close()

	// Write data to trigger storage
	data := []byte("Test data for custom backend")
	buf.Write(data)

	// Verify mock was used
	if !mockBackend.createCalled {
		t.Fatal("Create was not called on mock backend")
	}

	// Verify data - this will trigger Open
	result := buf.String()
	if result != string(data) {
		t.Fatalf("Expected %q, got %q", string(data), result)
	}

	if !mockBackend.openCalled {
		t.Fatal("Open was not called on mock backend")
	}
}

// Mock storage backend for testing
type mockStorageBackend struct {
	data         []byte
	createCalled bool
	openCalled   bool
	removeCalled bool
	writePos     int
	readPos      int
}

func (m *mockStorageBackend) Create() (io.WriteCloser, error) {
	m.createCalled = true
	m.writePos = 0
	return &mockWriteCloser{backend: m}, nil
}

func (m *mockStorageBackend) Open() (io.ReadCloser, error) {
	m.openCalled = true
	m.readPos = 0
	return &mockReadCloser{backend: m}, nil
}

func (m *mockStorageBackend) Remove() error {
	m.removeCalled = true
	m.data = nil
	return nil
}

type mockWriteCloser struct {
	backend *mockStorageBackend
}

func (mw *mockWriteCloser) Write(p []byte) (n int, err error) {
	// Extend data slice if needed
	if len(mw.backend.data) < mw.backend.writePos+len(p) {
		newData := make([]byte, mw.backend.writePos+len(p))
		copy(newData, mw.backend.data)
		mw.backend.data = newData
	}

	copy(mw.backend.data[mw.backend.writePos:], p)
	mw.backend.writePos += len(p)
	return len(p), nil
}

func (mw *mockWriteCloser) Close() error {
	return nil
}

type mockReadCloser struct {
	backend *mockStorageBackend
}

func (mr *mockReadCloser) Read(p []byte) (n int, err error) {
	if mr.backend.readPos >= len(mr.backend.data) {
		return 0, io.EOF
	}

	available := len(mr.backend.data) - mr.backend.readPos
	toRead := len(p)
	if toRead > available {
		toRead = available
	}

	copy(p, mr.backend.data[mr.backend.readPos:mr.backend.readPos+toRead])
	mr.backend.readPos += toRead

	if mr.backend.readPos >= len(mr.backend.data) {
		return toRead, io.EOF
	}
	return toRead, nil
}

func (mr *mockReadCloser) Close() error {
	return nil
}

func TestHybridBuffer_StringBytesConsumption(t *testing.T) {
	// Test that String() and Bytes() consume buffer content (unlike bytes.Buffer)
	buf := New()
	defer buf.Close()

	data := "Hello, World!"
	buf.WriteString(data)

	// Verify initial state
	if buf.Len() != len(data) {
		t.Fatalf("Expected Len() %d, got %d", len(data), buf.Len())
	}

	// First String() call should return data
	result1 := buf.String()
	if result1 != data {
		t.Fatalf("First String() call: expected %q, got %q", data, result1)
	}

	// Buffer should now be consumed
	if buf.Len() != 0 {
		t.Fatalf("After String() call, expected Len() 0, got %d", buf.Len())
	}

	// Second String() call should return empty
	result2 := buf.String()
	if result2 != "" {
		t.Fatalf("Second String() call: expected empty string, got %q", result2)
	}

	// Test same behavior with Bytes()
	buf2 := New()
	defer buf2.Close()
	buf2.WriteString(data)

	// First Bytes() call should return data
	bytes1 := buf2.Bytes()
	if string(bytes1) != data {
		t.Fatalf("First Bytes() call: expected %q, got %q", data, string(bytes1))
	}

	// Second Bytes() call should return empty
	bytes2 := buf2.Bytes()
	if len(bytes2) != 0 {
		t.Fatalf("Second Bytes() call: expected empty slice, got %q", string(bytes2))
	}
}

func TestHybridBuffer_CloseErrors(t *testing.T) {
	buf := New(WithThreshold(5))

	// Force creation of storage
	buf.Write([]byte("test data exceeding threshold"))

	// Force streams to exist
	_ = buf.String() // This creates read/write streams

	// Close should handle multiple stream closures
	err := buf.Close()
	if err != nil {
		t.Logf("Close error (expected): %v", err)
	}

	// Second close should not panic
	err2 := buf.Close()
	if err2 != nil {
		t.Logf("Second close error (expected): %v", err2)
	}
}

func TestHybridBuffer_TruncateEdgeCases(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("0123456789")
	buf.Write(data)

	// Test truncate with invalid values
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for negative truncate")
		}
	}()
	buf.Truncate(-1)
}

func TestHybridBuffer_TruncateOutOfRange(t *testing.T) {
	buf := New()
	defer buf.Close()

	data := []byte("0123456789")
	buf.Write(data)

	// Test truncate beyond size
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for truncate beyond size")
		}
	}()
	buf.Truncate(20) // More than data length
}

func TestHybridBuffer_ReadRuneIncomplete(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Write incomplete UTF-8 sequence
	buf.Write([]byte{0xC0}) // Incomplete UTF-8

	r, size, err := buf.ReadRune()
	if err != nil {
		t.Logf("ReadRune with incomplete UTF-8: r=%c, size=%d, err=%v", r, size, err)
	}
	// Should handle gracefully without panic
}

func TestHybridBuffer_NextEdgeCases(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Test Next with zero
	result := buf.Next(0)
	if result != nil {
		t.Fatalf("Next(0) should return nil, got %v", result)
	}

	// Test Next with negative
	result = buf.Next(-5)
	if result != nil {
		t.Fatalf("Next(-5) should return nil, got %v", result)
	}

	// Test Next on empty buffer
	result = buf.Next(10)
	if result != nil {
		t.Fatalf("Next(10) on empty buffer should return nil, got %v", result)
	}
}

func TestHybridBuffer_EdgeCases(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Test empty buffer operations
	if buf.Size() != 0 {
		t.Fatalf("Expected size 0 for empty buffer, got %d", buf.Size())
	}

	if buf.Len() != 0 {
		t.Fatalf("Expected len 0 for empty buffer, got %d", buf.Len())
	}

	// Test reading from empty buffer
	data := make([]byte, 10)
	n, err := buf.Read(data)
	if err != io.EOF {
		t.Fatalf("Expected EOF reading from empty buffer, got %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 bytes read from empty buffer, got %d", n)
	}

	// Test writing empty data
	n, err = buf.Write(nil)
	if err != nil {
		t.Fatalf("Writing nil data failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 bytes written for nil data, got %d", n)
	}

	n, err = buf.Write([]byte{})
	if err != nil {
		t.Fatalf("Writing empty data failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 bytes written for empty data, got %d", n)
	}
}

func TestHybridBuffer_StorageErrors(t *testing.T) {
	// Test storage backend creation with invalid temp dir
	buf := New(
		WithThreshold(1),
		WithStorage(filesystem.New(filesystem.WithTempDir("/nonexistent/path/that/should/fail"))),
	)
	defer buf.Close()

	// This should trigger storage creation and potentially fail
	_, err := buf.Write([]byte("test data"))
	if err != nil {
		t.Logf("Expected error with invalid temp dir: %v", err)
		// This is expected behavior
	}
}

func TestHybridBuffer_ReadErrors(t *testing.T) {
	buf := New(WithThreshold(2))
	defer buf.Close()

	// Write data to trigger storage
	buf.Write([]byte("test data"))

	// Test ReadByte on empty result
	_ = buf.String() // Consume all data

	_, err := buf.ReadByte()
	if err != io.EOF {
		t.Fatalf("Expected EOF, got %v", err)
	}
}

func TestHybridBuffer_WriteTo_ReadFrom_Errors(t *testing.T) {
	// Test error handling in WriteTo
	buf := New()
	defer buf.Close()

	buf.WriteString("test data")

	// WriteTo should handle errors from the writer
	failWriter := &failingWriter{}
	n, err := buf.WriteTo(failWriter)
	if err == nil {
		t.Fatal("Expected error from WriteTo")
	}
	t.Logf("WriteTo handled error correctly: wrote %d bytes, err=%v", n, err)

	// Test ReadFrom with failing reader
	buf2 := New()
	defer buf2.Close()

	failReader := &failingReader{}
	n2, err2 := buf2.ReadFrom(failReader)
	if err2 == nil {
		t.Fatal("Expected error from ReadFrom")
	}
	t.Logf("ReadFrom handled error correctly: read %d bytes, err=%v", n2, err2)
}

// Helper types for error testing
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	if len(p) > 5 {
		return 5, fmt.Errorf("simulated write failure")
	}
	return len(p), nil
}

type failingReader struct{}

func (fr *failingReader) Read(p []byte) (n int, err error) {
	if len(p) > 0 {
		p[0] = 'x'
		return 1, fmt.Errorf("simulated read failure")
	}
	return 0, fmt.Errorf("simulated read failure")
}

func TestHybridBuffer_ReadRuneEdgeCases(t *testing.T) {
	buf := New()
	defer buf.Close()

	// Test complete rune reading
	buf.WriteString("Hello")
	r, size, err := buf.ReadRune()
	if err != nil {
		t.Fatalf("ReadRune failed: %v", err)
	}
	if r != 'H' || size != 1 {
		t.Fatalf("Expected 'H', size 1, got %c, size %d", r, size)
	}

	// Test multi-byte rune
	buf2 := New()
	defer buf2.Close()
	buf2.WriteString("ñ") // 2-byte UTF-8
	r2, size2, err2 := buf2.ReadRune()
	if err2 != nil {
		t.Fatalf("ReadRune multi-byte failed: %v", err2)
	}
	if r2 != 'ñ' || size2 != 2 {
		t.Fatalf("Expected 'ñ', size 2, got %c, size %d", r2, size2)
	}

	// Test EOF case
	buf3 := New()
	defer buf3.Close()
	_, _, err3 := buf3.ReadRune()
	if err3 != io.EOF {
		t.Fatalf("Expected EOF, got %v", err3)
	}
}

func TestHybridBuffer_TruncateZero(t *testing.T) {
	buf := New()
	defer buf.Close()

	buf.WriteString("test data")

	// Test truncate to zero (should call Reset)
	buf.Truncate(0)

	if buf.Len() != 0 {
		t.Fatalf("Expected Len() 0 after Truncate(0), got %d", buf.Len())
	}

	if buf.Size() != 0 {
		t.Fatalf("Expected Size() 0 after Truncate(0), got %d", buf.Size())
	}
}

func TestHybridBuffer_WriteStreamAlreadyOpen(t *testing.T) {
	buf := New(WithThreshold(1))
	defer buf.Close()

	// Write to trigger storage
	buf.Write([]byte("test"))

	// Access the hybridBuffer directly to test writeStream reuse
	if hybridBuf, ok := buf.(*hybridBuffer); ok {
		// Force a write stream to be created
		if hybridBuf.writeStream == nil {
			hybridBuf.openWriteStream()
		}

		// Now try to open again (should be no-op)
		err := hybridBuf.openWriteStream()
		if err != nil {
			t.Fatalf("openWriteStream on already open stream failed: %v", err)
		}
	}
}

func TestHybridBuffer_ReadStreamAlreadyOpen(t *testing.T) {
	buf := New(WithThreshold(1))
	defer buf.Close()

	// Write to trigger storage
	buf.Write([]byte("test"))

	// Read to trigger read stream
	readBuf := make([]byte, 1)
	buf.Read(readBuf)

	// Access the hybridBuffer directly to test readStream reuse
	if hybridBuf, ok := buf.(*hybridBuffer); ok {
		// Now try to open again (should be no-op)
		err := hybridBuf.openReadStream()
		if err != nil {
			t.Fatalf("openReadStream on already open stream failed: %v", err)
		}
	}
}

// Benchmark tests
func BenchmarkHybridBuffer_Write(b *testing.B) {
	buf := New()
	defer buf.Close()

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Write(data)
	}
}

func BenchmarkHybridBuffer_Read(b *testing.B) {
	buf := New()
	defer buf.Close()

	// Prepare data
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	buf.Write(data)

	readData := make([]byte, 1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%1000 == 0 { // Reset periodically
			buf.Reset()
			buf.Write(data)
		}
		buf.Read(readData)
	}
}
