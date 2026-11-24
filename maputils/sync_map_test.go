package maputils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMap_Basic(t *testing.T) {
	s := NewSyncMap[string, int]()

	s.Store("a", 1)
	s.Store("b", 2)

	v, ok := s.Load("a")
	require.True(t, ok)
	assert.Equal(t, 1, v)

	v, ok = s.Load("missing")
	assert.False(t, ok)
	assert.Zero(t, v)

	s.Delete("a")
	_, ok = s.Load("a")
	assert.False(t, ok)
}

func TestSyncMap_LoadOrStore(t *testing.T) {
	s := NewSyncMap[string, string]()

	v, loaded := s.LoadOrStore("k", "first")
	assert.False(t, loaded)
	assert.Equal(t, "first", v)

	v, loaded = s.LoadOrStore("k", "second")
	assert.True(t, loaded, "已存在时返回既有值")
	assert.Equal(t, "first", v)

	v, _ = s.Load("k")
	assert.Equal(t, "first", v)
}

func TestSyncMap_RangeAndLen(t *testing.T) {
	s := NewSyncMap[int, string]()
	for i := 0; i < 10; i++ {
		s.Store(i, "v")
	}
	assert.Equal(t, 10, s.Len())

	seen := 0
	s.Range(func(key int, val string) bool {
		seen++
		return seen < 5 // 第 5 个停
	})
	assert.Equal(t, 5, seen)
}

func TestSyncMap_ConcurrentAccess(t *testing.T) {
	s := NewSyncMap[int, int]()
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				k := id*1000 + i
				s.Store(k, i)
				if _, ok := s.Load(k); !ok {
					t.Errorf("own key %d missing", k)
				}
			}
		}(w)
	}
	wg.Wait()
	assert.Equal(t, 8000, s.Len())
}
