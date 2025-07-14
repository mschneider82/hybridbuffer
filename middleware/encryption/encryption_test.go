package encryption_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"schneider.vip/hybridbuffer/middleware/encryption"
)

func TestWithKey_InvalidSize(t *testing.T) {
	// Test panic with invalid key size
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic with invalid key size")
		}
	}()

	// This should panic
	encryption.New(encryption.WithKey([]byte("too short")))
}

func TestDefaultCipher(t *testing.T) {
	// Test that default cipher is AES256GCM
	m := encryption.New()
	
	// We can't directly access the cipher field, but we can test that it works
	testData := []byte("Hello, encryption!")
	
	// Encrypt
	var encryptedBuf bytes.Buffer
	encryptWriter := m.Writer(&encryptedBuf)
	encryptWriter.Write(testData)
	if closer, ok := encryptWriter.(io.Closer); ok {
		closer.Close()
	}
	
	// Decrypt
	decryptReader := m.Reader(bytes.NewReader(encryptedBuf.Bytes()))
	decryptedData, err := io.ReadAll(decryptReader)
	if err != nil {
		t.Fatalf("Failed to decrypt with default cipher: %v", err)
	}
	
	if !bytes.Equal(testData, decryptedData) {
		t.Fatal("Default cipher encryption/decryption failed")
	}
}

func TestEncryptionDecryption(t *testing.T) {
	ciphers := []struct {
		name   string
		cipher encryption.Cipher
	}{
		{"AES256GCM", encryption.AES256GCM},
		{"ChaCha20Poly1305", encryption.ChaCha20Poly1305},
	}

	for _, cipher := range ciphers {
		t.Run(cipher.name, func(t *testing.T) {
			// Create middleware with known key and cipher
			key := make([]byte, 32)
			rand.Read(key)
			m := encryption.New(encryption.WithKey(key), encryption.WithCipher(cipher.cipher))

			// Test data
			testData := []byte("Hello, this is a test message for encryption!")

			// Encrypt
			var encryptedBuf bytes.Buffer
			encryptWriter := m.Writer(&encryptedBuf)

			n, err := encryptWriter.Write(testData)
			if err != nil {
				t.Fatalf("Failed to write encrypted data: %v", err)
			}
			if n != len(testData) {
				t.Fatalf("Expected to write %d bytes, got %d", len(testData), n)
			}

			// Close the writer to finalize encryption
			if closer, ok := encryptWriter.(io.Closer); ok {
				closer.Close()
			}

			// Verify data is actually encrypted (should be different)
			encryptedData := encryptedBuf.Bytes()
			if bytes.Equal(testData, encryptedData) {
				t.Fatal("Encrypted data should be different from original")
			}

			// Decrypt
			decryptReader := m.Reader(bytes.NewReader(encryptedData))

			decryptedData := make([]byte, len(testData))
			n, err = io.ReadFull(decryptReader, decryptedData)
			if err != nil {
				t.Fatalf("Failed to read decrypted data: %v", err)
			}
			if n != len(testData) {
				t.Fatalf("Expected to read %d bytes, got %d", len(testData), n)
			}

			// Verify decrypted data matches original
			if !bytes.Equal(testData, decryptedData) {
				t.Fatalf("Decrypted data doesn't match original: got %q, expected %q",
					string(decryptedData), string(testData))
			}

			t.Logf("Successfully encrypted/decrypted with %s", cipher.name)
		})
	}
}

