# HybridBuffer - Go Hybrid Memory/Disk Buffer

![HybridBuffer Gopher Cyborg](gophercyborg.png)

A modular and efficient Go package for hybrid memory/disk buffering with pluggable middleware and storage backends. When memory usage exceeds a configurable threshold, data is automatically moved to storage, allowing you to handle large data streams without consuming excessive memory.

## 🚀 Features

### Smart Memory Management
- **Automatic Overflow**: Seamlessly switches from memory to disk when threshold is reached
- **Configurable Thresholds**: Set memory limits that fit your application's needs
- **Simple Logic**: When threshold is exceeded, ALL data moves to storage (no complex dual-source reading)
- **Pre-allocation**: Optimize performance with configurable memory pre-allocation

### Modular Architecture
- ✅ **Pluggable Middleware** - Encryption, compression, custom processing
- ✅ **Storage Backends** - Filesystem, S3, Redis, custom implementations
- ✅ **Separate Modules** - Import only what you need
- ✅ **Clean Interfaces** - Standard io.Reader/Writer throughout

### Core Functionality
- ✅ **Read/Write Operations** - Standard io.ReadWriter interface
- ✅ **Sequential Access** - Efficient streaming operations
- ✅ **Buffer Management** - Truncate, Reset, Grow operations
- ✅ **bytes.Buffer Compatibility** - Drop-in replacement for most use cases

## 📦 Installation

### Core Package
```bash
go get schneider.vip/hybridbuffer
```

### Optional Modules
```bash
# Encryption middleware
go get schneider.vip/hybridbuffer/middleware/encryption

# Compression middleware (high-performance)
go get schneider.vip/hybridbuffer/middleware/compression

# Compression middleware (stdlib-based)
go get schneider.vip/hybridbuffer/middleware/compressionstdlib

# Storage backends
go get schneider.vip/hybridbuffer/storage/filesystem  # Built-in default
go get schneider.vip/hybridbuffer/storage/s3         # AWS S3
go get schneider.vip/hybridbuffer/storage/redis      # Redis
```

## 🎯 Quick Start

### Basic Usage
```go
import "schneider.vip/hybridbuffer"

// Create standard buffer (uses filesystem storage by default)
buf := hybridbuffer.New()
defer buf.Close()

// Write data
buf.WriteString("Hello, World!")

// Read data back
data := make([]byte, 13)
n, err := buf.Read(data)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Read %d bytes: %s\n", n, string(data[:n]))
```

### With Middleware and Custom Storage
```go
import (
    "context"
    
    "schneider.vip/hybridbuffer"
    "schneider.vip/hybridbuffer/middleware/encryption"
    "schneider.vip/hybridbuffer/middleware/compression"
    "schneider.vip/hybridbuffer/storage/s3"
    
    awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/config"
)

// Create S3 client (example)
cfg, _ := config.LoadDefaultConfig(context.TODO())
s3Client := awss3.NewFromConfig(cfg)

// Create buffer with compression, encryption, and S3 storage
buf := hybridbuffer.New(
    hybridbuffer.WithThreshold(1024*1024),                    // 1MB memory threshold
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd), encryption.New()), // Multiple middlewares
    hybridbuffer.WithStorage(s3.New(s3Client, "my-bucket")), // S3 storage
)
defer buf.Close()
```

### Custom Encryption Key
```go
import "schneider.vip/hybridbuffer/middleware/encryption"

// With custom encryption key
key := make([]byte, 32)
// ... fill key with secure data

buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(encryption.WithKey(key))),
)
defer buf.Close()
```

### Multiple Middleware Usage
```go
import (
    "schneider.vip/hybridbuffer"
    "schneider.vip/hybridbuffer/middleware/encryption"
    "schneider.vip/hybridbuffer/middleware/compression"
)

// Single middleware
buf1 := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New()),
)

// Multiple middlewares in one call (recommended)
buf2 := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd), encryption.New()),
)

// Multiple middlewares in separate calls (also supported)
buf3 := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd)),
    hybridbuffer.WithMiddleware(encryption.New()),
)
```

