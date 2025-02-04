package kvstore

import (
	"errors"
	"sync"
)

var (
	instance *KVStore
	once     sync.Once
)

// KVStore is a simple key-value store.
type KVStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func New() *KVStore {
	once.Do(func() {
		instance = &KVStore{
			store: make(map[string]string),
		}
	})
	return instance
}

func (kv *KVStore) Get(key string) (string, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.store[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
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
