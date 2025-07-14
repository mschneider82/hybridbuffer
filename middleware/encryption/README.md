# Encryption Middleware

Authenticated encryption middleware for HybridBuffer using MinIO's SIO library with configurable cipher algorithms.

## Features

- **Authenticated Encryption**: Uses MinIO SIO for tamper-proof encryption
- **Multiple Cipher Support**: AES-256-GCM and ChaCha20-Poly1305
- **Hardware Acceleration**: AES-256-GCM benefits from hardware acceleration on most systems
- **Cross-Platform Performance**: ChaCha20-Poly1305 provides better performance on systems without AES hardware
- **Automatic Key Generation**: Secure random key generation by default
- **Custom Key Support**: Bring your own 32-byte encryption key

## Supported Ciphers

### AES-256-GCM (Default)
- **Best choice for most systems** with AES hardware acceleration
- Widely supported and battle-tested
- Hardware accelerated on Intel/AMD CPUs with AES-NI
- Standard in many security frameworks

### ChaCha20-Poly1305
- **Best choice for systems without AES hardware** (ARM, older CPUs)
- Designed by Daniel J. Bernstein
- Constant-time implementation resistant to timing attacks
- Often faster than AES on mobile devices and embedded systems

## Usage

### Basic Usage

```go
import "schneider.vip/hybridbuffer/middleware/encryption"

// Auto-generated key with AES-256-GCM (default)
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New()),
)
```

### With Custom Cipher

```go
// Use ChaCha20-Poly1305 for better performance on ARM/mobile
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithCipher(encryption.ChaCha20Poly1305),
    )),
)
```

### With Custom Key

```go
// Use your own 32-byte key
key := make([]byte, 32)
// ... fill key with secure random data or derive from password
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithKey(key),
    )),
)
```

### Combined Options

```go
// Custom key with ChaCha20-Poly1305
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithKey(key),
        encryption.WithCipher(encryption.ChaCha20Poly1305),
    )),
)
```

### Combined with Other Middleware

```go
import "schneider.vip/hybridbuffer/middleware/compression"

// Combine compression and encryption
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(
        compression.New(compression.Zstd),
        encryption.New(encryption.WithKey(key)),
    ),
)
```

## Cipher Selection Guide

### Choose **AES-256-GCM** if:
- You're running on modern Intel/AMD CPUs with AES-NI
- You need maximum compatibility and standards compliance
- You're working in enterprise environments
- Hardware acceleration is available

### Choose **ChaCha20-Poly1305** if:
- You're running on ARM processors (mobile, embedded)
- You're on older CPUs without AES hardware acceleration
- You need consistent performance across different architectures
- You're working with real-time applications where timing consistency matters

## Performance Characteristics

Based on typical hardware:

| Cipher | Intel/AMD with AES-NI | ARM/Mobile | Older CPUs |
|--------|----------------------|------------|------------|
| AES-256-GCM | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| ChaCha20-Poly1305 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

## Security

Both ciphers provide:
- **Authenticated Encryption**: Prevents tampering and detects modifications
- **Semantic Security**: Same plaintext produces different ciphertext
- **Forward Security**: Compromised keys don't affect past communications
- **Nonce Handling**: Automatic nonce generation and management

## API Reference

### Types

```go
type Cipher int

const (
    AES256GCM Cipher = iota           // AES-256-GCM
    ChaCha20Poly1305                  // ChaCha20-Poly1305
)
```

### Functions

```go
// Create new encryption middleware
func New(opts ...Option) *Middleware

// Options
func WithKey(key []byte) Option              // 32-byte key required
func WithCipher(cipher Cipher) Option        // Choose cipher algorithm
```

## Examples

### Web Application (Intel Server)
```go
// Use AES-256-GCM for web servers with AES-NI
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithCipher(encryption.AES256GCM),
    )),
)
```

### Mobile App (ARM)
```go
// Use ChaCha20-Poly1305 for mobile devices
buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithCipher(encryption.ChaCha20Poly1305),
    )),
)
```

### Key Derivation from Password
```go
import "golang.org/x/crypto/pbkdf2"
import "crypto/sha256"

// Derive key from password (example - use proper key derivation in production)
password := "user-password"
salt := []byte("application-specific-salt")
key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

buf := hybridbuffer.New(
    hybridbuffer.WithMiddleware(encryption.New(
        encryption.WithKey(key),
    )),
)
```

## Important Notes

### Key Management
- Keys must be exactly 32 bytes (256 bits)
- Use cryptographically secure random key generation
- Store keys securely - consider using key management services
- Rotate keys periodically in production environments

### Compatibility
- Different ciphers produce different encrypted formats
- Use the same cipher for encryption and decryption
- Encrypted data is not compatible between cipher types

### Performance
- Test both ciphers on your target hardware
- Consider your threat model and performance requirements
- Use benchmarks to validate performance assumptions

## Testing

```bash
# Run tests
go test -v

# Run benchmarks
go test -bench=. -benchmem
```

## Dependencies

- `github.com/minio/sio` - Authenticated encryption library
- Go standard library `crypto/rand` for secure random number generation

## License

MIT License - same as parent project.