## 🏗️ Architecture

```
┌─────────────────────────────────────┐
│       Buffer Interface              │  <- Clean external API
├─────────────────────────────────────┤
│         hybridBuffer                │  <- Core logic with middleware pipeline
├─────────────────────────────────────┤
│         Middleware                  │  <- Pluggable: encryption, compression, etc.
│    ┌─────────────────────────────┐   │
│    │  encryption  │  compression │   │
│    └─────────────────────────────┘   │
├─────────────────────────────────────┤
│       Storage Backend               │  <- Pluggable: filesystem, S3, Redis, etc.
│    ┌─────────────────────────────┐   │
│    │ filesystem │ s3 │  redis   │   │
│    └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Middleware Pipeline
- **Writing**: Applied in reverse order (last middleware first)
- **Reading**: Applied in forward order (first middleware first)
- **Example**: `Data → Compression → Encryption → Storage` (writing)
- **Example**: `Storage → Encryption → Compression → Data` (reading)

## 🔌 Available Modules

### Middleware

#### Encryption (`schneider.vip/hybridbuffer/middleware/encryption`)
```go
// Auto-generated key with AES-256-GCM (default)
encMiddleware := encryption.New()

// Custom key with AES-256-GCM
encMiddleware := encryption.New(encryption.WithKey(key))

// ChaCha20-Poly1305 cipher (better performance on systems without AES hardware)
encMiddleware := encryption.New(encryption.WithCipher(encryption.ChaCha20Poly1305))

// Custom key with ChaCha20-Poly1305
encMiddleware := encryption.New(
    encryption.WithKey(key),
    encryption.WithCipher(encryption.ChaCha20Poly1305),
)
```

#### Compression (High-Performance - Recommended)
**`schneider.vip/hybridbuffer/middleware/compression`**

Uses `klauspost/compress` for superior performance and more algorithms:

```go
// Zstd compression (recommended - best balance of speed and compression)
zstdMiddleware := compression.New(compression.Zstd)

// S2 compression (fastest compression/decompression)
s2Middleware := compression.New(compression.S2)

// Snappy compression (very fast, Google's algorithm)
snappyMiddleware := compression.New(compression.Snappy)

// Gzip compression (faster than stdlib)
gzipMiddleware := compression.New(compression.Gzip)

// With compression levels (Fastest, Default, Better, Best)
bestMiddleware := compression.New(compression.Zstd, compression.WithLevel(compression.Best))
fastMiddleware := compression.New(compression.S2, compression.WithLevel(compression.Fastest))
```

#### Compression (Standard Library)
**`schneider.vip/hybridbuffer/middleware/compressionstdlib`**

Uses Go's standard library for basic compression needs:

```go
// Gzip compression (stdlib)
gzipMiddleware := compressionstdlib.New(compressionstdlib.Gzip)

// Zlib compression (stdlib)
zlibMiddleware := compressionstdlib.New(compressionstdlib.Zlib, compressionstdlib.WithLevel(9))
```

**Recommendation**: Use the high-performance `compression` module for better performance and more algorithm choices.

### Storage Backends

#### Filesystem (`schneider.vip/hybridbuffer/storage/filesystem`)
```go
// Default options
fsStorage := filesystem.New()

// Custom options
fsStorage := filesystem.New(
    filesystem.WithTempDir("/custom/temp"),
    filesystem.WithPrefix("myapp"),
)
```

#### S3 (`schneider.vip/hybridbuffer/storage/s3`)
```go
s3Storage := s3.New(s3Client, "bucket-name")

// With options
s3Storage := s3.New(s3Client, "bucket-name",
    s3.WithKeyPrefix("myapp/temp"),
    s3.WithTimeout(60*time.Second),
)
```

#### Redis (`schneider.vip/hybridbuffer/storage/redis`)
```go
redisStorage := redis.New(redisClient)

// With options
redisStorage := redis.New(redisClient,
    redis.WithKeyPrefix("myapp:buffers"),
    redis.WithExpiration(24*time.Hour),
)
```

## 🎨 API Reference

### Core Options
```go
// Memory management
hybridbuffer.WithThreshold(size int)    // Memory threshold before storage
hybridbuffer.WithPreAlloc(size int)     // Pre-allocate memory buffer

