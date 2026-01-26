package aggregator

import "testing"

type testItem struct {
	ID   uint32
	Name string
}

func TestPopulate_SetsValuesForPointers(t *testing.T) {
	items := []*testItem{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	data := ResourceMap[uint32, string]{
		1: "alice",
		3: "carol",
	}

	Populate[uint32, string, *testItem](
		items,
		data,
		func(r *testItem) uint32 { return r.ID },
		func(r *testItem, v string) { r.Name = v },
	)

	if items[0].Name != "alice" {
		t.Fatalf("expected items[0].Name == %q, got %q", "alice", items[0].Name)
	}
	if items[1].Name != "" {
		t.Fatalf("expected items[1].Name == empty, got %q", items[1].Name)
	}
	if items[2].Name != "carol" {
		t.Fatalf("expected items[2].Name == %q, got %q", "carol", items[2].Name)
	}
}

func TestPopulate_SkipsWhenDataMissing(t *testing.T) {
	items := []*testItem{
		{ID: 10},
	}
	data := ResourceMap[uint32, string]{} // empty map

	Populate[uint32, string, *testItem](
		items,
		data,
		func(r *testItem) uint32 { return r.ID },
		func(r *testItem, v string) { r.Name = v },
	)

	if items[0].Name != "" {
		t.Fatalf("expected items[0].Name to remain empty when data missing, got %q", items[0].Name)
	}
}

type treeItem struct {
	ID       uint32
	Name     string
	Children []*treeItem
}

func TestPopulateTree_SetsValuesRecursively(t *testing.T) {
	// 构建树：root(1) -> child(2) -> grand(3)
	grand := &treeItem{ID: 3}
	child := &treeItem{ID: 2, Children: []*treeItem{grand}}
	root := &treeItem{ID: 1, Children: []*treeItem{child}}
	items := []*treeItem{root}

	data := map[uint32]string{
		1: "alice",
		3: "carol",
	}

	PopulateTree[uint32, string, *treeItem](
		items,
		data,
		func(r *treeItem) uint32 { return r.ID },
		func(r *treeItem, v string) { r.Name = v },
		func(r *treeItem) []*treeItem { return r.Children },
	)

	if root.Name != "alice" {
		t.Fatalf("expected root.Name == %q, got %q", "alice", root.Name)
	}
	if child.Name != "" {
		t.Fatalf("expected child.Name to remain empty, got %q", child.Name)
	}
	if grand.Name != "carol" {
		t.Fatalf("expected grand.Name == %q, got %q", "carol", grand.Name)
	}
}

func TestPopulateTree_SkipsWhenDataMissing(t *testing.T) {
	item := &treeItem{ID: 10}
	items := []*treeItem{item}
	data := map[uint32]string{} // empty map

	PopulateTree[uint32, string, *treeItem](
		items,
		data,
		func(r *treeItem) uint32 { return r.ID },
		func(r *treeItem, v string) { r.Name = v },
		func(r *treeItem) []*treeItem { return r.Children },
	)

	if item.Name != "" {
		t.Fatalf("expected item.Name to remain empty when data missing, got %q", item.Name)
	}
}

func TestPopulateOne_SetsValueForPointer(t *testing.T) {
	item := &testItem{ID: 1}

	data := ResourceMap[uint32, string]{
		1: "alice",
	}

	PopulateOne[uint32, string, *testItem](
		item,
		data,
		func(r *testItem) uint32 { return r.ID },
		func(r *testItem, v string) { r.Name = v },
	)

	if item.Name != "alice" {
		t.Fatalf("expected item.Name == %q, got %q", "alice", item.Name)
	}
}

func TestPopulateOne_SkipsWhenDataMissing(t *testing.T) {
	item := &testItem{ID: 10}

	data := ResourceMap[uint32, string]{} // empty map

	PopulateOne[uint32, string, *testItem](
		item,
		data,
		func(r *testItem) uint32 { return r.ID },
		func(r *testItem, v string) { r.Name = v },
	)

	if item.Name != "" {
		t.Fatalf("expected item.Name to remain empty when data missing, got %q", item.Name)
	}
}
