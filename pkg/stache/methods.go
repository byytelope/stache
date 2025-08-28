// Package stache provides an in-memory cache with TTL and
// JSON/string helpers. It is designed to be simple and safe
// for concurrent use.
package stache

import (
	"encoding/json"
	"fmt"
	"time"
)

// NewCache returns a pointer to an empty instance of Cache.
func NewCache() *Cache {
	return &Cache{index: map[string]cacheEntry{}}
}

// Set stores data in the cache under the given key, with the provided metadata.
// If TTL <= 0, the entry never expires.
func (c *Cache) Set(key string, data []byte, meta Meta) error {
	now := time.Now()
	meta.TTL = max(meta.TTL, 0)

	var expiresAt time.Time
	if meta.TTL > 0 {
		expiresAt = now.Add(meta.TTL)
	}

	buf := make([]byte, len(data))
	copy(buf, data)

	c.mutex.Lock()
	c.index[key] = cacheEntry{buf, meta.ContentType, expiresAt}
	c.mutex.Unlock()

	return nil
}

// SetJSON marshals the given value to JSON and stores it under the given key.
// The entry will expire after ttl, unless ttl <= 0 (no expiry).
func (c *Cache) SetJSON(key string, data any, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.Set(key, bytes, Meta{ttl, JSON})
}

// SetString stores a string value in the cache under the given key.
// The entry will expire after ttl, unless ttl <= 0 (no expiry).
func (c *Cache) SetString(key string, data string, ttl time.Duration) error {
	return c.Set(key, []byte(data), Meta{ttl, Text})
}

func (c *Cache) get(key string) (cacheEntry, error) {
	now := time.Now()

	c.mutex.RLock()
	entry, ok := c.index[key]
	c.mutex.RUnlock()

	if !ok {
		return cacheEntry{}, ErrNotFound
	}

	if entry.expiresAt.Before(now) && !entry.expiresAt.IsZero() {
		c.mutex.Lock()
		if entry2, ok2 := c.index[key]; ok2 && entry2.expiresAt.Equal(entry.expiresAt) {
			delete(c.index, key)
		}

		c.mutex.Unlock()

		return cacheEntry{}, ErrNotFound
	}

	return entry, nil
}

// GetBytes returns the raw byte slice for the given key.
// If the key does not exist or is expired, ErrNotFound is returned.
func (c *Cache) GetBytes(key string) ([]byte, error) {
	data, err := c.get(key)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, len(data.value))
	copy(bytes, data.value)

	return bytes, nil
}

// GetString returns the string value for the given key.
// If the entry is not of type Text, ErrIncorrectType is returned.
func (c *Cache) GetString(key string) (string, error) {
	data, err := c.get(key)
	if err != nil {
		return "", err
	}

	if data.contentType != Text {
		return "", ErrIncorrectType
	}

	return string(data.value), nil
}

// GetJSON unmarshals the JSON-encoded value into out, which must be a pointer.
// If the entry is not of type JSON, ErrIncorrectType is returned.
// If the key is missing or expired, ErrNotFound is returned.
func (c *Cache) GetJSON(key string, out any) error {
	item, err := c.get(key)
	if err != nil {
		return err
	}

	if item.contentType != JSON {
		return ErrIncorrectType
	}

	jsonErr := json.Unmarshal(item.value, out)
	if jsonErr != nil {
		return jsonErr
	}

	return nil
}

// Delete removes the entry for the given key, if present.
// It returns the removed entry and a boolean indicating whether it existed.
func (c *Cache) Delete(key string) (cacheEntry, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.index[key]
	delete(c.index, key)

	return item, exists
}

// Len returns the number of entries currently stored in the cache.
func (c *Cache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.index)
}

// Entries returns a snapshot of the current entries in the cache.
// Each entry is described by its key, size, content type, and expiry.
func (c *Cache) Entries() []EntryInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	info := []EntryInfo{}
	for k, v := range c.index {
		info = append(info, EntryInfo{
			Key:         k,
			Size:        len(v.value),
			ContentType: v.contentType,
			ExpiresAt:   v.expiresAt,
		})
	}

	return info
}

// String returns a summary string in the format `Cache(len={int})`.
// It implements the fmt.Stringer interface.
func (c *Cache) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return fmt.Sprintf("Cache(len=%d)", len(c.index))
}