// Middleware and storage
hybridbuffer.WithMiddleware(middlewares ...middleware.Middleware)  // Add one or more middlewares
hybridbuffer.WithStorage(provider func() storage.Backend)  // Set storage backend
```

### Buffer Interface
```go
type Buffer interface {
    // Full io.* interface support
    io.ReadWriter
    io.WriterTo
    io.ReaderFrom
    io.ByteReader
    io.ByteWriter
    io.StringWriter
    
    // bytes.Buffer-compatible methods
    ReadBytes(delim byte) ([]byte, error)
    ReadString(delim byte) (string, error)
    ReadRune() (rune, int, error)
    WriteRune(r rune) (int, error)
    Next(n int) []byte
    
    // Buffer management
    Len() int                    // Unread bytes
    Cap() int                    // Capacity (= Len)
    Available() int              // Available capacity before storage switch
    Size() int64                 // Total size
    Reset()                      // Clear buffer
    Close() error                // Clean up resources
    
    // Data access (WARNING: These CONSUME the buffer content!)
    Bytes() []byte               // Get remaining data as bytes (consumes content)
    String() string              // Get remaining data as string (consumes content)
    
    // Buffer manipulation
    Truncate(n int)              // Reduce size
    Grow(n int)                  // Expand memory buffer
}
```

### Constructors
```go
// Main constructor with functional options
hybridbuffer.New(opts ...Option) Buffer

// With initial data
hybridbuffer.NewFromBytes(data, opts ...Option) Buffer
hybridbuffer.NewFromString("Hello", opts ...Option) Buffer
```

## 🔒 Security Features

### Encryption
```go
import "schneider.vip/hybridbuffer/middleware/encryption"

// Automatic key generation (recommended)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New()),
)

// Custom key
key := make([]byte, 32)
// ... fill with secure random data
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(encryption.WithKey(key))),
)
```

**Encryption Details:**
- Uses MinIO SIO for authenticated encryption
- **AES-256-GCM** (default): Hardware accelerated on most systems
- **ChaCha20-Poly1305**: Better performance on systems without AES hardware
- Both ciphers provide tamper detection and authentication
- Memory data remains unencrypted for performance
- Only storage data is encrypted

### Compression

#### High-Performance Compression (Recommended)
```go
import "schneider.vip/hybridbuffer/middleware/compression"

// Zstd compression (recommended - best balance)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd)),
)

// S2 compression (fastest)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.S2)),
)

// Maximum compression with Zstd
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd, compression.WithLevel(compression.Best))),
)

// Snappy compression (very fast)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Snappy)),
)
```

#### Standard Library Compression
```go
import "schneider.vip/hybridbuffer/middleware/compressionstdlib"

