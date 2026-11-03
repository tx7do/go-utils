# ip2region

[English](README.md) | 简体中文

基于 [ip2region](https://github.com/lionsoul2014/ip2region) 离线 xdb 数据库的 IP 归属地查询包，是 `github.com/tx7do/go-utils/geoip` 模块的子包（import 路径 `github.com/tx7do/go-utils/geoip/ip2region`）。IPv4 与 IPv6 两个 xdb 数据文件均已内嵌，开箱即用。

本包提供两个使用层级：

- **`Client`**（推荐）：实现父包的 `geoip.GeoIP` 接口，查询结果解析为 `geoip.Result`；
- **`Ip2Region`**（底层）：可自定义缓存策略与检索器数量，返回 xdb 原始区域字符串。

## Features

- 内置 IPv4（`ip2region_v4.xdb`，约 11 MB）与 IPv6（`ip2region_v6.xdb`，约 39 MB）双数据库，原生双栈查询
- 三种缓存策略：`NoCache` / `VIndexCache` / `BufferCache`
- v4 / v6 可独立启用或禁用（配置传 `nil` 即禁用对应版本）
- 池化检索器（`SearcherPool`）并发复用，并发安全

## Installation

```bash
go get github.com/tx7do/go-utils/geoip@latest
```

```go
import "github.com/tx7do/go-utils/geoip/ip2region"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip/ip2region"
)

func main() {
	c, err := ip2region.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// IPv4 查询
	res, err := c.Query("113.92.157.29")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// 输出: {IP: Country:中国 Province:广东省 City:深圳市 ISP:电信}

	// IPv6 查询
	res, err = c.Query("240e:3b7:3272:d8d0:db09:c067:8d59:539e")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// 输出: {IP: Country:中国 Province:广东省 City:深圳市 ISP:电信}
}
```

### 进阶：自定义配置（Ip2Region）

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip/ip2region"
	"github.com/tx7do/go-utils/geoip/ip2region/assets"
)

func main() {
	// 仅启用 IPv4：整库缓存（BufferCache）、单个检索器
	v4Config, err := ip2region.NewV4Config(ip2region.BufferCache, assets.Ip2RegionV4, 1)
	if err != nil {
		log.Fatal(err)
	}

	// v6 配置传 nil 表示禁用 IPv6
	svc, err := ip2region.NewIp2Region(v4Config, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// 返回 xdb 原始区域字符串（以 "|" 分段），例如: 中国|广东省|深圳市|电信
	region, err := svc.SearchByStr("113.92.157.29")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(region)
}
```

## 数据加载方式

数据文件通过 Go `embed` 内嵌在本包的 [assets](assets/) 子包中（`assets.Ip2RegionV4`、`assets.Ip2RegionV6`），以内存字节形式提供：

- `NewClient()` 内部即为 v4/v6 分别以 `VIndexCache` 策略、20 个检索器构建配置，再创建 `Ip2Region` 服务；
- `NewV4Config` / `NewV6Config` 会解析并校验传入的 xdb 字节流头部，确保数据版本与目标 IP 版本一致（v4 配置必须搭配 v4 xdb，反之亦然）；
- 本包不接受磁盘路径；如需使用自备的 xdb 文件，自行读文件后将字节切片传给 `NewV4Config` / `NewV6Config`。

## 缓存策略

| 常量 | 值 | 行为 |
| -------- | :--: | -------- |
| `NoCache` | 0 | 不预加载向量索引；检索器池方式 |
| `VIndexCache` | 1 | 向 `Config` 预加载向量索引（可通过 `Config.VIndex()` 获取）；检索器池方式 |
| `BufferCache` | 2 | 整库缓存，使用单个常驻内存检索器 |

- 所有策略的数据均来自内嵌字节，查询全程在内存中完成；
- `BufferCache` 之外的模式都会创建含 `searchers` 个检索器的 `SearcherPool`，查询时借用（`BorrowSearcher`）、用毕归还（`ReturnSearcher`）；
- `CachePolicyFromName(name string) (int, error)` 支持从字符串解析策略：`"file"` / `"nocache"` → `NoCache`；`"vectorindex"` / `"vindex"` / `"vindexcache"` → `VIndexCache`；`"content"` / `"buffercache"` → `BufferCache`。

## API

### Client（geoip.GeoIP 封装）

```go
func NewClient() (*Client, error)
func (g *Client) Query(rawIP string) (geoip.Result, error)
func (g *Client) Close()
```

- `NewClient`：v4/v6 双栈全启用，`VIndexCache` 策略，每栈 20 个检索器；
- `Query`：调用底层 `SearchByStr`，把结果按 `|` 拆分为 4 段并映射为 `geoip.Result`（`Country`/`Province`/`City`/`ISP`）；段数不为 4 时返回 `invalid region data` 错误；`Result.IP` 不回填；
- `Close`：关闭底层服务（默认 10 秒等待超时）。

### Ip2Region（底层服务）

```go
func NewIp2Region(v4Config *Config, v6Config *Config) (*Ip2Region, error)
func (ip2r *Ip2Region) SearchByStr(ipStr string) (string, error)
func (ip2r *Ip2Region) Search(ipBytes []byte) (string, error)
func (ip2r *Ip2Region) Close()
func (ip2r *Ip2Region) CloseTimeout(d time.Duration)
```

- 配置传 `nil` 即禁用对应 IP 版本；查询被禁用的版本返回空字符串且**无错误**；
- `Search` / `SearchByStr` 按解析后的 IP 字节长度路由：4 字节走 v4，16 字节走 v6，其余长度报错；
- `Close` 以 10 秒超时等待检索器归还；`CloseTimeout` 可自定义等待时长。

### Config

```go
func NewV4Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error)
func NewV6Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error)
```

- `searchers` 必须 > 0，否则报错；
- 解析并校验 `xdbContent` 头部，核对 IP 版本是否匹配；
- 只读访问器：`CachePolicy()`、`IPVersion()`、`Header()`、`VIndex()`、`CBuffer()`、`Searchers()`。

### 缓存策略常量与解析

```go
const (
	NoCache     = 0
	VIndexCache = 1
	BufferCache = 2
)

func CachePolicyFromName(name string) (int, error)
```

### SearcherPool

```go
func NewSearcherPool(config *Config) (*SearcherPool, error)
func (sp *SearcherPool) LoanCount() int
func (sp *SearcherPool) BorrowSearcher() *xdb.Searcher
func (sp *SearcherPool) ReturnSearcher(searcher *xdb.Searcher)
func (sp *SearcherPool) Close()
func (sp *SearcherPool) CloseTimeout(d time.Duration)
```

## Notes

- 内嵌数据合计约 50 MB，会相应增大二进制体积并常驻内存。
- 上游 `xdb.Searcher` 的实现注释标明其**非线程安全**：本包用检索器池解决并发安全；但 `BufferCache` 是所有查询共享单个检索器，高并发场景建议改用池化策略（`NoCache` / `VIndexCache`，`NewClient` 默认即 `VIndexCache`）。
- 查询被禁用（配置为 `nil`）的 IP 版本会静默返回空结果，不会报错，调用方需自行处理。
- xdb 数据格式与生成器见上游 ip2region 项目；本包未附带数据制作工具。

## Acknowledgements

- [ip2region](https://github.com/lionsoul2014/ip2region) — 数据与 xdb 检索实现来源（作者 Lion，Apache-2.0）
- [ip2region.net](https://ip2region.net/)
