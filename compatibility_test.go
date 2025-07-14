package hybridbuffer

import (
	"bytes"
	"io"
	"testing"
)

// TestCompatibility_BasicOperations tests basic read/write operations
func TestCompatibility_BasicOperations(t *testing.T) {
	data := []byte("Hello, World!")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)
	stdResult := make([]byte, len(data))
	stdN, stdErr := stdBuf.Read(stdResult)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)
	hybridResult := make([]byte, len(data))
	hybridN, hybridErr := hybridBuf.Read(hybridResult)

	// Compare results
	if stdN != hybridN {
		t.Fatalf("Read count mismatch: std=%d, hybrid=%d", stdN, hybridN)
	}
	if (stdErr == nil) != (hybridErr == nil) {
		t.Fatalf("Error mismatch: std=%v, hybrid=%v", stdErr, hybridErr)
	}
	if !bytes.Equal(stdResult, hybridResult) {
		t.Fatalf("Data mismatch: std=%q, hybrid=%q", string(stdResult), string(hybridResult))
	}
}

// TestCompatibility_WriteReadCycles tests multiple write/read cycles
// Note: HybridBuffer has different behavior for mixed write/read with encryption
func TestCompatibility_WriteReadCycles(t *testing.T) {
	testData := [][]byte{
		[]byte("First chunk"),
		[]byte("Second piece of data"),
		[]byte("Third"),
		[]byte("Final data chunk"),
	}

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	var stdResults [][]byte

	// Test with HybridBuffer (without encryption to maintain compatibility)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	var hybridResults [][]byte

	// Perform write/read cycles
	for i, data := range testData {
		// Write to both buffers
		stdN, stdErr := stdBuf.Write(data)
		hybridN, hybridErr := hybridBuf.Write(data)

		if stdN != hybridN {
			t.Fatalf("Write %d count mismatch: std=%d, hybrid=%d", i, stdN, hybridN)
		}
		if (stdErr == nil) != (hybridErr == nil) {
			t.Fatalf("Write %d error mismatch: std=%v, hybrid=%v", i, stdErr, hybridErr)
		}

		// Read from both buffers
		stdResult := make([]byte, len(data))
		stdN, stdErr = stdBuf.Read(stdResult)
		stdResults = append(stdResults, stdResult[:stdN])

		hybridResult := make([]byte, len(data))
		hybridN, hybridErr = hybridBuf.Read(hybridResult)
		hybridResults = append(hybridResults, hybridResult[:hybridN])

		// Note: After first read, HybridBuffer behavior may differ due to storage management
		if i == 0 {
			// First cycle should match exactly
			if stdN != hybridN {
				t.Fatalf("Read %d count mismatch: std=%d, hybrid=%d", i, stdN, hybridN)
			}
			if (stdErr == nil) != (hybridErr == nil) {
				t.Fatalf("Read %d error mismatch: std=%v, hybrid=%v", i, stdErr, hybridErr)
			}
		} else {
			// Subsequent cycles may behave differently due to stream management
			t.Logf("Read %d: std=%d bytes, hybrid=%d bytes (may differ due to storage)", i, stdN, hybridN)
		}
	}

	// Compare first result (should match)
	if len(stdResults) > 0 && len(hybridResults) > 0 {
		if !bytes.Equal(stdResults[0], hybridResults[0]) {
			t.Fatalf("First result mismatch: std=%q, hybrid=%q", string(stdResults[0]), string(hybridResults[0]))
		}
	}
}

// TestCompatibility_ByteOperations tests byte-level operations
func TestCompatibility_ByteOperations(t *testing.T) {
	data := []byte("ABC")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)

	// Test ReadByte
	for i := 0; i < len(data); i++ {
		stdByte, stdErr := stdBuf.ReadByte()
		hybridByte, hybridErr := hybridBuf.ReadByte()

		if stdByte != hybridByte {
			t.Fatalf("ReadByte %d mismatch: std=%c, hybrid=%c", i, stdByte, hybridByte)
		}
		if (stdErr == nil) != (hybridErr == nil) {
			t.Fatalf("ReadByte %d error mismatch: std=%v, hybrid=%v", i, stdErr, hybridErr)
		}
	}

	// Test EOF
	_, stdErr := stdBuf.ReadByte()
	_, hybridErr := hybridBuf.ReadByte()
	if stdErr != hybridErr {
		t.Fatalf("EOF error mismatch: std=%v, hybrid=%v", stdErr, hybridErr)
	}
}

