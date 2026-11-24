package maputils

import "sync"

// SyncMap is a generic, typed wrapper around sync.Map: keys of type K,
// values of type V, no interface{} casts at the call sites.
//
// Prefer a plain map + sync.Mutex/RWMutex when you need iteration,
// snapshots, or complex invariants — sync.Map (and this wrapper) is tuned
// for append-only / disjoint-key concurrent workloads.
type SyncMap[K comparable, V any] struct {
	m *sync.Map
}

// NewSyncMap creates an empty typed concurrent map.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	s := new(SyncMap[K, V])
	s.m = new(sync.Map)
	return s
}

// Load returns the value stored under key and whether it was present.
// A missing key yields the zero value of V and false.
func (s *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := s.m.Load(key)
	if ok {
		return v.(V), ok
	}
	var empty V
	return empty, false
}

// LoadOrStore returns the existing value for the key if present;
// otherwise it stores and returns the given value. The loaded result is
// true if the value was loaded, false if stored.
func (s *SyncMap[K, V]) LoadOrStore(key K, val V) (V, bool) {
	actual, loaded := s.m.LoadOrStore(key, val)
	return actual.(V), loaded
}

// Store sets the value for a key.
func (s *SyncMap[K, V]) Store(key K, val V) {
	s.m.Store(key, val)
}

// Delete removes the value for a key.
func (s *SyncMap[K, V]) Delete(key K) {
	s.m.Delete(key)
}

// Range calls f sequentially for each key and value; f returning false
// stops the iteration.
func (s *SyncMap[K, V]) Range(f func(key K, val V) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

// Len counts the entries (O(n) — sync.Map does not track size).
func (s *SyncMap[K, V]) Len() int {
	n := 0
	s.m.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}
