package aggregator

import (
	"errors"
	"testing"
)

func TestPopulateWithLoader_Success(t *testing.T) {
	items := []*testItem{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	// 为每个 ID 提供对应的 TypedThunk
	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "alice", nil },
		2: func() (string, error) { return "bob", nil },
		3: func() (string, error) { return "carol", nil },
	}

	thunkGetter := func(r *testItem) TypedThunk[string] {
		return thunks[r.ID]
	}

	setter := func(r *testItem, v string) {
		r.Name = v
	}

	if err := PopulateWithLoader(items, thunkGetter, setter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[uint32]string{1: "alice", 2: "bob", 3: "carol"}
	for _, it := range items {
		if got := it.Name; got != expected[it.ID] {
			t.Fatalf("id=%d: expected name=%q, got=%q", it.ID, expected[it.ID], got)
		}
	}
}

func TestPopulateWithLoader_ThunkErrorIsSkipped(t *testing.T) {
	items := []*testItem{
		{ID: 1},
		{ID: 2},
	}

	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "alice", nil },
		2: func() (string, error) { return "", errors.New("load-failed") },
	}

	thunkGetter := func(r *testItem) TypedThunk[string] {
		return thunks[r.ID]
	}

	setCalls := 0
	setter := func(r *testItem, v string) {
		setCalls++
		r.Name = v
	}

	if err := PopulateWithLoader(items, thunkGetter, setter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if items[0].Name != "alice" {
		t.Fatalf("expected first item populated, got %q", items[0].Name)
	}
	// 第二个 thunk 返回错误，应该被跳过，Name 保持空
	if items[1].Name != "" {
		t.Fatalf("expected second item unchanged, got %q", items[1].Name)
	}
	if setCalls != 1 {
		t.Fatalf("expected setter called once, got %d", setCalls)
	}
}

func TestPopulateWithLoader_NilThunkIsSkipped(t *testing.T) {
	items := []*testItem{
		{ID: 1},
		{ID: 42}, // 没有 thunk，thunkGetter 返回 nil
	}

	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "alice", nil },
	}

	thunkGetter := func(r *testItem) TypedThunk[string] {
		return thunks[r.ID] // 未命中的会返回 nil
	}

	setter := func(r *testItem, v string) {
		r.Name = v
	}

	if err := PopulateWithLoader(items, thunkGetter, setter); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if items[0].Name != "alice" {
		t.Fatalf("expected first item populated, got %q", items[0].Name)
	}
	if items[1].Name != "" {
		t.Fatalf("expected second item unchanged (no thunk), got %q", items[1].Name)
	}
}

type testNode struct {
	ID       uint32
	Name     string
	Children []*testNode
}

func TestPopulateTreeWithLoader_Success(t *testing.T) {
	// 构建树： root(1) -> childA(2)->grand(4), childB(3)
	root := &testNode{ID: 1}
	childA := &testNode{ID: 2}
	childB := &testNode{ID: 3}
	grand := &testNode{ID: 4}
	childA.Children = []*testNode{grand}
	root.Children = []*testNode{childA, childB}
	items := []*testNode{root}

	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "root-name", nil },
		2: func() (string, error) { return "childA-name", nil },
		3: func() (string, error) { return "childB-name", nil },
		4: func() (string, error) { return "grand-name", nil },
	}

	thunkGetter := func(n *testNode) TypedThunk[string] { return thunks[n.ID] }
	setter := func(n *testNode, v string) { n.Name = v }
	children := func(n *testNode) []*testNode { return n.Children }

	if err := PopulateTreeWithLoader(items, thunkGetter, setter, children); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expect := map[uint32]string{1: "root-name", 2: "childA-name", 3: "childB-name", 4: "grand-name"}

	if root.Name != expect[root.ID] {
		t.Fatalf("id=%d: expected %q, got %q", root.ID, expect[root.ID], root.Name)
	}
	if childA.Name != expect[childA.ID] {
		t.Fatalf("id=%d: expected %q, got %q", childA.ID, expect[childA.ID], childA.Name)
	}
	if childB.Name != expect[childB.ID] {
		t.Fatalf("id=%d: expected %q, got %q", childB.ID, expect[childB.ID], childB.Name)
	}
	if grand.Name != expect[grand.ID] {
		t.Fatalf("id=%d: expected %q, got %q", grand.ID, expect[grand.ID], grand.Name)
	}
}

func TestPopulateTreeWithLoader_ThunkErrorIsSkipped(t *testing.T) {
	// root(1) -> childA(2), childB(3)
	root := &testNode{ID: 1}
	childA := &testNode{ID: 2}
	childB := &testNode{ID: 3}
	root.Children = []*testNode{childA, childB}
	items := []*testNode{root}

	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "root-name", nil },
		2: func() (string, error) { return "", errors.New("load-failed") }, // error for node 2
		3: func() (string, error) { return "childB-name", nil },
	}

	thunkGetter := func(n *testNode) TypedThunk[string] { return thunks[n.ID] }
	setterCalls := 0
	setter := func(n *testNode, v string) {
		setterCalls++
		n.Name = v
	}
	children := func(n *testNode) []*testNode { return n.Children }

	if err := PopulateTreeWithLoader(items, thunkGetter, setter, children); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// node 2 should remain empty, others set
	if root.Name != "root-name" {
		t.Fatalf("expected root populated, got %q", root.Name)
	}
	if childA.Name != "" {
		t.Fatalf("expected childA unchanged due to thunk error, got %q", childA.Name)
	}
	if childB.Name != "childB-name" {
		t.Fatalf("expected childB populated, got %q", childB.Name)
	}
	if setterCalls != 2 {
		t.Fatalf("expected setter called twice (root and childB), got %d", setterCalls)
	}
}

func TestPopulateTreeWithLoader_NilThunkIsSkippedAndNilNodeIgnored(t *testing.T) {
	// 包含一个 nil 子节点以验证不会 panic
	root := &testNode{ID: 1}
	childA := &testNode{ID: 2}
	var nilChild *testNode = nil
	childB := &testNode{ID: 3}
	root.Children = []*testNode{childA, nilChild, childB}
	items := []*testNode{root}

	thunks := map[uint32]TypedThunk[string]{
		1: func() (string, error) { return "root-name", nil },
		2: func() (string, error) { return "childA-name", nil },
		// 3 没有 thunk，thunkGetter 返回 nil，应该被跳过
	}

	thunkGetter := func(n *testNode) TypedThunk[string] {
		return thunks[n.ID]
	}
	setter := func(n *testNode, v string) { n.Name = v }
	children := func(n *testNode) []*testNode { return n.Children }

	if err := PopulateTreeWithLoader(items, thunkGetter, setter, children); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Name != "root-name" {
		t.Fatalf("expected root populated, got %q", root.Name)
	}
	if childA.Name != "childA-name" {
		t.Fatalf("expected childA populated, got %q", childA.Name)
	}
	if childB.Name != "" {
		t.Fatalf("expected childB unchanged (no thunk), got %q", childB.Name)
	}
}
