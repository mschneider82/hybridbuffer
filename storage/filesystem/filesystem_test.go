package filesystem_test

import (
	"os"
	"testing"

	"schneider.vip/hybridbuffer/storage/filesystem"
)

func TestBackend_ErrorCases(t *testing.T) {
	// Test with invalid temp directory
	factory := filesystem.New(filesystem.WithTempDir("/nonexistent/path"))
	backend := factory()

	// If backend was created, test operations
	if backend != nil {
		defer backend.Remove()

		// Try to create file - this might fail
		writer, err := backend.Create()
		if err != nil {
			t.Logf("Create failed as expected: %v", err)
		} else {
			writer.Close()
		}
	}
}

func TestBackend_Operations(t *testing.T) {
	// Test successful creation and operations
	factory := filesystem.New(
		filesystem.WithTempDir(os.TempDir()),
		filesystem.WithPrefix("test"),
	)
	backend := factory()
	defer backend.Remove()

	// Test Create
	writer, err := backend.Create()
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Write some data
	testData := []byte("Hello, Storage!")
	n, err := writer.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("Expected to write %d bytes, got %d", len(testData), n)
	}

	// Close writer
	if err = writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Test Open
	reader, err := backend.Open()
	if err != nil {
		t.Fatalf("Failed to open reader: %v", err)
	}
	defer reader.Close()

	// Read data back
	readData := make([]byte, len(testData))
	n, err = reader.Read(readData)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("Expected to read %d bytes, got %d", len(testData), n)
	}

	if string(readData) != string(testData) {
		t.Fatalf("Data mismatch: expected %q, got %q", string(testData), string(readData))
	}
}

func TestBackend_Remove(t *testing.T) {
	factory := filesystem.New()
	backend := factory()

	// Create and write some data
	writer, err := backend.Create()
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	writer.Write([]byte("test"))
	writer.Close()

	// Remove should work
	err = backend.Remove()
	if err != nil {
		t.Logf("Remove error (may be expected): %v", err)
	}

	// Second remove should not panic
	err = backend.Remove()
	if err != nil {
		t.Logf("Second remove error (may be expected): %v", err)
	}
}

func TestBackend_OpenNonExistent(t *testing.T) {
	factory := filesystem.New()
	backend := factory()

	defer backend.Remove()

	// Try to open without creating first
	_, err := backend.Open()
	if err == nil {
		t.Fatal("Expected error when opening non-existent file")
	}
	t.Logf("Open non-existent file failed as expected: %v", err)
}

func TestFactory(t *testing.T) {
	// Test factory creation
	factory := filesystem.New(
		filesystem.WithTempDir("/tmp"),
		filesystem.WithPrefix("factory-test"),
	)

	// Test backend creation
	backend := factory()

	defer backend.Remove()

	// Test that backend works
	writer, err := backend.Create()
	if err != nil {
		t.Fatalf("Backend Create failed: %v", err)
	}
	defer writer.Close()

	writer.Write([]byte("factory test"))
}

func TestOptions(t *testing.T) {
	// Test default options
	factory1 := filesystem.New()
	backend1 := factory1()

	defer backend1.Remove()

	// Test custom temp dir
	factory2 := filesystem.New(filesystem.WithTempDir(os.TempDir()))
	backend2 := factory2()

	defer backend2.Remove()

	// Test custom prefix
	factory3 := filesystem.New(filesystem.WithPrefix("custom"))
	backend3 := factory3()

	defer backend3.Remove()

	// Verify the prefix is used (this is hard to test directly, but we can create and see if it works)
	writer, err := backend3.Create()
	if err != nil {
		t.Fatalf("Create failed with custom prefix: %v", err)
	}
	writer.Write([]byte("prefix test"))
	writer.Close()

	reader, err := backend3.Open()
	if err != nil {
		t.Fatalf("Open failed with custom prefix: %v", err)
	}
	defer reader.Close()

	data := make([]byte, 11)
	n, _ := reader.Read(data)
	if string(data[:n]) != "prefix test" {
		t.Fatalf("Data mismatch with custom prefix: got %q", string(data[:n]))
	}
}
