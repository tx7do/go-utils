# aggregator

[English](README.md) | 简体中文

`aggregator` 是一个高性能的数据聚合与关联数据回填工具包，旨在解决微服务架构中的 **N+1 查询问题**与 **DTO 关联数据填充（Data Populating）** 的复杂性。

## 特性

- **并行 Fetch**：`ExecuteParallel` 基于 `errgroup` 实现任务级并发，支持并发数限制、整体超时、失败重试（指数退避 + 抖动）与 panic 恢复；
- **强类型泛型回填**：利用 Go 泛型实现声明式数据回填，覆盖单个对象、扁平切片与递归树结构，支持"一个 ID"与"多个 ID"两种关联形态；
- **DataLoader 集成**：提供 `NewLoaderFromFetch` / `ThunkGetterFromLoader` 将按批抓取函数封装为 [`graph-gophers/dataloader`](https://github.com/graph-gophers/dataloader) 的 Loader，并配合 `PopulateWithLoader` / `PopulateTreeWithLoader` 完成请求内合并批量的回填；
- **类型安全**：消除 `any` 类型断言，在编译期拦截类型错误。

## 安装

```bash
go get github.com/tx7do/go-utils/aggregator
```

```go
import "github.com/tx7do/go-utils/aggregator"
```

依赖：Go 标准库、`golang.org/x/sync`、`github.com/graph-gophers/dataloader/v7`（仅 DataLoader 相关 API 使用）。

## 核心概念

1. **Collect（收集）**：从原始数据中提取所有需要关联查询的唯一 ID；
2. **Fetch（抓取）**：并发调用外部服务（RPC/SQL）批量获取数据；
3. **Populate（回填）**：将获取到的关联数据注入回原始 DTO 对象。

## 快速开始

### 1. 并行 Fetch 数据

使用 `ExecuteParallel` 并发执行多个抓取任务，隐藏锁管理与等待逻辑。任务以**切片**传入：

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
				// 执行 User 抓取逻辑并写入 userMap
				return nil
			},
			func(ctx context.Context) error {
				// 执行 Dept 抓取逻辑并写入 deptMap
				return nil
			},
		},
		aggregator.WithLimit(10),                      // 最大并发数（默认 20）
		aggregator.WithTimeout(5*time.Second),          // 整体超时（默认不限）
		aggregator.WithRetry(2),                        // 失败重试次数（默认 0）
	)
}
```

任一任务返回错误时，`errgroup` 派生的 context 会被取消，其余任务尽快退出，`ExecuteParallel` 返回该错误。空任务切片直接返回 `nil`，`nil` 任务被跳过。

### 2. 声明式回填（Populate）

通过闭包声明"取 ID"与"赋值"逻辑。注意 `PopulateTree` 的数据源参数是普通的 `map[K]T`：

```go
// 回填组织架构树中的 Leader 信息
aggregator.PopulateTree(
	orgUnits,                                              // 原始对象列表
	userMap,                                               // 数据源 map[uint32]*User
	func(o *OrgUnit) uint32 { return o.LeaderID },         // 取关联 ID
	func(o *OrgUnit, u *User) { o.LeaderName = u.Name },   // 执行赋值
	func(o *OrgUnit) []*OrgUnit { return o.Children },     // 递归子节点路径
)
```

### 3. DataLoader 批量回填

适合同一请求内大量"按 ID 取关联对象"的场景，多个 Load 自动合并为一个批量请求：

```go
// 1) 将批量抓取函数封装为 Loader
loader := aggregator.NewLoaderFromFetch[uint32, *User](
	func(ctx context.Context, ids []uint32) (map[uint32]*User, error) {
		// 一次 RPC/SQL 按 ids 批量查询，返回 map
		return fetchUsersByIDs(ctx, ids)
	},
)

// 2) 得到 ThunkGetter，可直接传给 PopulateWithLoader
getter := aggregator.ThunkGetterFromLoader[*OrgUnit, uint32, *User](
	ctx, loader,
	func(o *OrgUnit) uint32 { return o.LeaderID },
)

// 3) 回填：Thunk 在收集阶段入批，在回填阶段统一等待结果
_ = aggregator.PopulateWithLoader(orgUnits, getter,
	func(o *OrgUnit, u *User) { o.Leader = u })
