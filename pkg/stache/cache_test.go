package stache

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSetGetBytes(t *testing.T) {
	c := NewCache()
	val := []byte("dababy")

	if err := c.Set("k1", val, Meta{time.Second, Text}); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	got, err := c.GetBytes("k1")
	if err != nil {
		t.Fatalf("GetBytes error: %v", err)
	}

	// Expect exact contents
	if !reflect.DeepEqual(got, val) {
		t.Fatalf("GetBytes mismatch: got=%q want=%q", string(got), string(val))
	}
}

func TestSetStringGetString(t *testing.T) {
	c := NewCache()

	if err := c.SetString("k", "yo", 500*time.Millisecond); err != nil {
		t.Fatalf("SetString error: %v", err)
	}

	s, err := c.GetString("k")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}

	if s != "yo" {
		t.Fatalf("GetString mismatch: got=%q want=%q", s, "yo")
	}
}

func TestSetJSONGetJSON(t *testing.T) {
	c := NewCache()

	type user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	in := user{ID: 69, Name: "DaBaby"}

	if err := c.SetJSON("u:69", in, time.Second); err != nil {
		t.Fatalf("SetJSON error: %v", err)
	}

	var out user
	if err := c.GetJSON("u:69", &out); err != nil {
		t.Fatalf("GetJSON error: %v", err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("GetJSON mismatch: got=%v want=%v", out, in)
	}
}

func TestGetStringWrongType(t *testing.T) {
	c := NewCache()

	// Store JSON, then try GetString → ErrIncorrectType
	if err := c.SetJSON("j", map[string]any{"a": 1}, time.Second); err != nil {
		t.Fatalf("SetJSON error: %v", err)
	}

	_, err := c.GetString("j")
	if !errors.Is(err, ErrIncorrectType) {
		t.Fatalf("expected ErrIncorrectType, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	c := NewCache()

	_, err := c.GetBytes("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTTLExpiryOnAccess(t *testing.T) {
	c := NewCache()

	if err := c.SetString("e", "deez", 30*time.Millisecond); err != nil {
		t.Fatalf("SetString error: %v", err)
	}
	time.Sleep(60 * time.Millisecond)

	_, err := c.GetString("e")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after expiry, got %v", err)
	}

	// Access after expiry should also delete; Len should drop to 0
	if n := c.Len(); n != 0 {
		t.Fatalf("expected Len()=0 after expiry cleanup, got %d", n)
	}
}

func TestDelete(t *testing.T) {
	c := NewCache()

	if err := c.SetString("d", "x", time.Second); err != nil {
		t.Fatalf("SetString error: %v", err)
	}

	_, ok := c.Delete("d")
	if !ok {
		t.Fatalf("expected delete ok=true")
	}

	_, err := c.GetString("d")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	// Deleting again should be ok=false
	_, ok = c.Delete("d")
	if ok {
		t.Fatalf("expected ok=false for second delete")
	}
}

func TestEntriesAndLen(t *testing.T) {
	c := NewCache()
	_ = c.SetString("a", "A", time.Second)
	_ = c.SetString("b", "BB", time.Second)

	if n := c.Len(); n != 2 {
		t.Fatalf("Len() mismatch: got=%d want=2", n)
	}

	ents := c.Entries()
	if len(ents) != 2 {
		t.Fatalf("Entries length: got=%d want=2", len(ents))
	}

	// Verify sizes/content types are sensible
	m := map[string]EntryInfo{}
	for _, e := range ents {
		m[e.Key] = e
	}
	if m["a"].Size != 1 || m["a"].ContentType != Text {
		t.Fatalf("entry a info unexpected: %+v", m["a"])
	}
	if m["b"].Size != 2 || m["b"].ContentType != Text {
		t.Fatalf("entry b info unexpected: %+v", m["b"])
	}
}

func TestConcurrencyBasic(t *testing.T) {
	c := NewCache()
	const N = 100

	var wg sync.WaitGroup
	wg.Add(2)

	// writer
	go func() {
		defer wg.Done()
		for range N {
			_ = c.SetString("k", "v", time.Second)
		}
	}()

	// reader
	go func() {
		defer wg.Done()
		for range N {
			_, _ = c.GetString("k")
		}
	}()

	wg.Wait()
}