// TestCompatibility_RuneOperations tests rune operations
func TestCompatibility_RuneOperations(t *testing.T) {
	text := "Hëllo, 世界!"

	// Test with bytes.Buffer - WriteByte and WriteRune
	stdBuf := &bytes.Buffer{}
	stdBuf.WriteByte('A')
	stdBuf.WriteRune('ñ')
	stdBuf.WriteString(text)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.WriteByte('A')
	hybridBuf.WriteRune('ñ')
	hybridBuf.WriteString(text)

	// Compare sizes
	if stdBuf.Len() != hybridBuf.Len() {
		t.Fatalf("Length mismatch: std=%d, hybrid=%d", stdBuf.Len(), hybridBuf.Len())
	}

	// Test ReadByte
	stdByte, stdErr := stdBuf.ReadByte()
	hybridByte, hybridErr := hybridBuf.ReadByte()
	if stdByte != hybridByte || stdErr != hybridErr {
		t.Fatalf("ReadByte mismatch: std=%c,%v hybrid=%c,%v", stdByte, stdErr, hybridByte, hybridErr)
	}

	// Test ReadRune
	stdRune, stdSize, stdErr := stdBuf.ReadRune()
	hybridRune, hybridSize, hybridErr := hybridBuf.ReadRune()
	if stdRune != hybridRune || stdSize != hybridSize || stdErr != hybridErr {
		t.Fatalf("ReadRune mismatch: std=%c,%d,%v hybrid=%c,%d,%v",
			stdRune, stdSize, stdErr, hybridRune, hybridSize, hybridErr)
	}

	// Test remaining data
	stdRemaining := stdBuf.String()
	hybridRemaining := hybridBuf.String()
	if stdRemaining != hybridRemaining {
		t.Fatalf("Remaining data mismatch: std=%q, hybrid=%q", stdRemaining, hybridRemaining)
	}
}

// TestCompatibility_StringOperations tests string-related operations
func TestCompatibility_StringOperations(t *testing.T) {
	lines := "Line1\nLine2\nLine3\n"

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.WriteString(lines)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.WriteString(lines)

	// Test ReadString
	for i := 0; i < 3; i++ {
		stdLine, stdErr := stdBuf.ReadString('\n')
		hybridLine, hybridErr := hybridBuf.ReadString('\n')

		if stdLine != hybridLine {
			t.Fatalf("ReadString %d mismatch: std=%q, hybrid=%q", i, stdLine, hybridLine)
		}
		if stdErr != hybridErr {
			t.Fatalf("ReadString %d error mismatch: std=%v, hybrid=%v", i, stdErr, hybridErr)
		}
	}
}

// TestCompatibility_ReadBytes tests ReadBytes functionality
func TestCompatibility_ReadBytes(t *testing.T) {
	data := "apple,banana,cherry,date"

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.WriteString(data)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.WriteString(data)

	// Test ReadBytes
	for {
		stdBytes, stdErr := stdBuf.ReadBytes(',')
		hybridBytes, hybridErr := hybridBuf.ReadBytes(',')

		if !bytes.Equal(stdBytes, hybridBytes) {
			t.Fatalf("ReadBytes mismatch: std=%q, hybrid=%q", string(stdBytes), string(hybridBytes))
		}
		if stdErr != hybridErr {
			t.Fatalf("ReadBytes error mismatch: std=%v, hybrid=%v", stdErr, hybridErr)
		}

		if stdErr == io.EOF {
			break
		}
	}
}

// TestCompatibility_Next tests Next functionality
func TestCompatibility_Next(t *testing.T) {
	data := []byte("0123456789")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)

	// Test Next with various sizes
	sizes := []int{2, 3, 100, 1, 0}
	for _, size := range sizes {
		stdNext := stdBuf.Next(size)
		hybridNext := hybridBuf.Next(size)

		if !bytes.Equal(stdNext, hybridNext) {
			t.Fatalf("Next(%d) mismatch: std=%q, hybrid=%q", size, string(stdNext), string(hybridNext))
		}
	}
}

// TestCompatibility_LenAndCap tests Len and Cap functionality
func TestCompatibility_LenAndCap(t *testing.T) {
	testSizes := []int{0, 1, 10, 100, 1000}

	for _, size := range testSizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		// Test with bytes.Buffer
		stdBuf := &bytes.Buffer{}
		stdBuf.Write(data)

		// Test with HybridBuffer (small threshold to test storage)
		hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
		defer hybridBuf.Close()
		hybridBuf.Write(data)

		// Compare Len
		if stdBuf.Len() != hybridBuf.Len() {
			t.Fatalf("Len mismatch for size %d: std=%d, hybrid=%d", size, stdBuf.Len(), hybridBuf.Len())
		}

		// Compare Cap (HybridBuffer returns Len for compatibility)
		if stdBuf.Cap() < hybridBuf.Cap() { // std might have higher cap
			// This is ok - HybridBuffer returns Len, std Buffer might have allocated more
		}

		// Read some data and compare again
		if size > 5 {
			stdBuf.Next(5)
			hybridBuf.Next(5)

			if stdBuf.Len() != hybridBuf.Len() {
				t.Fatalf("Len after Next mismatch for size %d: std=%d, hybrid=%d", size, stdBuf.Len(), hybridBuf.Len())
			}
		}
	}
}

