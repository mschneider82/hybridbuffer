# Compression Middleware

Advanced compression middleware for HybridBuffer using [`klauspost/compress`](https://github.com/klauspost/compress) for superior performance and more algorithms.

## Features

- **High Performance**: Uses `klauspost/compress` which is significantly faster than stdlib
- **Multiple Algorithms**: Gzip, Zstd, S2, Snappy, Zlib, Flate
- **Configurable Levels**: Fastest, Default, Better, Best
- **Streaming Support**: Efficient streaming compression/decompression
- **Drop-in Replacement**: Compatible with existing HybridBuffer middleware API

## Supported Algorithms

### Zstd (Recommended)
- **Best overall choice** - excellent compression ratio and speed
- Fast compression and decompression
- Good for general use cases

### S2 (Speed Focus)
- **Fastest compression/decompression**
- Moderate compression ratio
- Perfect for real-time applications

### Snappy (Google)
- Very fast compression/decompression
- Moderate compression ratio
- Good for high-throughput scenarios

### Gzip (Compatible)
- Standard gzip compression (faster than stdlib)
- Good compression ratio
- Wide compatibility

### Zlib (Deflate)
- Standard zlib compression
- Good compression ratio
- Wide compatibility

### Flate (Raw Deflate)
- Raw deflate compression
- Good compression ratio
- Lower overhead than gzip/zlib

## Usage

### Basic Usage

```go
import "schneider.vip/hybridbuffer/middleware/compression"

// Zstd compression (recommended)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.Zstd)),
)

// S2 compression (fastest)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(compression.S2)),
)
```

### With Compression Levels

```go
// Maximum compression
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(
        compression.Zstd,
        compression.WithLevel(compression.Best),
    )),
)

// Fastest compression
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(
        compression.S2,
        compression.WithLevel(compression.Fastest),
    )),
)
```

### Combined with Other Middleware

```go
import "schneider.vip/hybridbuffer/middleware/encryption"

// Combine compression and encryption
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(
        compression.New(compression.Zstd),
        encryption.New(),
    ),
)
```

## Performance Comparison

Based on typical text data:

| Algorithm | Compression Speed | Decompression Speed | Compression Ratio |
|-----------|------------------|-------------------|------------------|
| S2        | ⭐⭐⭐⭐⭐           | ⭐⭐⭐⭐⭐             | ⭐⭐⭐              |
| Snappy    | ⭐⭐⭐⭐⭐           | ⭐⭐⭐⭐⭐             | ⭐⭐⭐              |
| Zstd      | ⭐⭐⭐⭐            | ⭐⭐⭐⭐⭐             | ⭐⭐⭐⭐⭐            |
| Gzip      | ⭐⭐⭐             | ⭐⭐⭐⭐              | ⭐⭐⭐⭐             |
| Zlib      | ⭐⭐⭐             | ⭐⭐⭐⭐              | ⭐⭐⭐⭐             |
| Flate     | ⭐⭐⭐             | ⭐⭐⭐⭐              | ⭐⭐⭐⭐             |

## Algorithm Selection Guide

### Choose **Zstd** if:
- You want the best overall balance of speed and compression
- You're processing general data
- You want good compression ratios with reasonable speed

### Choose **S2** if:
- Speed is your primary concern
- You're processing real-time data streams
- You can accept moderate compression ratios

### Choose **Snappy** if:
- You need very fast compression/decompression
- You're working with high-throughput scenarios
- Moderate compression is sufficient

### Choose **Gzip** if:
- You need compatibility with standard gzip
- You're working with web-related data
- You want good compression with wide support

## Examples

### Real-time Data Processing
```go
// Use S2 for real-time performance
buf := hybridbuffer.New(
    hybridbuffer.WithThreshold(1024*1024), // 1MB
    hybridbuffer.WithMiddleware(compression.New(
        compression.S2,
        compression.WithLevel(compression.Fastest),
    )),
)
```

### Archival Storage
```go
// Use Zstd with best compression for storage
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(
        compression.Zstd,
        compression.WithLevel(compression.Best),
    )),
)
```

### Web API Response
```go
// Use Gzip for web compatibility
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(compression.New(
        compression.Gzip,
        compression.WithLevel(compression.Default),
    )),
)
```

## Performance Benefits

Compared to stdlib compression:

- **Gzip**: 2-3x faster compression, 1.5-2x faster decompression
- **Zlib**: Similar performance improvements
- **Additional algorithms**: Zstd, S2, Snappy provide much better performance characteristics

## Testing

```bash
# Run tests
go test -v

# Run benchmarks
go test -bench=. -benchmem
```

## Dependencies

- `github.com/klauspost/compress` - High-performance compression library
- `github.com/pkg/errors` - Error handling

## License

MIT License - same as parent project.