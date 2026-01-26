package aggregator

// ChildrenFunc 定义如何获取子节点列表
type ChildrenFunc[R any] func(R) []R

// IDGetter 定义如何从对象中提取关联 ID
type IDGetter[R any, K comparable] func(R) K

// Setter 定义如何将获取到的关联实体填充回对象
type Setter[R any, T any] func(R, T)

// ResourceMap 通用的键值对容器
type ResourceMap[K comparable, T any] map[K]T

// Populate 处理扁平切片的数据回填
func Populate[K comparable, T any, R any](
	items []R,
	data ResourceMap[K, T],
	idGetter IDGetter[R, K],
	setter Setter[R, T],
) {
	if len(items) == 0 || len(data) == 0 {
		return
	}

	for _, item := range items {
		if val, ok := data[idGetter(item)]; ok {
			setter(item, val)
		}
	}
}

// PopulateTree 处理树状结构的数据回填
func PopulateTree[K comparable, T any, R any](
	items []R,
	data map[K]T,
	idGetter IDGetter[R, K],
	setter Setter[R, T],
	children ChildrenFunc[R],
) {
	if len(items) == 0 || len(data) == 0 {
		return
	}

	for _, item := range items {
		if val, ok := data[idGetter(item)]; ok {
			setter(item, val)
		}

		if childList := children(item); len(childList) > 0 {
			PopulateTree[K, T, R](childList, data, idGetter, setter, children)
		}
	}
}

// PopulateOne 处理单个对象的数据回填
func PopulateOne[K comparable, T any, R any](
	item R,
	data ResourceMap[K, T],
	idGetter IDGetter[R, K],
	setter Setter[R, T],
) {
	if val, ok := data[idGetter(item)]; ok {
		setter(item, val)
	}
}
