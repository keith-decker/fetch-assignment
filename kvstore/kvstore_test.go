package kvstore_test

import (
	"testing"

	"github.com/keith-decker/fetch-assignment/kvstore"
)

func TestKVStore(t *testing.T) {
	t.Run("Set and Get", func(t *testing.T) {
		store := kvstore.New()
		store.Set("key1", "value1")
		val, ok := store.Get("key1")
		if ok != nil || val != "value1" {
			t.Errorf("expected value1, got %v", val)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		store := kvstore.New()
		_, ok := store.Get("missing")
		if ok == nil {
			t.Error("did not return an error for missing key")
		}
	})

	t.Run("Delete key", func(t *testing.T) {
		store := kvstore.New()
		store.Set("key1", "value1")
		store.Delete("key1")
		_, ok := store.Get("key1")
		if ok == nil {
			t.Error("expected key1 to be deleted")
		}
	})
}