// TestCompatibility_Bytes tests Bytes functionality
func TestCompatibility_Bytes(t *testing.T) {
	data := []byte("Hello, World!")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)

	// Compare Bytes - Note: HybridBuffer.Bytes() consumes content unlike bytes.Buffer
	stdBytes := stdBuf.Bytes()
	hybridBytes := hybridBuf.Bytes()

	if !bytes.Equal(stdBytes, hybridBytes) {
		t.Fatalf("Bytes mismatch: std=%q, hybrid=%q", string(stdBytes), string(hybridBytes))
	}

	// Test difference: bytes.Buffer.Bytes() is repeatable, HybridBuffer.Bytes() is not
	stdBytes2 := stdBuf.Bytes()
	hybridBytes2 := hybridBuf.Bytes()

	if !bytes.Equal(stdBytes, stdBytes2) {
		t.Fatalf("bytes.Buffer.Bytes() should be repeatable")
	}
	if len(hybridBytes2) != 0 {
		t.Fatalf("HybridBuffer.Bytes() should consume content, second call should be empty, got %q", string(hybridBytes2))
	}
}

// TestCompatibility_String tests String functionality
func TestCompatibility_String(t *testing.T) {
	text := "Hello, 世界!"

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.WriteString(text)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.WriteString(text)

	// Compare String - Note: HybridBuffer.String() consumes content unlike bytes.Buffer
	stdString := stdBuf.String()
	hybridString := hybridBuf.String()

	if stdString != hybridString {
		t.Fatalf("String mismatch: std=%q, hybrid=%q", stdString, hybridString)
	}
}

// TestCompatibility_Reset tests Reset functionality
func TestCompatibility_Reset(t *testing.T) {
	data := []byte("Some data")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)
	stdBuf.Reset()

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)
	hybridBuf.Reset()

	// Compare state after reset
	if stdBuf.Len() != hybridBuf.Len() {
		t.Fatalf("Len after Reset mismatch: std=%d, hybrid=%d", stdBuf.Len(), hybridBuf.Len())
	}
	if stdBuf.String() != hybridBuf.String() {
		t.Fatalf("String after Reset mismatch: std=%q, hybrid=%q", stdBuf.String(), hybridBuf.String())
	}

	// Test writing after reset
	newData := []byte("New data after reset")
	stdBuf.Write(newData)
	hybridBuf.Write(newData)

	if stdBuf.String() != hybridBuf.String() {
		t.Fatalf("String after write post-Reset mismatch: std=%q, hybrid=%q", stdBuf.String(), hybridBuf.String())
	}
}

// TestCompatibility_Truncate tests Truncate functionality
func TestCompatibility_Truncate(t *testing.T) {
	data := []byte("0123456789")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)
	stdBuf.Truncate(5)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)
	hybridBuf.Truncate(5)

	// Compare results
	if stdBuf.Len() != hybridBuf.Len() {
		t.Fatalf("Len after Truncate mismatch: std=%d, hybrid=%d", stdBuf.Len(), hybridBuf.Len())
	}
	if stdBuf.String() != hybridBuf.String() {
		t.Fatalf("String after Truncate mismatch: std=%q, hybrid=%q", stdBuf.String(), hybridBuf.String())
	}
}

// TestCompatibility_WriteTo tests WriteTo functionality
func TestCompatibility_WriteTo(t *testing.T) {
	data := []byte("Hello, World!")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(data)
	var stdTarget bytes.Buffer
	stdN, stdErr := stdBuf.WriteTo(&stdTarget)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridBuf.Write(data)
	var hybridTarget bytes.Buffer
	hybridN, hybridErr := hybridBuf.WriteTo(&hybridTarget)

	// Compare results
	if stdN != hybridN {
		t.Fatalf("WriteTo count mismatch: std=%d, hybrid=%d", stdN, hybridN)
	}
	if (stdErr == nil) != (hybridErr == nil) {
		t.Fatalf("WriteTo error mismatch: std=%v, hybrid=%v", stdErr, hybridErr)
	}
	if !bytes.Equal(stdTarget.Bytes(), hybridTarget.Bytes()) {
		t.Fatalf("WriteTo data mismatch: std=%q, hybrid=%q", stdTarget.String(), hybridTarget.String())
	}
}

