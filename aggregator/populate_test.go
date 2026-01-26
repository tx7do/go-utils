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

type multiItem struct {
	IDs   []uint32
	Names []string
}

func TestPopulateMulti_SetsValuesForPointers(t *testing.T) {
	items := []*multiItem{
		{IDs: []uint32{1, 2}},
		{IDs: []uint32{}},
		{IDs: []uint32{3, 4}},
	}

	data := ResourceMap[uint32, string]{
		1: "alice",
		3: "carol",
		4: "dave",
	}

	PopulateMulti[uint32, string, *multiItem](
		items,
		data,
		func(r *multiItem) []uint32 { return r.IDs },
		func(r *multiItem, vals []string) { r.Names = vals },
	)

	// items[0] 应只包含 id 1 的匹配结果 "alice"
	if len(items[0].Names) != 1 || items[0].Names[0] != "alice" {
		t.Fatalf("expected items[0].Names == [\"alice\"], got %#v", items[0].Names)
	}

	// items[1] IDs 为空，Names 应保持 nil 或空
	if len(items[1].Names) != 0 {
		t.Fatalf("expected items[1].Names to remain empty, got %#v", items[1].Names)
	}

	// items[2] 应包含 id 3 和 4 的匹配结果，且顺序保持 [\"carol\",\"dave\"]
	if len(items[2].Names) != 2 || items[2].Names[0] != "carol" || items[2].Names[1] != "dave" {
		t.Fatalf("expected items[2].Names == [\"carol\",\"dave\"], got %#v", items[2].Names)
	}
}

func TestPopulateMulti_SkipsWhenDataMissing(t *testing.T) {
	items := []*multiItem{
		{IDs: []uint32{10}},
	}
	data := ResourceMap[uint32, string]{} // empty map

	PopulateMulti[uint32, string, *multiItem](
		items,
		data,
		func(r *multiItem) []uint32 { return r.IDs },
		func(r *multiItem, vals []string) { r.Names = vals },
	)

	// 数据缺失，不应填充 Names
	if len(items[0].Names) != 0 {
		t.Fatalf("expected items[0].Names to remain empty when data missing, got %#v", items[0].Names)
	}
}

type treeMultiItem struct {
	IDs      []uint32
	Names    []string
	Children []*treeMultiItem
}

func TestPopulateTreeMulti_SetsValuesRecursively(t *testing.T) {
	// 构建树：root(IDs: [1]) -> child(IDs: [2]) -> grand(IDs: [3,4])
	grand := &treeMultiItem{IDs: []uint32{3, 4}}
	child := &treeMultiItem{IDs: []uint32{2}, Children: []*treeMultiItem{grand}}
	root := &treeMultiItem{IDs: []uint32{1}, Children: []*treeMultiItem{child}}
	items := []*treeMultiItem{root}

	data := map[uint32]string{
		1: "alice",
		3: "carol",
		4: "dave",
	}

	PopulateTreeMulti[uint32, string, *treeMultiItem](
		items,
		data,
		func(r *treeMultiItem) []uint32 { return r.IDs },
		func(r *treeMultiItem, vals []string) { r.Names = vals },
		func(r *treeMultiItem) []*treeMultiItem { return r.Children },
	)

	// root 应包含 id 1 的匹配结果
	if len(root.Names) != 1 || root.Names[0] != "alice" {
		t.Fatalf("expected root.Names == [\"alice\"], got %#v", root.Names)
	}

	// child IDs 为 [2] 且 data 中无 2，Names 应保持空
	if len(child.Names) != 0 {
		t.Fatalf("expected child.Names to remain empty, got %#v", child.Names)
	}

	// grand IDs 为 [3,4]，应按顺序填充 ["carol","dave"]
	if len(grand.Names) != 2 || grand.Names[0] != "carol" || grand.Names[1] != "dave" {
		t.Fatalf("expected grand.Names == [\"carol\",\"dave\"], got %#v", grand.Names)
	}
}

func TestPopulateTreeMulti_SkipsWhenDataMissing(t *testing.T) {
	item := &treeMultiItem{IDs: []uint32{10}}
	items := []*treeMultiItem{item}
	data := map[uint32]string{} // empty map

	PopulateTreeMulti[uint32, string, *treeMultiItem](
		items,
		data,
		func(r *treeMultiItem) []uint32 { return r.IDs },
		func(r *treeMultiItem, vals []string) { r.Names = vals },
		func(r *treeMultiItem) []*treeMultiItem { return r.Children },
	)

	if len(item.Names) != 0 {
		t.Fatalf("expected item.Names to remain empty when data missing, got %#v", item.Names)
	}
}
