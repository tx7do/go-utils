package aggregator

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestNewLoaderFromFetch_SuccessAndMiss(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	var calledBatches [][]uint32

	fetch := func(ctx context.Context, keys []uint32) (map[uint32]string, error) {
		mu.Lock()
		kcopy := make([]uint32, len(keys))
		copy(kcopy, keys)
		calledBatches = append(calledBatches, kcopy)
		mu.Unlock()

		return map[uint32]string{
			1: "one",
			3: "three",
		}, nil
	}

	loader := NewLoaderFromFetch[uint32, string](fetch)

	// 在同一批次发起多个 Load 调用以触发 batch
	th1 := loader.Load(ctx, uint32(1))
	th2 := loader.Load(ctx, uint32(2))
	th3 := loader.Load(ctx, uint32(3))

	// 等待并断言结果（返回具体类型 string）
	r1, err1 := th1()
	if err1 != nil {
		t.Fatalf("unexpected error for key 1: %v", err1)
	}
	if r1 != "one" {
		t.Fatalf("expected key 1 -> \"one\", got %#v", r1)
	}

	r2, err2 := th2()
	if err2 != nil {
		t.Fatalf("unexpected error for key 2: %v", err2)
	}
	// 缺失项应返回零值（空字符串）
	if r2 != "" {
		t.Fatalf("expected key 2 to be missing (zero value), got %#v", r2)
	}

	r3, err3 := th3()
	if err3 != nil {
		t.Fatalf("unexpected error for key 3: %v", err3)
	}
	if r3 != "three" {
		t.Fatalf("expected key 3 -> \"three\", got %#v", r3)
	}

	// 验证 fetch 至少被调用一次且包含预期键
	mu.Lock()
	if len(calledBatches) == 0 {
		mu.Unlock()
		t.Fatalf("expected fetch to be called at least once")
	}
	found := map[uint32]bool{}
	for _, k := range calledBatches[0] {
		found[k] = true
	}
	mu.Unlock()
	if !found[uint32(1)] || !found[uint32(2)] || !found[uint32(3)] {
		t.Fatalf("expected batch to contain keys 1,2,3; got %#v", calledBatches[0])
	}
}

func TestNewLoaderFromFetch_ErrorPropagatedToAllResults(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("fetch-failed")

	fetch := func(ctx context.Context, keys []uint32) (map[uint32]string, error) {
		return nil, testErr
	}

	loader := NewLoaderFromFetch[uint32, string](fetch)

	th1 := loader.Load(ctx, uint32(1))
	th2 := loader.Load(ctx, uint32(2))

	_, e1 := th1()
	if e1 == nil {
		t.Fatalf("expected error for key 1, got nil")
	}
	if !errors.Is(e1, testErr) {
		t.Fatalf("expected error %v, got %v", testErr, e1)
	}

	_, e2 := th2()
	if e2 == nil {
		t.Fatalf("expected error for key 2, got nil")
	}
	if !errors.Is(e2, testErr) {
		t.Fatalf("expected error %v, got %v", testErr, e2)
	}
}

type Item struct {
	ID uint32
}

func TestThunkGetterFromLoader_SuccessAndZeroKey(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	var calledBatches [][]uint32

	fetch := func(ctx context.Context, keys []uint32) (map[uint32]string, error) {
		mu.Lock()
		kcopy := make([]uint32, len(keys))
		copy(kcopy, keys)
		calledBatches = append(calledBatches, kcopy)
		mu.Unlock()

		return map[uint32]string{
			1: "one",
			3: "three",
		}, nil
	}

	loader := NewLoaderFromFetch[uint32, string](fetch)
	getter := ThunkGetterFromLoader[Item, uint32, string](ctx, loader, func(it Item) uint32 { return it.ID })

	// 准备项：包含零值键 (ID = 0) 和有效键
	th1 := getter(Item{ID: 1})
	th0 := getter(Item{ID: 0}) // 零值，应该立即返回零值，不触发 loader
	th3 := getter(Item{ID: 3})

	// 调用并断言结果
	r1, e1 := th1()
	if e1 != nil {
		t.Fatalf("unexpected error for ID=1: %v", e1)
	}
	if r1 != "one" {
		t.Fatalf("expected \"one\" for ID=1, got %#v", r1)
	}

	r0, e0 := th0()
	if e0 != nil {
		t.Fatalf("unexpected error for ID=0: %v", e0)
	}
	if r0 != "" {
		t.Fatalf("expected zero value for ID=0, got %#v", r0)
	}

	r3, e3 := th3()
	if e3 != nil {
		t.Fatalf("unexpected error for ID=3: %v", e3)
	}
	if r3 != "three" {
		t.Fatalf("expected \"three\" for ID=3, got %#v", r3)
	}

	// 验证 fetch 被调用且不包含零值键
	mu.Lock()
	if len(calledBatches) == 0 {
		mu.Unlock()
		t.Fatalf("expected fetch to be called at least once")
	}
	found := map[uint32]bool{}
	for _, k := range calledBatches[0] {
		found[k] = true
	}
	mu.Unlock()

	if !found[1] || !found[3] {
		t.Fatalf("expected batch to contain keys 1 and 3; got %#v", calledBatches[0])
	}
	if found[0] {
		t.Fatalf("expected zero key 0 NOT to be passed to fetch, but it was present in %#v", calledBatches[0])
	}
}

func TestThunkGetterFromLoader_ErrorPropagated(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("fetch-failed")

	fetch := func(ctx context.Context, keys []uint32) (map[uint32]string, error) {
		return nil, testErr
	}

	loader := NewLoaderFromFetch[uint32, string](fetch)
	getter := ThunkGetterFromLoader[Item, uint32, string](ctx, loader, func(it Item) uint32 { return it.ID })

	th1 := getter(Item{ID: 1})
	th2 := getter(Item{ID: 2})

	_, e1 := th1()
	if e1 == nil {
		t.Fatalf("expected error for ID=1, got nil")
	}
	if !errors.Is(e1, testErr) {
		t.Fatalf("expected error %v for ID=1, got %v", testErr, e1)
	}

	_, e2 := th2()
	if e2 == nil {
		t.Fatalf("expected error for ID=2, got nil")
	}
	if !errors.Is(e2, testErr) {
		t.Fatalf("expected error %v for ID=2, got %v", testErr, e2)
	}
}
