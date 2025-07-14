package compression

import (
	"bytes"
	"io"
	"testing"
)

func TestNew_DefaultLevel(t *testing.T) {
	m := New(Gzip)
	if m.level != Default {
		t.Fatalf("Expected default level, got %v", m.level)
	}
}

func TestNew_CustomLevel(t *testing.T) {
	m := New(Gzip, WithLevel(Best))
	if m.level != Best {
		t.Fatalf("Expected Best level, got %v", m.level)
	}
}

func TestAllAlgorithms(t *testing.T) {
	algorithms := []struct {
		name      string
		algorithm Algorithm
	}{
		{"Gzip", Gzip},
		{"Zstd", Zstd},
		{"S2", S2},
		{"Snappy", Snappy},
		{"Zlib", Zlib},
		{"Flate", Flate},
	}

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			testCompressionAlgorithm(t, alg.algorithm, alg.name)
		})
	}
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

	// Verify data is compressed
	compressedData := compressedBuf.Bytes()
	t.Logf("%s: Original %d bytes -> Compressed %d bytes", name, len(testData), len(compressedData))

	// Decompress
	decompressReader := m.Reader(bytes.NewReader(compressedData))
	
	decompressedData, err := io.ReadAll(decompressReader)
	if err != nil {
		t.Fatalf("%s: Failed to read decompressed data: %v", name, err)
	}

	// Verify decompressed data matches original
	if !bytes.Equal(testData, decompressedData) {
		t.Fatalf("%s: Decompressed data doesn't match original", name)
	}

	// Close reader if it supports it
	if closer, ok := decompressReader.(io.Closer); ok {
		closer.Close()
	}
}

func TestCompressionLevels(t *testing.T) {
	levels := []struct {
		name  string
		level Level
	}{
		{"Fastest", Fastest},
		{"Default", Default},
		{"Better", Better},
		{"Best", Best},
	}

	// Test different compression levels with Zstd (supports all levels well)
	testData := bytes.Repeat([]byte("This is a test string for compression level testing. "), 50)
	
	for _, level := range levels {
		t.Run(level.name, func(t *testing.T) {
			m := New(Zstd, WithLevel(level.level))
			
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
				t.Fatalf("Level %s: Failed to decompress: %v", level.name, err)
			}
			
			if !bytes.Equal(testData, decompressedData) {
				t.Fatalf("Level %s: Data mismatch", level.name)
			}
			
			// Close reader if it supports it
			if closer, ok := decompressReader.(io.Closer); ok {
				closer.Close()
			}
			
			t.Logf("Level %s: %d bytes -> %d bytes (%.1f%%)", 
				level.name, len(testData), compressedBuf.Len(),
				float64(compressedBuf.Len())/float64(len(testData))*100)
		})
	}
}

func TestLargeData(t *testing.T) {
	// Test with larger data using fast S2 compression
	m := New(S2)

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
	
	// Close reader if it supports it
	if closer, ok := decompressReader.(io.Closer); ok {
		closer.Close()
	}
	
	t.Logf("Large data (S2): %d bytes -> %d bytes (%.1f%% ratio)", 
		len(testData), compressedBuf.Len(),
		float64(compressedBuf.Len())/float64(len(testData))*100)
}

func TestMultipleWrites(t *testing.T) {
	// Test multiple writes with Zstd
	m := New(Zstd)

	testParts := [][]byte{
		[]byte("Part 1: Hello "),
		[]byte("Part 2: World "),
		[]byte("Part 3: This "),
		[]byte("Part 4: Is "),
		[]byte("Part 5: A "),
		[]byte("Part 6: Test."),
	}
	
	expectedData := bytes.Join(testParts, nil)

	// Compress with multiple writes
	var compressedBuf bytes.Buffer
	compressWriter := m.Writer(&compressedBuf)
	
	for i, part := range testParts {
		_, err := compressWriter.Write(part)
		if err != nil {
			t.Fatalf("Failed to write part %d: %v", i, err)
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
	
	// Close reader if it supports it
	if closer, ok := decompressReader.(io.Closer); ok {
		closer.Close()
	}
}

func TestEmptyData(t *testing.T) {
	// Test compression of empty data with all algorithms
	algorithms := []Algorithm{Gzip, Zstd, S2, Snappy, Zlib, Flate}
	
	for _, alg := range algorithms {
		m := New(alg)
		
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
			t.Fatalf("Algorithm %d: Failed to read decompressed empty data: %v", alg, err)
		}
		
		if len(decompressedData) != 0 {
			t.Fatalf("Algorithm %d: Expected empty data, got %d bytes", alg, len(decompressedData))
		}
		
		// Close reader if it supports it
		if closer, ok := decompressReader.(io.Closer); ok {
			closer.Close()
		}
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

// Benchmark different algorithms
func BenchmarkCompression(b *testing.B) {
	// Create test data
	testData := bytes.Repeat([]byte("This is a benchmark test for compression performance. "), 1000)
	
	algorithms := []struct {
		name string
		alg  Algorithm
	}{
		{"Gzip", Gzip},
		{"Zstd", Zstd},
		{"S2", S2},
		{"Snappy", Snappy},
		{"Zlib", Zlib},
		{"Flate", Flate},
	}
	
	for _, alg := range algorithms {
		b.Run(alg.name, func(b *testing.B) {
			m := New(alg.alg)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var compressedBuf bytes.Buffer
				compressWriter := m.Writer(&compressedBuf)
				
				compressWriter.Write(testData)
				if closer, ok := compressWriter.(io.Closer); ok {
					closer.Close()
				}
				
				// Also test decompression
				decompressReader := m.Reader(bytes.NewReader(compressedBuf.Bytes()))
				io.ReadAll(decompressReader)
				if closer, ok := decompressReader.(io.Closer); ok {
					closer.Close()
				}
			}
		})
	}
}