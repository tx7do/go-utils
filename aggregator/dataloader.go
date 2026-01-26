package aggregator

import (
	"reflect"
)

// TypedThunk 是一个返回特定类型 T 的函数
type TypedThunk[T any] func() (T, error)

// ThunkGetter 现在是完全强类型的
// R 是资源对象（如 OrgUnit），T 是关联对象（如 User）
type ThunkGetter[R any, T any] func(R) TypedThunk[T]

// PopulateWithLoader 专门为 DataLoader 设计的回填函数
func PopulateWithLoader[R any, T any](
	items []R,
	thunkGetter ThunkGetter[R, T],
	setter Setter[R, T],
) error {
	if len(items) == 0 {
		return nil
	}

	type pair struct {
		item  R
		thunk TypedThunk[T]
	}
	pairs := make([]pair, 0, len(items))

	for _, item := range items {
		if t := thunkGetter(item); t != nil {
			pairs = append(pairs, pair{item, t})
		}
	}

	for _, p := range pairs {
		val, err := p.thunk()
		if err != nil {
			continue
		}
		setter(p.item, val)
	}

	return nil
}

// PopulateTreeWithLoader 支持树状结构的 DataLoader 回填
func PopulateTreeWithLoader[R any, T any](
	items []R,
	thunkGetter ThunkGetter[R, T],
	setter Setter[R, T],
	children ChildrenFunc[R],
) error {
	if len(items) == 0 {
		return nil
	}

	type pair struct {
		item  R
		thunk TypedThunk[T]
	}
	var pairs []pair

	// 递归收集所有节点的 Thunk
	var collect func(list []R)
	collect = func(list []R) {
		for _, item := range list {
			if isNil(item) {
				continue
			}
			if t := thunkGetter(item); t != nil {
				pairs = append(pairs, pair{item: item, thunk: t})
			}
			if childList := children(item); len(childList) > 0 {
				collect(childList)
			}
		}
	}
	collect(items)

	// 执行所有 Thunk 并回填数据
	for _, p := range pairs {
		val, err := p.thunk()
		if err != nil {
			continue
		}
		setter(p.item, val)
	}

	return nil
}

// isNil 使用 reflect 判断任意值在运行时是否为 nil（仅对可为 nil 的类型返回 true）
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
