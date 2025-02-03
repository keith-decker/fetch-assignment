package kvstore

import (
	"sync"
)

// KVStore is a simple key-value store.
type KVStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func New() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (kv *KVStore) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.store[key]
	return val, ok
}

func (kv *KVStore) Set(key, val string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store[key] = val
}

func (kv *KVStore) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.store, key)
}
