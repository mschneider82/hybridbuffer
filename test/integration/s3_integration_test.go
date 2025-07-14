package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"schneider.vip/hybridbuffer"
	s3storage "schneider.vip/hybridbuffer/storage/s3"
)

// mockS3Client for integration testing
type mockS3Client struct {
	objects map[string][]byte
}

func newMockS3Client() *mockS3Client {
	return &mockS3Client{
		objects: make(map[string][]byte),
	}
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	key := fmt.Sprintf("%s:%s", *params.Bucket, *params.Key)
	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	m.objects[key] = data
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := fmt.Sprintf("%s:%s", *params.Bucket, *params.Key)
	data, exists := m.objects[key]
	if !exists {
		return nil, fmt.Errorf("NoSuchKey: The specified key does not exist")
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(data)),
	}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	key := fmt.Sprintf("%s:%s", *params.Bucket, *params.Key)
	delete(m.objects, key)
	return &s3.DeleteObjectOutput{}, nil
}

func TestHybridBuffer_S3Integration(t *testing.T) {
	client := newMockS3Client()
	bucket := "test-hybridbuffer-bucket"

	// Create HybridBuffer with S3 storage
	buf := hybridbuffer.New(
		hybridbuffer.WithThreshold(50), // Very small threshold to force S3 usage
		hybridbuffer.WithStorage(s3storage.New(client, bucket,
			s3storage.WithKeyPrefix("integration-test"),
		)),
	)
	defer buf.Close()

	// Write data that exceeds threshold
	testData := "This is a test string that is longer than 50 bytes and should trigger S3 storage backend usage!"
	t.Logf("Writing %d bytes with threshold 50", len(testData))

	n, err := buf.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write to HybridBuffer with S3: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("Expected to write %d bytes, got %d", len(testData), n)
	}

	// Force flush to storage by writing more data
	extraData := " Additional data to ensure storage flush."
	buf.WriteString(extraData)
	testData += extraData

	// Read the data to trigger storage operations
	result := buf.String()

	// Verify data was stored in S3 mock
	if len(client.objects) == 0 {
		t.Fatal("No objects found in S3 mock - data was not stored")
	}

	// Verify the read data is correct
	if result != testData {
		t.Fatalf("Data mismatch: expected %q, got %q", testData, result)
	}

	// Verify S3 object was created with correct prefix
	found := false
	for key := range client.objects {
		if len(key) > len(bucket) && key[:len(bucket)] == bucket {
			objectKey := key[len(bucket)+1:] // Remove "bucket:" prefix
			if len(objectKey) > 16 && objectKey[:16] == "integration-test" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("S3 object with correct prefix not found")
	}
}

func TestHybridBuffer_S3LargeData(t *testing.T) {
	client := newMockS3Client()
	bucket := "test-hybridbuffer-bucket"

	// Create HybridBuffer with very small threshold
	buf := hybridbuffer.New(
		hybridbuffer.WithThreshold(50),
		hybridbuffer.WithStorage(s3storage.New(client, bucket)),
	)
	defer buf.Close()

	// Write large amount of data in chunks
	totalSize := 0
	for i := 0; i < 100; i++ {
		chunk := fmt.Sprintf("This is chunk number %d with some data. ", i)
		n, err := buf.WriteString(chunk)
		if err != nil {
			t.Fatalf("Failed to write chunk %d: %v", i, err)
		}
		totalSize += n
	}

	// Read all data back to trigger storage
	readData := buf.String()

	// Verify data was stored
	if len(client.objects) == 0 {
		t.Fatal("No objects found in S3 mock")
	}

	// Verify the read data size
	if len(readData) != totalSize {
		t.Fatalf("Expected to read %d bytes, got %d", totalSize, len(readData))
	}

	// Verify data contains expected chunks
	for i := 0; i < 100; i++ {
		expected := fmt.Sprintf("This is chunk number %d with some data. ", i)
		if !bytes.Contains([]byte(readData), []byte(expected)) {
			t.Fatalf("Chunk %d not found in read data", i)
		}
	}
}
