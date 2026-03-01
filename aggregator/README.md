# Package aggregator

`aggregator` 是一个高性能的分布式数据聚合工具包，旨在解决微服务架构中的 **N+1 查询问题** 以及 **DTO 关联数据回填（Data Populating）** 的复杂性。

## 核心特性

- 🚀 **高性能并行 Fetch**：基于 `errgroup` 实现任务级并发，支持 `Context` 取消传播与错误捕获。
- 🧬 **强类型泛型 Populate**：利用 Go 泛型实现声明式数据回填，支持扁平切片与递归树结构。
- 🛡️ **类型安全**：消除 `any` 类型断言，在编译期拦截类型错误。
- 🍃 **轻量无依赖**：仅依赖核心标准库与 `x/sync`。

## 核心概念

1. **Collect (收集)**：从原始数据源中提取所有需要关联查询的唯一 ID。
2. **Fetch (抓取)**：并发调用外部微服务（RPC/SQL）获取数据。
3. **Populate (回填)**：将获取到的元数据映射并注入到原始 DTO 对象中。

## 快速开始

### 1. 并行 Fetch 数据

使用 `ExecuteParallel` 隐藏复杂的锁管理与等待逻辑。

```go
import "your-project/pkg/aggregator"

func (s *Service) GetDetails(ctx context.Context, units []*OrgUnit) error {
    var (
        userMap = make(aggregator.ResourceMap[uint32, *User])
        deptMap = make(aggregator.ResourceMap[uint32, *Dept])
    )

    err := aggregator.ExecuteParallel(ctx,
        func(ctx context.Context) error {
            // 执行 User 抓取逻辑并写入 userMap
            return nil
        },
        func(ctx context.Context) error {
            // 执行 Dept 抓取逻辑并写入 deptMap
            return nil
        },
    )
    return err
}
```

### 2. 声明式回填 (Populate)

支持递归树状结构，通过闭包定义提取与赋值逻辑。

```go
// 回填组织架构树中的 Leader 名称
aggregator.PopulateTree(
    orgUnits,                                     // 原始对象列表
    userMap,                                      // 数据源 Map
    func(o *OrgUnit) uint32 { return o.LeaderId },// 获取 ID 的逻辑
    func(o *OrgUnit, u *User) { o.LeaderName = &u.Name }, // 执行赋值的逻辑
    func(o *OrgUnit) []*OrgUnit { return o.Children },    // 递归路径
)
```

## API 参考

### ExecuteParallel

`func ExecuteParallel(ctx context.Context, fetchers ...ParallelFetcher) error`

并发执行多个任务。只要其中一个任务返回错误，整个过程将停止并返回该错误。

### Populate
`func Populate[K comparable, T any, R any](items []R, data ResourceMap[K, T], idGetter func(R) K, setter func(R, T))`

处理扁平切片的字段填充。

### PopulateTree

`func PopulateTree[K comparable, T any, R any](items []R, data ResourceMap[K, T], idGetter func(R) K, setter func(R, T), children func(R) []R)`

专门用于处理树形结构（如组织架构、评论树）的深度优先填充。

## 最佳实践

- **Map 预分配**：在 `Fetch` 任务构建 `Map` 时，请始终使用 `make(ResourceMap, len(ids))` 以减少内存重新分配。
- **ID 过滤**：在 `Collect` 阶段请务必过滤 `0` 或空字符串 ID，避免发起无效的 RPC 请求。
- **无锁并发**：在 `ExecuteParallel` 中，建议每个任务写自己独立的 Map 变量。只有当多个任务必须写入同一个 `Map` 时才使用 `sync.Mutex`。
- **Context 穿透**：始终将 `ctx` 传递给下游 RPC 调用，以便在超时或任一任务失败时实现级联取消。
