package hybridbuffer

import (
	"schneider.vip/hybridbuffer/middleware"
	"schneider.vip/hybridbuffer/storage"
)

// Option defines functional options for buffer configuration
type Option func(*hybridBuffer)

// WithThreshold sets the memory threshold before switching to storage
// Default: 2MB
func WithThreshold(size int) Option {
	return func(b *hybridBuffer) {
		if size > 0 {
			b.threshold = size
		}
	}
}

// WithMiddleware adds one or more middlewares to the processing pipeline
// Middlewares are applied in the order they are added:
// - For writing: applied in reverse order (last middleware first)
// - For reading: applied in forward order (first middleware first)
//
// Example usage:
//
//	WithMiddleware(encryption.New(key))
//	WithMiddleware(compression.New(), encryption.New(key))
func WithMiddleware(middlewares ...middleware.Middleware) Option {
	return func(b *hybridBuffer) {
		b.middlewares = append(b.middlewares, middlewares...)
	}
}

// WithStorage sets the storage backend provider function
// If not specified, filesystem storage is used by default
//
// Example usage:
//
//	WithStorage(filesystem.New())
//	WithStorage(s3.New(client, bucket))
//	WithStorage(redis.New(client))
func WithStorage(provider func() storage.Backend) Option {
	return func(b *hybridBuffer) {
		b.storageProvider = provider
	}
}

// WithPreAlloc sets the pre-allocation size for the memory buffer
// This improves performance by avoiding multiple allocations during writes
// Default: threshold/2 (half of the memory threshold)
func WithPreAlloc(size int) Option {
	return func(b *hybridBuffer) {
		if size > 0 {
			b.preAllocSize = size
		}
	}
}
