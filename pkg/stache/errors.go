package stache

import "errors"

var (
	// ErrNotFound is returned when a key is missing or expired.
	ErrNotFound = errors.New("cache: key not found")

	// ErrIncorrectType is returned when a value is requested with the wrong content type (e.g. GetString on a JSON entry).
	ErrIncorrectType = errors.New("cache: incorrect data type")
)
