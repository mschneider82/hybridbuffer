// Package encryption provides SIO-based encryption middleware for HybridBuffer
package encryption

import (
	"crypto/rand"
	"io"

	"github.com/minio/sio"
	"schneider.vip/hybridbuffer/middleware"
)

// Cipher represents the encryption cipher algorithm
type Cipher int

const (
	// AES256GCM uses AES-256-GCM (default, hardware accelerated on most systems)
	AES256GCM Cipher = iota
	// ChaCha20Poly1305 uses ChaCha20-Poly1305 (better performance on systems without AES hardware)
	ChaCha20Poly1305
)

// Middleware implements encryption/decryption using MinIO's SIO library
type Middleware struct {
	key    [32]byte
	cipher Cipher
}

// Ensure Middleware implements middleware.Middleware interface
var _ middleware.Middleware = (*Middleware)(nil)

// Option configures encryption middleware
type Option func(*Middleware)

// WithKey sets a custom 32-byte encryption key
func WithKey(key []byte) Option {
	return func(m *Middleware) {
		if len(key) != 32 {
			panic("encryption key must be exactly 32 bytes")
		}
		copy(m.key[:], key)
	}
}

// WithCipher sets the encryption cipher algorithm
func WithCipher(cipher Cipher) Option {
	return func(m *Middleware) {
		m.cipher = cipher
	}
}

// New creates a new encryption middleware with optional key and cipher configuration
// If no key is provided via WithKey(), a random key is generated
// If no cipher is provided via WithCipher(), AES256GCM is used by default
func New(opts ...Option) *Middleware {
	m := &Middleware{
		cipher: AES256GCM, // Default cipher
	}

	// Generate random key by default
	if _, err := rand.Read(m.key[:]); err != nil {
		panic("failed to generate encryption key: " + err.Error())
	}

	// Apply options (may override the random key and cipher)
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Writer wraps an io.Writer with SIO encryption
func (m *Middleware) Writer(w io.Writer) io.Writer {
	config := m.getSIOConfig()
	encrypted, err := sio.EncryptWriter(w, config)
	if err != nil {
		panic("failed to create encryption writer: " + err.Error())
	}
	return encrypted
}

// Reader wraps an io.Reader with SIO decryption
func (m *Middleware) Reader(r io.Reader) io.Reader {
	config := m.getSIOConfig()
	decrypted, err := sio.DecryptReader(r, config)
	if err != nil {
		panic("failed to create decryption reader: " + err.Error())
	}
	return decrypted
}

// getSIOConfig returns the appropriate SIO configuration based on the cipher
func (m *Middleware) getSIOConfig() sio.Config {
	config := sio.Config{
		Key: m.key[:],
	}
	
	switch m.cipher {
	case AES256GCM:
		config.CipherSuites = []byte{sio.AES_256_GCM}
	case ChaCha20Poly1305:
		config.CipherSuites = []byte{sio.CHACHA20_POLY1305}
	default:
		// Default to AES256GCM if unknown cipher
		config.CipherSuites = []byte{sio.AES_256_GCM}
	}
	
	return config
}
