package stache

import (
	"sync"
	"time"
)

// Cache is an in-memory key/value store with optional TTL expiration.
// It is safe for concurrent use by multiple goroutines.
type Cache struct {
	mutex sync.RWMutex
	index map[string]cacheEntry
}

type cacheEntry struct {
	value       []byte
	contentType ContentType
	expiresAt   time.Time
}

// ContentType indicates the encoding format of a cache entry value.
type ContentType string

const (
	JSON ContentType = "application/json"
	Text ContentType = "text/plain"
)

// Meta holds metadata for a cache entry, including its TTL and content type.
type Meta struct {
	// TTL specifies the time-to-live for a value.
	// If 0 or negative, the value never expires,
	TTL time.Duration

	// ContentType describes the MIME content type of the cached value.
	// Currently only supports `application/json` and `text/plain`
	ContentType ContentType
}

// EntryInfo describes a cached entry for introspection.
type EntryInfo struct {
	Key         string
	Size        int
	ContentType ContentType
	ExpiresAt   time.Time
}