func TestCipherCompatibility(t *testing.T) {
	// Test that different ciphers produce different encrypted data
	key := make([]byte, 32)
	rand.Read(key)
	
	testData := []byte("Hello, cipher compatibility test!")
	
	// Encrypt with AES256GCM
	m1 := encryption.New(encryption.WithKey(key), encryption.WithCipher(encryption.AES256GCM))
	var encryptedBuf1 bytes.Buffer
	encryptWriter1 := m1.Writer(&encryptedBuf1)
	encryptWriter1.Write(testData)
	if closer, ok := encryptWriter1.(io.Closer); ok {
		closer.Close()
	}
	
	// Encrypt with ChaCha20Poly1305
	m2 := encryption.New(encryption.WithKey(key), encryption.WithCipher(encryption.ChaCha20Poly1305))
	var encryptedBuf2 bytes.Buffer
	encryptWriter2 := m2.Writer(&encryptedBuf2)
	encryptWriter2.Write(testData)
	if closer, ok := encryptWriter2.(io.Closer); ok {
		closer.Close()
	}
	
	// Encrypted data should be different (different cipher formats)
	if bytes.Equal(encryptedBuf1.Bytes(), encryptedBuf2.Bytes()) {
		t.Fatal("Different ciphers should produce different encrypted data")
	}
	
	// But each should decrypt correctly with its own cipher
	decryptReader1 := m1.Reader(bytes.NewReader(encryptedBuf1.Bytes()))
	decryptedData1, err := io.ReadAll(decryptReader1)
	if err != nil {
		t.Fatalf("Failed to decrypt AES256GCM data: %v", err)
	}
	
	decryptReader2 := m2.Reader(bytes.NewReader(encryptedBuf2.Bytes()))
	decryptedData2, err := io.ReadAll(decryptReader2)
	if err != nil {
		t.Fatalf("Failed to decrypt ChaCha20Poly1305 data: %v", err)
	}
	
	if !bytes.Equal(testData, decryptedData1) {
		t.Fatal("AES256GCM decryption failed")
	}
	if !bytes.Equal(testData, decryptedData2) {
		t.Fatal("ChaCha20Poly1305 decryption failed")
	}
	
	t.Logf("AES256GCM encrypted size: %d bytes", len(encryptedBuf1.Bytes()))
	t.Logf("ChaCha20Poly1305 encrypted size: %d bytes", len(encryptedBuf2.Bytes()))
}

func TestEncryptionWithDifferentKeys(t *testing.T) {
	// Create two middlewares with different keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	m1 := encryption.New(encryption.WithKey(key1))
	m2 := encryption.New(encryption.WithKey(key2))

	testData := []byte("Secret message")

	// Encrypt with first key
	var encryptedBuf bytes.Buffer
	encryptWriter := m1.Writer(&encryptedBuf)
	encryptWriter.Write(testData)
	if closer, ok := encryptWriter.(io.Closer); ok {
		closer.Close()
	}

	// Try to decrypt with second key (should fail)
	decryptReader := m2.Reader(bytes.NewReader(encryptedBuf.Bytes()))

	decryptedData := make([]byte, len(testData))
	_, err := io.ReadFull(decryptReader, decryptedData)
	if err == nil {
		t.Fatal("Expected error when decrypting with wrong key")
	}
	t.Logf("Correctly failed to decrypt with wrong key: %v", err)
}

func TestLargeData(t *testing.T) {
	// Test with larger data to ensure streaming works
	ciphers := []struct {
		name   string
		cipher encryption.Cipher
	}{
		{"AES256GCM", encryption.AES256GCM},
		{"ChaCha20Poly1305", encryption.ChaCha20Poly1305},
	}

	for _, cipher := range ciphers {
		t.Run(cipher.name, func(t *testing.T) {
			m := encryption.New(encryption.WithCipher(cipher.cipher))

			// Create 1MB of test data
			testData := make([]byte, 1024*1024)
			for i := range testData {
				testData[i] = byte(i % 256)
			}

			// Encrypt
			var encryptedBuf bytes.Buffer
			encryptWriter := m.Writer(&encryptedBuf)

			// Write in chunks to test streaming
			chunkSize := 4096
			for i := 0; i < len(testData); i += chunkSize {
				end := i + chunkSize
				if end > len(testData) {
					end = len(testData)
				}

				_, err := encryptWriter.Write(testData[i:end])
				if err != nil {
					t.Fatalf("Failed to write chunk at %d: %v", i, err)
				}
			}

			if closer, ok := encryptWriter.(io.Closer); ok {
				closer.Close()
			}

			// Decrypt
			decryptReader := m.Reader(bytes.NewReader(encryptedBuf.Bytes()))

			decryptedData, err := io.ReadAll(decryptReader)
			if err != nil {
				t.Fatalf("Failed to read all decrypted data: %v", err)
			}

			// Verify
			if !bytes.Equal(testData, decryptedData) {
				t.Fatal("Large data encryption/decryption failed")
			}

			t.Logf("Successfully encrypted/decrypted %d bytes with %s", len(testData), cipher.name)
		})
	}
}

