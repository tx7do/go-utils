package aggregator

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"
)

// NewLoaderFromFetch 将一个按 K 批量抓取函数封装为 *dataloader.Loader[K,V]。
// fetch 接受 ctx 和 []K，返回 map[K]V 和可选错误；若 fetch 返回 error，则该 batch 的所有结果都会带上该 error。
func NewLoaderFromFetch[K comparable, V any](fetch func(context.Context, []K) (map[K]V, error)) *dataloader.Loader[K, V] {
	batch := func(ctx context.Context, keys []K) []*dataloader.Result[V] {
		resMap, err := fetch(ctx, keys)

		results := make([]*dataloader.Result[V], len(keys))
		var zero V
		for i, k := range keys {
			if err != nil {
				results[i] = &dataloader.Result[V]{Error: err}
				continue
			}
			if v, ok := resMap[k]; ok {
				results[i] = &dataloader.Result[V]{Data: v}
			} else {
				results[i] = &dataloader.Result[V]{Data: zero}
			}
		}
		return results
	}

	return dataloader.NewBatchedLoader[K, V](batch)
}

// ThunkGetterFromLoader 返回一个 ThunkGetter，可直接传入 PopulateWithLoader。
// keyOf 用于从项 R 提取键 K；传入的 ctx 会被用于 loader.Load。
func ThunkGetterFromLoader[R any, K comparable, V any](
	ctx context.Context,
	loader *dataloader.Loader[K, V],
	keyOf func(R) K,
) ThunkGetter[R, V] {
	return func(r R) TypedThunk[V] {
		k := keyOf(r)
		var zeroK K
		var zeroV V

		// 零值键立即返回零值，不触发 loader
		if k == zeroK {
			return func() (V, error) { return zeroV, nil }
		}

		// 立即调用 Load 以将键加入当前批次，返回的 thunk 等待结果
		th := loader.Load(ctx, k)

		return func() (V, error) {
			val, err := th()
			if err != nil {
				return zeroV, err
			}
			return val, nil
		}
	}
}
