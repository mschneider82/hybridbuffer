# Filesystem Storage Backend

This package provides a file system storage backend for HybridBuffer that stores data in temporary files.

## Features

- **Temporary file management** with automatic cleanup
- **Configurable temp directory** and file prefix
- **Zero external dependencies** (only standard library)
- **Cross-platform compatibility** (works on all Go-supported platforms)

## Usage

```go
import "schneider.vip/hybridbuffer/storage/filesystem"

// Basic usage with defaults
storage := filesystem.New()

// With custom options
storage := filesystem.New(
    filesystem.WithTempDir("/custom/temp"),
    filesystem.WithPrefix("myapp"),
)

// Use with HybridBuffer
buf := hybridbuffer.New(
    hybridbuffer.WithStorage(storage),
)
```

## Configuration Options

### WithTempDir(dir string)
Sets the directory where temporary files are created. If not specified, uses the system's default temporary directory (`os.TempDir()`).

```go
storage := filesystem.New(
    filesystem.WithTempDir("/var/tmp"),
)
```

### WithPrefix(prefix string)
Sets the prefix for temporary file names. Default is "hybridbuffer".

```go
storage := filesystem.New(
    filesystem.WithPrefix("myapp-buffer"),
)
```

This will create files like: `myapp-buffer-123456789.tmp`

## File Management

- **Temporary files** are created with a unique name using Go's `os.CreateTemp()`
- **File extension** is always `.tmp`
- **Automatic cleanup** occurs when `Remove()` is called
- **No automatic cleanup** on process exit - ensure proper `Close()` calls

## Error Handling

The filesystem backend returns standard Go errors for:
- **Permission denied** when unable to create/read files
- **Disk full** when unable to write data
- **File not found** when trying to read before writing
- **Directory not found** when temp directory doesn't exist

## Security Considerations

- **File permissions** follow system defaults (usually 0600 for temp files)
- **No encryption** at the storage level (use encryption middleware)
- **Temporary directory** should be on a secure filesystem
- **Cleanup responsibility** lies with the application

## Performance

- **Local filesystem speed** - performance depends on underlying storage
- **No network overhead** unlike cloud storage backends
- **Memory efficient** - streams data directly to/from disk
- **Platform optimized** using Go's standard library implementations