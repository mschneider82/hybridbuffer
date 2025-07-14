package compressionstdlib

import (
	"bytes"
	"io"
	"testing"
)

func TestNew_DefaultLevel(t *testing.T) {
	// Test default compression level
	m := New(Gzip)
	if m.level != 6 {
		t.Fatalf("Expected default level 6, got %d", m.level)
	}
}

func TestNew_CustomLevel(t *testing.T) {
	// Test custom compression level
	m := New(Gzip, WithLevel(9))
	if m.level != 9 {
		t.Fatalf("Expected level 9, got %d", m.level)
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	// Test invalid compression level (should use default)
	m := New(Gzip, WithLevel(15)) // Invalid level
	if m.level != 6 {
		t.Fatalf("Expected default level 6 for invalid input, got %d", m.level)
	}
}

func TestGzipCompression(t *testing.T) {
	testCompressionAlgorithm(t, Gzip, "Gzip")
}

func TestZlibCompression(t *testing.T) {
	testCompressionAlgorithm(t, Zlib, "Zlib")
}

func testCompressionAlgorithm(t *testing.T, algorithm Algorithm, name string) {
	m := New(algorithm)

	// Test data - something that compresses well
	testData := []byte("Hello, world! This is a test message that should compress well. " +
		"Hello, world! This is a test message that should compress well. " +
		"Hello, world! This is a test message that should compress well.")

	// Compress
	var compressedBuf bytes.Buffer
	compressWriter := m.Writer(&compressedBuf)
	
	n, err := compressWriter.Write(testData)
	if err != nil {
		t.Fatalf("%s: Failed to write compressed data: %v", name, err)
	}
	if n != len(testData) {
		t.Fatalf("%s: Expected to write %d bytes, got %d", name, len(testData), n)
	}

	// Close the writer to finalize compression
	if closer, ok := compressWriter.(io.Closer); ok {
		err = closer.Close()
		if err != nil {
			t.Fatalf("%s: Failed to close compressor: %v", name, err)
		}
	}

	// Verify data is actually compressed (should be smaller for this repetitive data)
	compressedData := compressedBuf.Bytes()
	if len(compressedData) >= len(testData) {
		t.Logf("%s: Compressed size %d >= original size %d (may be normal for small data)", 
			name, len(compressedData), len(testData))
	}

	// Decompress
	decompressReader := m.Reader(bytes.NewReader(compressedData))
	
	decompressedData, err := io.ReadAll(decompressReader)
	if err != nil {
		t.Fatalf("%s: Failed to read decompressed data: %v", name, err)
	}

	// Verify decompressed data matches original
	if !bytes.Equal(testData, decompressedData) {
		t.Fatalf("%s: Decompressed data doesn't match original: got %q, expected %q", 
			name, string(decompressedData), string(testData))
	}

	t.Logf("%s: Successfully compressed %d bytes to %d bytes (%.1f%% ratio)", 
		name, len(testData), len(compressedData), 
		float64(len(compressedData))/float64(len(testData))*100)
}

func TestCompressionLevels(t *testing.T) {
	// Test different compression levels
	testData := bytes.Repeat([]byte("This is a test string for compression. "), 100)
	
	for level := 1; level <= 9; level++ {
		m := New(Gzip, WithLevel(level))
		
		var compressedBuf bytes.Buffer
		compressWriter := m.Writer(&compressedBuf)
		
		compressWriter.Write(testData)
		if closer, ok := compressWriter.(io.Closer); ok {
			closer.Close()
		}
		
		// Decompress to verify
		decompressReader := m.Reader(bytes.NewReader(compressedBuf.Bytes()))
		decompressedData, err := io.ReadAll(decompressReader)
		if err != nil {
			t.Fatalf("Level %d: Failed to decompress: %v", level, err)
		}
		
		if !bytes.Equal(testData, decompressedData) {
			t.Fatalf("Level %d: Data mismatch", level)
		}
		
		t.Logf("Level %d: %d bytes -> %d bytes (%.1f%%)", 
			level, len(testData), compressedBuf.Len(),
			float64(compressedBuf.Len())/float64(len(testData))*100)
	}
}

func TestLargeData(t *testing.T) {
	// Test with larger data
	m := New(Gzip)

	// Create 100KB of test data
	testData := make([]byte, 100*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Compress
	var compressedBuf bytes.Buffer
	compressWriter := m.Writer(&compressedBuf)
	
	// Write in chunks to test streaming
	chunkSize := 4096
	for i := 0; i < len(testData); i += chunkSize {
		end := i + chunkSize
		if end > len(testData) {
			end = len(testData)
		}
		
		_, err := compressWriter.Write(testData[i:end])
		if err != nil {
			t.Fatalf("Failed to write chunk at %d: %v", i, err)
		}
	}
	
	if closer, ok := compressWriter.(io.Closer); ok {
		closer.Close()
	}

	// Decompress
	decompressReader := m.Reader(bytes.NewReader(compressedBuf.Bytes()))
	
	decompressedData, err := io.ReadAll(decompressReader)
	if err != nil {
		t.Fatalf("Failed to read all decompressed data: %v", err)
	}

	// Verify
	if !bytes.Equal(testData, decompressedData) {
		t.Fatal("Large data compression/decompression failed")
	}
	
	t.Logf("Large data: %d bytes -> %d bytes (%.1f%% ratio)", 
		len(testData), compressedBuf.Len(),
		float64(compressedBuf.Len())/float64(len(testData))*100)
}

func TestMultipleWrites(t *testing.T) {
	// Test multiple writes to the same compressed writer
	m := New(Gzip)

	testParts := [][]byte{
		[]byte("Part 1: "),
		[]byte("Part 2: "),
		[]byte("Part 3: "),
		[]byte("Part 4: "),
		[]byte("Final part."),
	}
	
	expectedData := bytes.Join(testParts, nil)

	// Compress with multiple writes
	var compressedBuf bytes.Buffer
	compressWriter := m.Writer(&compressedBuf)
	
	for _, part := range testParts {
		_, err := compressWriter.Write(part)
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}
	}
	
	if closer, ok := compressWriter.(io.Closer); ok {
		closer.Close()
	}

	// Decompress
	decompressReader := m.Reader(bytes.NewReader(compressedBuf.Bytes()))
	
	decompressedData, err := io.ReadAll(decompressReader)
	if err != nil {
		t.Fatalf("Failed to read decompressed data: %v", err)
	}

	// Verify
	if !bytes.Equal(expectedData, decompressedData) {
		t.Fatalf("Multiple writes test failed: got %q, expected %q", 
			string(decompressedData), string(expectedData))
	}
}

func TestUnsupportedAlgorithm(t *testing.T) {
	// Test panic with unsupported algorithm
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic with unsupported algorithm")
		}
	}()

	// This should panic
	m := New(Algorithm(999)) // Invalid algorithm
	m.Writer(&bytes.Buffer{})
}

func TestEmptyData(t *testing.T) {
	// Test compression of empty data
	m := New(Gzip)
	
	var compressedBuf bytes.Buffer
	compressWriter := m.Writer(&compressedBuf)
	
	// Write nothing
	if closer, ok := compressWriter.(io.Closer); ok {
		closer.Close()
	}
	
	// Decompress
	decompressReader := m.Reader(bytes.NewReader(compressedBuf.Bytes()))
	
	decompressedData, err := io.ReadAll(decompressReader)
	if err != nil {
		t.Fatalf("Failed to read decompressed empty data: %v", err)
	}
	
	if len(decompressedData) != 0 {
		t.Fatalf("Expected empty data, got %d bytes", len(decompressedData))
	}
}