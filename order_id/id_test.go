package order_id

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tx7do/go-utils/trans"
)

func TestGenerateOrderIdWithRandom(t *testing.T) {
	fmt.Println(GenerateOrderIdWithRandom("PT", "-", trans.Time(time.Now())))

	tm := time.Now()
	var ids map[string]bool
	ids = make(map[string]bool)
	count := 100
	for i := 0; i < count; i++ {
		ids[GenerateOrderIdWithRandom("PT", "", trans.Time(tm))] = true
	}
	assert.Equal(t, count, len(ids))
}

func TestGenerateOrderIdWithIndex(t *testing.T) {
	tm := time.Now()

	fmt.Println(GenerateOrderIdWithIncreaseIndex("PT", "", trans.Time(tm)))

	ids := make(map[string]bool)
	count := 100
	for i := 0; i < count; i++ {
		ids[GenerateOrderIdWithIncreaseIndex("PT", "", trans.Time(tm))] = true
	}
	assert.Equal(t, count, len(ids))
}

func TestGenerateOrderIdWithIndexThread(t *testing.T) {
	tm := time.Now()

	var wg sync.WaitGroup
	var ids sync.Map
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				id := GenerateOrderIdWithIncreaseIndex("PT", "", trans.Time(tm))
				ids.Store(id, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	aLen := 0
	ids.Range(func(k, v interface{}) bool {
		aLen++
		return true
	})
	assert.Equal(t, 1000, aLen)
}