```

树状结构请使用 `PopulateTreeWithLoader`，它会先递归收集所有节点的 Thunk 再统一执行。

## 选项（Options）

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| `WithLimit(limit int)` | `20` | 最大并发数；传入 `<=0` 时重置为默认值 20 |
| `WithTimeout(d time.Duration)` | `0`（不限时） | 整体执行的超时时间；`>0` 时为 ctx 附加超时 |
| `WithRetry(n int)` | `0` | 每个任务失败后的额外重试次数 |

重试等待采用**指数退避 + 随机抖动**：初始 20ms、每次翻倍、上限 2s，实际延迟在基准值的 [0.5, 1.5] 倍间抖动，退避等待可被 ctx 取消。每次任务执行（含重试）都有 panic 恢复，panic 被转换为 error 返回而不会击穿进程。

## API 参考

### 并行执行

- `type ParallelFetcher func(ctx context.Context) error` —— 并发任务契约；
- `func ExecuteParallel(ctx context.Context, fetchers []ParallelFetcher, opts ...Option) error` —— 并行执行任务切片，任一任务最终失败（重试耗尽或 ctx 取消）即返回该错误。

### 数据回填

- `type ResourceMap[K comparable, T any] map[K]T` —— 通用键值容器；
- `type IDGetter[R any, K comparable] func(R) K` / `type IDListGetter[R any, K comparable] func(R) []K` —— 取单个 / 多个关联 ID；
- `type Setter[R any, T any] func(R, T)` / `type MultiSetter[R any, T any] func(R, []T)` —— 单值 / 多值赋值；
- `type ChildrenFunc[R any] func(R) []R` —— 树结构的子节点路径；
- `func Populate[K, T, R any](items []R, data ResourceMap[K, T], idGetter IDGetter[R, K], setter Setter[R, T])` —— 扁平切片回填；
- `func PopulateOne[K, T, R any](item R, data ResourceMap[K, T], idGetter IDGetter[R, K], setter Setter[R, T])` —— 单对象回填；
- `func PopulateMulti[K, T, R any](items []R, data ResourceMap[K, T], idGetter IDListGetter[R, K], setter MultiSetter[R, T])` —— 多 ID 关联回填，取到的值按 ID 顺序组成切片；仅当至少命中一个时才调用 setter；
- `func PopulateTree[K, T, R any](items []R, data map[K]T, idGetter IDGetter[R, K], setter Setter[R, T], children ChildrenFunc[R])` —— 树状结构深度优先回填；
- `func PopulateTreeMulti[K, T, R any](items []R, data map[K]T, idGetter IDListGetter[R, K], setter MultiSetter[R, T], children ChildrenFunc[R])` —— 树状结构多 ID 回填。

### DataLoader 回填

- `type TypedThunk[T any] func() (T, error)` / `type ThunkGetter[R any, T any] func(R) TypedThunk[T]`；
- `func PopulateWithLoader[R, T any](items []R, thunkGetter ThunkGetter[R, T], setter Setter[R, T]) error` —— 收集所有 Thunk 后逐个执行并回填；
- `func PopulateTreeWithLoader[R, T any](items []R, thunkGetter ThunkGetter[R, T], setter Setter[R, T], children ChildrenFunc[R]) error` —— 先递归收集整棵树的 Thunk，再统一执行回填；
- `func NewLoaderFromFetch[K comparable, V any](fetch func(context.Context, []K) (map[K]V, error)) *dataloader.Loader[K, V]` —— 将按批抓取函数封装为 `dataloader.Loader`；`fetch` 返回 error 时该批所有 key 都得到该错误，未命中的 key 得到零值；
- `func ThunkGetterFromLoader[R any, K comparable, V any](ctx context.Context, loader *dataloader.Loader[K, V], keyOf func(R) K) ThunkGetter[R, V]` —— 把 Loader 适配为 `ThunkGetter`；零值键（如 `0`、`""`）直接返回零值结果，不触发 Loader。

## 注意事项与最佳实践

- **Map 预分配**：在 Fetch 任务构建 Map 时用 `make(ResourceMap, len(ids))` 预分配容量，减少重分配；
- **ID 过滤**：Collect 阶段过滤 `0`、空字符串等无效 ID（`ThunkGetterFromLoader` 对零值键已做短路，但本地 Map 回填依赖调用方自查）；
- **无锁并发**：每个任务写各自独立的 Map 变量即可无锁并行；多个任务必须写同一个 Map 时才加 `sync.Mutex`；
- **Context 穿透**：始终把 `ctx` 传给下游 RPC/SQL 调用，以便超时或任一任务失败时级联取消；
- **失败即跳过**：Populate 系列对 nil 项、nil 值与（DataLoader 版）出错的 Thunk 一律静默跳过，不会向调用方报错——需要感知缺失数据时请在 Fetch/Thunk 层记录；
- **nil 判定**：`Populate`/`PopulateWithLoader` 使用 reflect 判空，对指针、map、slice、func、chan、interface 等可 nil 类型做 `IsNil` 检查。

## 致谢

- [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [graph-gophers/dataloader](https://github.com/graph-gophers/dataloader)
