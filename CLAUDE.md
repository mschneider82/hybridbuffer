# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Common Commands

### Testing
- `go test -v ./...` - Run all tests with verbose output
- `go test -race ./...` - Run tests with race detection
- `go test -bench=. -benchmem` - Run benchmarks with memory allocation stats

### Building and Development
- `go build` - Build the package
- `go mod tidy` - Clean up module dependencies
- `go fmt ./...` - Format code
- `go vet ./...` - Run static analysis

## Multi-Module Architecture

This is a modern Go hybrid memory/disk buffer library with a **multi-module architecture** using Go workspaces for clean separation of concerns and modular design.

### Repository Structure

This repository contains multiple Go modules in a single repository:

```
hybridbuffer/
├── go.mod                           # Main module: schneider.vip/hybridbuffer
├── go.work                          # Workspace configuration
├── storage/
│   ├── go.mod                       # Storage module: schneider.vip/hybridbuffer/storage
│   ├── filesystem/go.mod            # Filesystem storage: schneider.vip/hybridbuffer/storage/filesystem
│   ├── redis/go.mod                 # Redis storage: schneider.vip/hybridbuffer/storage/redis
│   └── s3/go.mod                    # S3 storage: schneider.vip/hybridbuffer/storage/s3
└── middleware/
    ├── go.mod                       # Middleware module: schneider.vip/hybridbuffer/middleware
    ├── compression/go.mod           # Compression middleware: schneider.vip/hybridbuffer/middleware/compression
    ├── compressionstdlib/go.mod     # Stdlib compression: schneider.vip/hybridbuffer/middleware/compressionstdlib
    └── encryption/go.mod            # Encryption middleware: schneider.vip/hybridbuffer/middleware/encryption
```

### Main Components

**HybridBuffer (main module)**
- Main API with full bytes.Buffer compatibility
- Optional thread-safety layer with RWMutex
- Delegates core functionality to CoreBuffer
- Provides convenience methods: Copy(), String(), Truncate(), Grow()

**Storage Modules**
- `storage/` - Base storage interfaces and implementations
- `storage/filesystem/` - File-based storage backend
- `storage/redis/` - Redis-based storage backend  
- `storage/s3/` - S3-compatible storage backend

**Middleware Modules**
- `middleware/` - Base middleware interfaces
- `middleware/compression/` - High-performance compression using klauspost/compress
- `middleware/compressionstdlib/` - Standard library compression
- `middleware/encryption/` - SIO-based encryption middleware

### Development Workflow

**Local Development with Workspaces**
This repository uses Go workspaces (`go.work`) for local development:

```bash
# Initialize workspace (already done)
go work init . ./storage ./middleware ./storage/filesystem # etc.

# Sync workspace
go work sync

# Build all modules
go build ./...

# Test all modules  
go test ./...

# Update dependencies in specific module
cd storage/redis
go get -u
```

**Replace Directives**
Each submodule uses replace directives to reference local dependencies:

```go
// In storage/filesystem/go.mod
replace schneider.vip/hybridbuffer/storage => ../

// In middleware/compression/go.mod  
replace schneider.vip/hybridbuffer/middleware => ../
```

**Module Dependencies**
- Main module depends on: storage, middleware, storage/filesystem
- Storage submodules depend on: storage (parent)
- Middleware submodules depend on: middleware (parent)
- Each module has its own specific external dependencies

### Versioning Strategy

**Git Tags for Submodules**
Each module requires prefixed tags:

```bash
# Main module
git tag v1.0.2

# Storage modules
git tag storage/v1.0.2
git tag storage/filesystem/v1.0.2
git tag storage/redis/v1.0.2
git tag storage/s3/v1.0.2

# Middleware modules  
git tag middleware/v1.0.2
git tag middleware/compression/v1.0.2
git tag middleware/compressionstdlib/v1.0.2
git tag middleware/encryption/v1.0.2
```

**Version Updates**
When updating versions:

1. Update all go.mod files with new version numbers
2. Commit changes
3. Create all prefixed tags
4. Push tags: `git push origin --tags`

### External Usage

**gopkg.in Integration**
This repository integrates with a custom gopkg Caddy plugin for vanity import paths:

```
schneider.vip/hybridbuffer                           → Main module
schneider.vip/hybridbuffer/storage                   → Storage interfaces
schneider.vip/hybridbuffer/storage/filesystem        → Filesystem backend
schneider.vip/hybridbuffer/middleware/compression    → Compression middleware
# etc.
```

**Consumer Usage**
External users can import specific modules:

```go
import (
    "schneider.vip/hybridbuffer"                              // Core functionality
    "schneider.vip/hybridbuffer/storage/filesystem"           // File storage
    "schneider.vip/hybridbuffer/middleware/compressionstdlib" // Compression
)
```

### Key Design Principles

**Clean Separation**
- Main module stays lightweight with core functionality only
- Storage backends are separate modules with their own dependencies
- Middleware components are modular and optional
- Each module can be versioned and developed independently

**Dependency Management**
- Storage modules only depend on storage interfaces
- Middleware modules only depend on middleware interfaces  
- External dependencies are isolated to specific modules
- No circular dependencies between modules

**Workspace Benefits**
- Local development works seamlessly across modules
- IDE support for cross-module references
- Unified testing and building across all modules
- Replace directives handle local dependencies automatically

### Important Notes

- **Always use go.work**: The workspace file is essential for local development
- **Replace directives**: Required for local module dependencies
- **Prefixed tags**: Essential for Go module discovery of submodules
- **Version consistency**: Keep all module versions synchronized
- **Clean dependencies**: Each module should only import what it needs

### Common Patterns

**Adding a New Storage Backend**
1. Create `storage/newbackend/` directory
2. Add `storage/newbackend/go.mod` with correct module path
3. Add replace directive: `replace schneider.vip/hybridbuffer/storage => ../`
4. Update workspace: `go work use ./storage/newbackend`
5. Implement storage.Backend interface
6. Add to main module dependencies if needed

**Adding a New Middleware**
1. Create `middleware/newmiddleware/` directory
2. Add `middleware/newmiddleware/go.mod` with correct module path
3. Add replace directive: `replace schneider.vip/hybridbuffer/middleware => ../`
4. Update workspace: `go work use ./middleware/newmiddleware`
5. Implement middleware.Middleware interface
6. Add to main module dependencies if needed

### Error Handling

**Common Issues**
- Module path conflicts: Ensure each go.mod has correct module declaration
- Missing replace directives: Local dependencies need replace statements
- Workspace sync: Run `go work sync` after adding modules
- Tag conflicts: Use prefixed tags for submodules (e.g., `storage/v1.0.2`)

**Debugging**
- Check workspace: `go work edit -print`
- Verify modules: `go list -m all` in each directory
- Test builds: `go build ./...` from root
- Check dependencies: `go mod graph` in each module