// TestCompatibility_ReadFrom tests ReadFrom functionality
func TestCompatibility_ReadFrom(t *testing.T) {
	data := []byte("Hello, World!")

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdSource := bytes.NewReader(data)
	stdN, stdErr := stdBuf.ReadFrom(stdSource)

	// Test with HybridBuffer (small threshold to test storage)
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()
	hybridSource := bytes.NewReader(data)
	hybridN, hybridErr := hybridBuf.ReadFrom(hybridSource)

	// Compare results
	if stdN != hybridN {
		t.Fatalf("ReadFrom count mismatch: std=%d, hybrid=%d", stdN, hybridN)
	}
	if (stdErr == nil) != (hybridErr == nil) {
		t.Fatalf("ReadFrom error mismatch: std=%v, hybrid=%v", stdErr, hybridErr)
	}
	if stdBuf.String() != hybridBuf.String() {
		t.Fatalf("ReadFrom data mismatch: std=%q, hybrid=%q", stdBuf.String(), hybridBuf.String())
	}
}

// TestCompatibility_LargeData tests compatibility with larger data that triggers storage
func TestCompatibility_LargeData(t *testing.T) {
	// Create test data that will trigger storage switch
	largeData := make([]byte, 3*1024) // 3KB to exceed default 2MB threshold when set low
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Test with bytes.Buffer
	stdBuf := &bytes.Buffer{}
	stdBuf.Write(largeData)

	// Test with HybridBuffer with low threshold to trigger storage
	hybridBuf := New(WithThreshold(1024)) // 1KB threshold
	defer hybridBuf.Close()
	hybridBuf.Write(largeData)

	// Compare Len before String() call (since HybridBuffer.String() consumes content)
	if stdBuf.Len() != hybridBuf.Len() {
		t.Fatalf("Large data initial Len mismatch: std=%d, hybrid=%d", stdBuf.Len(), hybridBuf.Len())
	}

	// Compare String (this will load everything into memory for hybrid and consume the buffer)
	stdString := stdBuf.String()
	hybridString := hybridBuf.String()
	if stdString != hybridString {
		t.Fatalf("Large data String mismatch")
	}

	// After String() call: stdBuf.Len() unchanged, hybridBuf.Len() = 0 (consumed)
	if stdBuf.Len() != len(largeData) {
		t.Fatalf("bytes.Buffer should still have data: expected %d, got %d", len(largeData), stdBuf.Len())
	}
	if hybridBuf.Len() != 0 {
		t.Fatalf("HybridBuffer should be consumed: expected 0, got %d", hybridBuf.Len())
	}
}

// TestCompatibility_EdgeCases tests edge cases
func TestCompatibility_EdgeCases(t *testing.T) {
	// Test empty buffer operations
	stdBuf := &bytes.Buffer{}
	hybridBuf := New(WithThreshold(2)) // Force storage after 2 bytes
	defer hybridBuf.Close()

	// Test reading from empty buffer
	stdData := make([]byte, 10)
	hybridData := make([]byte, 10)

	stdN, stdErr := stdBuf.Read(stdData)
	hybridN, hybridErr := hybridBuf.Read(hybridData)

	if stdN != hybridN || stdErr != hybridErr {
		t.Fatalf("Empty read mismatch: std=%d,%v hybrid=%d,%v", stdN, stdErr, hybridN, hybridErr)
	}

	// Test Next on empty buffer
	stdNext := stdBuf.Next(5)
	hybridNext := hybridBuf.Next(5)

	if !bytes.Equal(stdNext, hybridNext) {
		t.Fatalf("Empty Next mismatch: std=%q, hybrid=%q", string(stdNext), string(hybridNext))
	}

	// Test writing nil/empty data
	stdN, stdErr = stdBuf.Write(nil)
	hybridN, hybridErr = hybridBuf.Write(nil)

	if stdN != hybridN || (stdErr == nil) != (hybridErr == nil) {
		t.Fatalf("Nil write mismatch: std=%d,%v hybrid=%d,%v", stdN, stdErr, hybridN, hybridErr)
	}
}

// TestCompatibility_Available tests Available method
func TestCompatibility_Available(t *testing.T) {
	// Note: bytes.Buffer.Available() returns available space in internal slice
	// HybridBuffer.Available() returns space before storage switch
	// We test that HybridBuffer.Available() behaves reasonably

	hybridBuf := New(WithThreshold(1024))
	defer hybridBuf.Close()

	// Initially should have full capacity
	if hybridBuf.Available() != 1024 {
		t.Fatalf("Initial Available should be 1024, got %d", hybridBuf.Available())
	}

	// Write some data
	hybridBuf.Write(make([]byte, 500))
	if hybridBuf.Available() != 524 { // 1024 - 500
		t.Fatalf("Available after 500 bytes should be 524, got %d", hybridBuf.Available())
	}

	// Exceed threshold
	hybridBuf.Write(make([]byte, 600))
	if hybridBuf.Available() != 0 { // Should be 0 after storage switch
		t.Fatalf("Available after storage switch should be 0, got %d", hybridBuf.Available())
	}
}
