package log

import (
	"fmt"
	"sync"
)

// Record represents a single item in our log.
type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

// Store is an in-memory log store.
type Store struct {
	mu      sync.RWMutex
	records []Record
}

// NewStore creates a new empty store.
func NewStore() *Store {
	return &Store{
		records: make([]Record, 0),
	}
}

// Append adds a record to the log and returns its offset.
func (c *Store) Append(record Record) (uint64, error) {
	c.mu.Lock()         // Lock the store for writing
	defer c.mu.Unlock() // Unlock when the function finishes

	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)

	return record.Offset, nil
}

// Read retrieves a record at a specific offset.
func (c *Store) Read(offset uint64) (Record, error) {
	c.mu.RLock() // Read-Lock (allows multiple readers, but no writers)
	defer c.mu.RUnlock()

	if offset >= uint64(len(c.records)) {
		return Record{}, ErrOffsetNotFound
	}

	return c.records[offset], nil
}

// Define the error properly so other parts of the app can check for it
var ErrOffsetNotFound = fmt.Errorf("offset not found")
