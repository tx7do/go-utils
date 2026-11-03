# aggregator

English | [简体中文](README.zh-CN.md)

`aggregator` is a high-performance toolkit for data aggregation and relation back-filling, built to tame the **N+1 query problem** and the complexity of **DTO data populating** in microservice architectures.

## Features

- **Parallel fetching**: `ExecuteParallel` runs tasks concurrently on top of `errgroup`, with concurrency limits, overall timeout, failure retries (exponential backoff with jitter) and panic recovery;
- **Strongly typed generic population**: declarative data back-filling via Go generics, covering single objects, flat slices and recursive trees, with both single-ID and multi-ID relation shapes;
- **DataLoader integration**: `NewLoaderFromFetch` / `ThunkGetterFromLoader` wrap a batch-fetch function into a [`graph-gophers/dataloader`](https://github.com/graph-gophers/dataloader) Loader, and `PopulateWithLoader` / `PopulateTreeWithLoader` use it to coalesce loads within a request;
- **Type safety**: eliminates `any` type assertions, catching type errors at compile time.

## Installation

```bash
go get github.com/tx7do/go-utils/aggregator
```

```go
import "github.com/tx7do/go-utils/aggregator"
```

Dependencies: the Go standard library, `golang.org/x/sync`, and `github.com/graph-gophers/dataloader/v7` (used only by the DataLoader-related APIs).

## Core Concepts

1. **Collect**: extract the unique IDs that need to be fetched from the raw data;
2. **Fetch**: concurrently call external services (RPC/SQL) to fetch data in batch;
3. **Populate**: inject the fetched related data back into the original DTO objects.

## Quick Start

### 1. Fetch data in parallel

Use `ExecuteParallel` to run multiple fetch tasks concurrently without managing locks and waiting by hand. Tasks are passed as a **slice**:

```go
package main

import (
	"context"
	"time"

	"github.com/tx7do/go-utils/aggregator"
)

func fetchDetails(ctx context.Context) error {
	var (
		userMap = make(aggregator.ResourceMap[uint32, *User])
		deptMap = make(aggregator.ResourceMap[uint32, *Dept])
	)

	return aggregator.ExecuteParallel(ctx,
		[]aggregator.ParallelFetcher{
			func(ctx context.Context) error {
				// Fetch users and write into userMap
				return nil
			},
			func(ctx context.Context) error {
				// Fetch departments and write into deptMap
				return nil
			},
		},
		aggregator.WithLimit(10),                // max concurrency (default 20)
		aggregator.WithTimeout(5*time.Second),   // overall timeout (default unlimited)
		aggregator.WithRetry(2),                 // retries per task (default 0)
	)
}
```

When any task returns an error, the context derived by `errgroup` is cancelled so the remaining tasks exit promptly, and `ExecuteParallel` returns that error. An empty task slice returns `nil` immediately; `nil` tasks are skipped.

### 2. Declarative population (Populate)

Declare the "get ID" and "assign" logic as closures. Note that the data source parameter of `PopulateTree` is a plain `map[K]T`:

```go
// Back-fill leader info across an org tree
aggregator.PopulateTree(
	orgUnits,                                            // original object list
	userMap,                                             // data source map[uint32]*User
	func(o *OrgUnit) uint32 { return o.LeaderID },       // extract the related ID
	func(o *OrgUnit, u *User) { o.LeaderName = u.Name }, // assign the value
	func(o *OrgUnit) []*OrgUnit { return o.Children },   // path to child nodes
)
```

### 3. DataLoader-backed population

Ideal when a single request resolves many "fetch related object by ID" lookups; multiple loads are automatically coalesced into one batch request:

```go
// 1) Wrap a batch-fetch function into a Loader
loader := aggregator.NewLoaderFromFetch[uint32, *User](
	func(ctx context.Context, ids []uint32) (map[uint32]*User, error) {
		// One RPC/SQL call fetching all ids at once, returned as a map
		return fetchUsersByIDs(ctx, ids)
	},
)

// 2) Obtain a ThunkGetter, ready to be passed to PopulateWithLoader
getter := aggregator.ThunkGetterFromLoader[*OrgUnit, uint32, *User](
	ctx, loader,
	func(o *OrgUnit) uint32 { return o.LeaderID },
)

// 3) Populate: thunks are enqueued during collection and awaited during filling
_ = aggregator.PopulateWithLoader(orgUnits, getter,
	func(o *OrgUnit, u *User) { o.Leader = u })
```

For tree structures use `PopulateTreeWithLoader`, which first collects the thunks of every node recursively and then executes them in one pass.

## Options

| Option | Default | Description |
| --- | --- | --- |
| `WithLimit(limit int)` | `20` | Maximum concurrency; values `<=0` reset to the default of 20 |
| `WithTimeout(d time.Duration)` | `0` (no limit) | Overall execution timeout; when `>0` a deadline is attached to the context |
| `WithRetry(n int)` | `0` | Extra retries per task after a failure |

Retry waiting uses **exponential backoff with random jitter**: it starts at 20ms, doubles each attempt and caps at 2s; the actual delay varies in the range [0.5, 1.5]x of the base value, and the wait can be interrupted by context cancellation. Every task execution (including retries) recovers from panics and converts them into errors instead of crashing the process.

## API Reference

### Parallel execution

- `type ParallelFetcher func(ctx context.Context) error` — the concurrent task contract;
- `func ExecuteParallel(ctx context.Context, fetchers []ParallelFetcher, opts ...Option) error` — runs a slice of tasks in parallel; if any task still fails (retries exhausted or context cancelled), that error is returned.

### Data population

- `type ResourceMap[K comparable, T any] map[K]T` — generic key-value container;
- `type IDGetter[R any, K comparable] func(R) K` / `type IDListGetter[R any, K comparable] func(R) []K` — extract a single / multiple related IDs;
- `type Setter[R any, T any] func(R, T)` / `type MultiSetter[R any, T any] func(R, []T)` — assign a single value / a slice of values;
- `type ChildrenFunc[R any] func(R) []R` — path to child nodes of a tree;
- `func Populate[K, T, R any](items []R, data ResourceMap[K, T], idGetter IDGetter[R, K], setter Setter[R, T])` — populate a flat slice;
- `func PopulateOne[K, T, R any](item R, data ResourceMap[K, T], idGetter IDGetter[R, K], setter Setter[R, T])` — populate a single object;
- `func PopulateMulti[K, T, R any](items []R, data ResourceMap[K, T], idGetter IDListGetter[R, K], setter MultiSetter[R, T])` — populate with multi-ID relations; matched values are collected into a slice in ID order, and the setter is invoked only when at least one value matches;
- `func PopulateTree[K, T, R any](items []R, data map[K]T, idGetter IDGetter[R, K], setter Setter[R, T], children ChildrenFunc[R])` — depth-first population of a tree;
- `func PopulateTreeMulti[K, T, R any](items []R, data map[K]T, idGetter IDListGetter[R, K], setter MultiSetter[R, T], children ChildrenFunc[R])` — multi-ID population of a tree.

### DataLoader-backed population

- `type TypedThunk[T any] func() (T, error)` / `type ThunkGetter[R any, T any] func(R) TypedThunk[T]`;
- `func PopulateWithLoader[R, T any](items []R, thunkGetter ThunkGetter[R, T], setter Setter[R, T]) error` — collects all thunks, then executes them one by one and fills the results;
- `func PopulateTreeWithLoader[R, T any](items []R, thunkGetter ThunkGetter[R, T], setter Setter[R, T], children ChildrenFunc[R]) error` — collects the thunks of the whole tree recursively, then executes and fills them in one pass;
- `func NewLoaderFromFetch[K comparable, V any](fetch func(context.Context, []K) (map[K]V, error)) *dataloader.Loader[K, V]` — wraps a batch-fetch function into a `dataloader.Loader`; if `fetch` returns an error, every key of that batch receives it, and missing keys resolve to the zero value;
- `func ThunkGetterFromLoader[R any, K comparable, V any](ctx context.Context, loader *dataloader.Loader[K, V], keyOf func(R) K) ThunkGetter[R, V]` — adapts a Loader into a `ThunkGetter`; zero-valued keys (e.g. `0`, `""`) short-circuit to a zero-value result without touching the Loader.

## Notes and Best Practices

- **Pre-allocate maps**: build maps inside fetch tasks with `make(ResourceMap, len(ids))` to avoid re-allocation;
- **Filter IDs**: drop invalid IDs such as `0` or empty strings during the Collect phase (`ThunkGetterFromLoader` already short-circuits zero keys, but local-map population relies on the caller to check);
- **Lock-free concurrency**: let each task write its own map variable to run without locks; only use a `sync.Mutex` when multiple tasks must write to the same map;
- **Pass the context**: always forward `ctx` to downstream RPC/SQL calls so timeouts and any task failure cascade-cancel everything;
- **Failures are skipped**: the Populate family silently skips nil items, nil values and (in the DataLoader variants) failed thunks without reporting errors to the caller — log missing data at the Fetch/Thunk layer if you need visibility;
- **Nil detection**: `Populate`/`PopulateWithLoader` use reflection-based nil checks, calling `IsNil` on nillable kinds (pointer, map, slice, func, chan, interface).

## Acknowledgements

- [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [graph-gophers/dataloader](https://github.com/graph-gophers/dataloader)