func TestMultipleWrites(t *testing.T) {
	// Test multiple writes to the same encrypted writer
	ciphers := []struct {
		name   string
		cipher encryption.Cipher
	}{
		{"AES256GCM", encryption.AES256GCM},
		{"ChaCha20Poly1305", encryption.ChaCha20Poly1305},
	}

	for _, cipher := range ciphers {
		t.Run(cipher.name, func(t *testing.T) {
			m := encryption.New(encryption.WithCipher(cipher.cipher))

			testParts := [][]byte{
				[]byte("Hello "),
				[]byte("world! "),
				[]byte("This "),
				[]byte("is "),
				[]byte("a "),
				[]byte("test."),
			}

			expectedData := bytes.Join(testParts, nil)

			// Encrypt with multiple writes
			var encryptedBuf bytes.Buffer
			encryptWriter := m.Writer(&encryptedBuf)

			for _, part := range testParts {
				_, err := encryptWriter.Write(part)
				if err != nil {
					t.Fatalf("Failed to write part: %v", err)
				}
			}

			if closer, ok := encryptWriter.(io.Closer); ok {
				closer.Close()
			}

			// Decrypt
			decryptReader := m.Reader(bytes.NewReader(encryptedBuf.Bytes()))

			decryptedData, err := io.ReadAll(decryptReader)
			if err != nil {
				t.Fatalf("Failed to read decrypted data: %v", err)
			}

			// Verify
			if !bytes.Equal(expectedData, decryptedData) {
				t.Fatalf("Multiple writes test failed: got %q, expected %q",
					string(decryptedData), string(expectedData))
			}
		})
	}
}

func TestCombinedOptions(t *testing.T) {
	// Test using both WithKey and WithCipher options
	key := make([]byte, 32)
	rand.Read(key)
	
	m := encryption.New(
		encryption.WithKey(key),
		encryption.WithCipher(encryption.ChaCha20Poly1305),
	)
	
	testData := []byte("Combined options test")
	
	// Encrypt
	var encryptedBuf bytes.Buffer
	encryptWriter := m.Writer(&encryptedBuf)
	encryptWriter.Write(testData)
	if closer, ok := encryptWriter.(io.Closer); ok {
		closer.Close()
	}
	
	// Decrypt
	decryptReader := m.Reader(bytes.NewReader(encryptedBuf.Bytes()))
	decryptedData, err := io.ReadAll(decryptReader)
	if err != nil {
		t.Fatalf("Failed to decrypt with combined options: %v", err)
	}
	
	if !bytes.Equal(testData, decryptedData) {
		t.Fatal("Combined options encryption/decryption failed")
	}
}

func TestEmptyData(t *testing.T) {
	// Test encryption of empty data
	ciphers := []struct {
		name   string
		cipher encryption.Cipher
	}{
		{"AES256GCM", encryption.AES256GCM},
		{"ChaCha20Poly1305", encryption.ChaCha20Poly1305},
	}

	for _, cipher := range ciphers {
		t.Run(cipher.name, func(t *testing.T) {
			m := encryption.New(encryption.WithCipher(cipher.cipher))
			
			// Test empty data
			testData := []byte{}
			
			// Encrypt
			var encryptedBuf bytes.Buffer
			encryptWriter := m.Writer(&encryptedBuf)
			encryptWriter.Write(testData)
			if closer, ok := encryptWriter.(io.Closer); ok {
				closer.Close()
			}
			
			// Decrypt
			decryptReader := m.Reader(bytes.NewReader(encryptedBuf.Bytes()))
			decryptedData, err := io.ReadAll(decryptReader)
			if err != nil {
				t.Fatalf("Failed to decrypt empty data: %v", err)
			}
			
			if !bytes.Equal(testData, decryptedData) {
				t.Fatalf("Empty data encryption/decryption failed: got %d bytes, expected %d", 
					len(decryptedData), len(testData))
			}
		})
	}
}