// Basic Gzip compression
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compressionstdlib.New(compressionstdlib.Gzip)),
)
```

**Performance Comparison:**
- **High-Performance**: 2-3x faster, more algorithms (Zstd, S2, Snappy)
- **Standard Library**: Basic algorithms, no external dependencies

## 📊 Performance

### Benchmarks
```
BenchmarkHybridBuffer_Write-16               353005      3691 ns/op       5 B/op       0 allocs/op
BenchmarkHybridBuffer_ReadAt-16              58674091    20.66 ns/op      0 B/op       0 allocs/op
```

### Performance Tips
1. **Choose the right storage backend** for your use case
2. **Use high-performance compression** - `compression` module is 2-3x faster than `compressionstdlib`
3. **Algorithm selection**: S2 for speed, Zstd for balance, Snappy for very fast
4. **Optimize thresholds** based on your data patterns
5. **Use streaming operations** instead of Bytes()/String() for large data

## 🧪 Examples

### Large File Processing with Encryption
```go
func processLargeFile(filename string) error {
    buf := hybridbuffer.New(
        hybridbuffer.WithThreshold(10*1024*1024),     // 10MB memory
        hybridbuffer.WithMiddleware(encryption.New()), // Encrypt storage
    )
    defer buf.Close()
    
    // Read large file
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Stream file to buffer
    _, err = buf.ReadFrom(file)
    if err != nil {
        return err
    }
    
    // Process without memory issues
    return processBuffer(buf)
}
```

### Multi-Stage Processing Pipeline
```go
func processWithPipeline(data []byte) ([]byte, error) {
    buf := hybridbuffer.New(
        hybridbuffer.WithThreshold(1024*1024),
        hybridbuffer.WithMiddleware(compression.New(compression.Zstd), encryption.New()), // Multiple middlewares
        hybridbuffer.WithStorage(s3.New(s3Client, "temp-bucket")),
    )
    defer buf.Close()
    
    // Write data (gets compressed then encrypted to S3)
    if _, err := buf.Write(data); err != nil {
        return nil, err
    }
    
    // Read back (gets decrypted then decompressed from S3)
    result, err := io.ReadAll(buf)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### Custom Storage Backend
```go
import (
    "io"
    "schneider.vip/hybridbuffer"
    "schneider.vip/hybridbuffer/storage"
)

// Implement the interface
type CustomStorage struct {
    // your fields
}

func (c *CustomStorage) Create() (io.WriteCloser, error) {
    // return writer for your storage
}

func (c *CustomStorage) Open() (io.ReadCloser, error) {
    // return reader for your storage  
}

func (c *CustomStorage) Remove() error {
    // cleanup your storage
}

// Create provider function
func NewCustomStorage() func() storage.Backend {
    return func() storage.Backend {
        return &CustomStorage{
            // initialize your fields
        }
    }
}

// Use it
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(NewCustomStorage()),
)
```

## 🔧 Storage Provider Pattern

All storage modules return `func() storage.Backend` directly for clean integration:

```go
// Filesystem storage with options
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(filesystem.New(
        filesystem.WithTempDir("/custom/temp"),
        filesystem.WithPrefix("myapp"),
    )),
)

// S3 storage
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(s3.New(s3Client, "bucket-name")),
)

// Redis storage  
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(redis.New(redisClient)),
)

// Custom function
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(func() storage.Backend {
        return &MyCustomBackend{}
    }),
)
```

## ⚠️ Important Notes

### Design Philosophy
- **Sequential operations**: Optimized for streaming large data
- **Modular design**: Import only what you need
- **Standard interfaces**: Uses io.Reader/Writer throughout
- **Simple storage transition**: All data moves to storage when threshold exceeded

### Key Behavioral Notes

1. **String() and Bytes() consume content**:
   - These methods advance the read position
   - Subsequent calls return different/empty results
   - Avoid with large buffers - use streaming operations instead

2. **Middleware pipeline order**:
   - Writing: reverse order (compression → encryption → storage)
   - Reading: forward order (storage → encryption → compression)

3. **Storage backend requirements**:
   - Must implement Create(), Open(), Remove()
   - Should handle concurrent access if needed
   - Error handling is important for reliability

### Best Practices

#### ✅ Recommended Patterns
```go
// 1. Use appropriate storage for your use case
localBuf := hybridbuffer.New(
    hybridbuffer.WithStorage(filesystem.New()),
)

cloudBuf := hybridbuffer.New(
    hybridbuffer.WithStorage(s3.New(client, "bucket")),
)

// 2. Combine middleware for security
secureBuf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd), encryption.New()), // Multiple middlewares
)

// 3. Stream large data
io.Copy(destination, buf) // Good
result := buf.String()    // Avoid for large data
```

#### ❌ Avoid These Patterns
```go
// DON'T use String()/Bytes() for large data
buf.Write(hugeData)
result := buf.String() // ❌ Loads everything into memory

// DON'T ignore errors
buf.Write(data) // ❌ Always check errors
```

## 🔬 Testing

```bash
# Test core package
go test -v

# Test all modules
go test -v ./...

# Test specific module
go test -v ./storage/filesystem
go test -v ./middleware/encryption

# Benchmarks
go test -bench=. -benchmem
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**HybridBuffer** - Modular memory management for Go applications. 🚀