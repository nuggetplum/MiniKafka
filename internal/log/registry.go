package log

import (
	"os"
	"path/filepath"
	"sync"
)

// Registry manages multiple topics (multiple Stores)
type Registry struct {
	mu      sync.RWMutex
	stores  map[string]*Store
	baseDir string // Where to save all logs (e.g., "./my_data")
}

func NewRegistry(baseDir string) (*Registry, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &Registry{
		stores:  make(map[string]*Store),
		baseDir: baseDir,
	}, nil
}

// GetOrCreateStore finds an existing topic or creates a new one
func (r *Registry) GetOrCreateStore(topic string) (*Store, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Check if we already have it in memory
	if store, exists := r.stores[topic]; exists {
		return store, nil
	}

	// 2. If not, create the directory for this topic
	// Path: ./my_data/orders/store.bin
	topicDir := filepath.Join(r.baseDir, topic)
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		return nil, err
	}

	storeFile := filepath.Join(topicDir, "store.bin")

	// 3. Initialize the Store (reuses your code from Phase 3)
	store, err := NewStore(storeFile)
	if err != nil {
		return nil, err
	}

	// 4. Add to map
	r.stores[topic] = store
	return store, nil
}

// CloseAll closes all open file handles (Important for graceful shutdown)
func (r *Registry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, store := range r.stores {
		store.Close()
	}
	return nil
}
