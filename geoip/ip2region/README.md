# ip2region

English | [简体中文](README.zh-CN.md)

IP geolocation lookup backed by the offline [ip2region](https://github.com/lionsoul2014/ip2region) xdb databases. This package is a sub-package of the `github.com/tx7do/go-utils/geoip` module (import path `github.com/tx7do/go-utils/geoip/ip2region`). Both the IPv4 and IPv6 xdb data files are embedded, so it works out of the box.

The package offers two levels of API:

- **`Client`** (recommended): implements the parent package's `geoip.GeoIP` interface and parses results into `geoip.Result`;
- **`Ip2Region`** (low-level): customizable cache policy and searcher count, returning the raw xdb region string.

## Features

- Embedded dual databases: IPv4 (`ip2region_v4.xdb`, ~11 MB) and IPv6 (`ip2region_v6.xdb`, ~39 MB), native dual-stack queries
- Three cache policies: `NoCache` / `VIndexCache` / `BufferCache`
- v4 / v6 can be enabled or disabled independently (pass `nil` config to disable one)
- Pooled searchers (`SearcherPool`) reused across concurrent queries, concurrency-safe

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

	// IPv4 query
	res, err := c.Query("113.92.157.29")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// Output: {IP: Country:中国 Province:广东省 City:深圳市 ISP:电信}

	// IPv6 query
	res, err = c.Query("240e:3b7:3272:d8d0:db09:c067:8d59:539e")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// Output: {IP: Country:中国 Province:广东省 City:深圳市 ISP:电信}
}
```

### Advanced: custom configuration (Ip2Region)

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip/ip2region"
	"github.com/tx7do/go-utils/geoip/ip2region/assets"
)

func main() {
	// IPv4 only: whole-database cache (BufferCache) with a single searcher
	v4Config, err := ip2region.NewV4Config(ip2region.BufferCache, assets.Ip2RegionV4, 1)
	if err != nil {
		log.Fatal(err)
	}

	// passing nil for the v6 config disables IPv6
	svc, err := ip2region.NewIp2Region(v4Config, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// returns the raw xdb region string ("|"-separated), e.g.: 中国|广东省|深圳市|电信
	region, err := svc.SearchByStr("113.92.157.29")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(region)
}
```

## How the data is loaded

The data files are embedded via Go `embed` in this package's [assets](assets/) sub-package (`assets.Ip2RegionV4`, `assets.Ip2RegionV6`) and provided as in-memory bytes:

- `NewClient()` internally builds a `VIndexCache` config with 20 searchers for each of v4/v6, then creates the `Ip2Region` service;
- `NewV4Config` / `NewV6Config` parse and validate the xdb header of the given bytes, ensuring the data version matches the target IP version (a v4 config requires v4 xdb and vice versa);
- The package does not accept file paths; to use your own xdb files, read them yourself and pass the bytes to `NewV4Config` / `NewV6Config`.

## Cache policies

| Constant | Value | Behavior |
| -------- | :--: | -------- |
| `NoCache` | 0 | No vector index preloading; searcher-pool mode |
| `VIndexCache` | 1 | Preloads the vector index into the `Config` (accessible via `Config.VIndex()`); searcher-pool mode |
| `BufferCache` | 2 | Whole-database cache with a single resident in-memory searcher |

- Under every policy the data comes from the embedded bytes and queries are served fully from memory;
- All modes other than `BufferCache` create a `SearcherPool` holding `searchers` searcher instances; queries borrow (`BorrowSearcher`) and return (`ReturnSearcher`) them;
- `CachePolicyFromName(name string) (int, error)` parses a policy from a string: `"file"` / `"nocache"` → `NoCache`; `"vectorindex"` / `"vindex"` / `"vindexcache"` → `VIndexCache`; `"content"` / `"buffercache"` → `BufferCache`.

## API

### Client (geoip.GeoIP wrapper)

```go
func NewClient() (*Client, error)
func (g *Client) Query(rawIP string) (geoip.Result, error)
func (g *Client) Close()
```

- `NewClient`: both v4/v6 enabled, `VIndexCache` policy, 20 searchers per stack;
- `Query`: calls the underlying `SearchByStr`, splits the result by `|` into exactly 4 segments and maps them to `geoip.Result` (`Country`/`Province`/`City`/`ISP`); returns an `invalid region data` error when the segment count is not 4; `Result.IP` is not filled in;
- `Close`: closes the underlying service (10-second default wait timeout).

### Ip2Region (low-level service)

```go
func NewIp2Region(v4Config *Config, v6Config *Config) (*Ip2Region, error)
func (ip2r *Ip2Region) SearchByStr(ipStr string) (string, error)
func (ip2r *Ip2Region) Search(ipBytes []byte) (string, error)
func (ip2r *Ip2Region) Close()
func (ip2r *Ip2Region) CloseTimeout(d time.Duration)
```

- Passing `nil` for a config disables that IP version; querying a disabled version returns an empty string **without an error**;
- `Search` / `SearchByStr` route by the parsed IP byte length: 4 bytes go to v4, 16 bytes to v6, anything else is an error;
- `Close` waits up to 10 seconds for borrowed searchers to return; `CloseTimeout` accepts a custom duration.

### Config

```go
func NewV4Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error)
func NewV6Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error)
```

- `searchers` must be > 0, otherwise an error is returned;
- Parses and validates the `xdbContent` header and checks that the IP version matches;
- Read-only accessors: `CachePolicy()`, `IPVersion()`, `Header()`, `VIndex()`, `CBuffer()`, `Searchers()`.

### Cache policy constants and parsing

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

- The embedded data totals ~50 MB, proportionally increasing binary size and staying resident in memory.
- The upstream `xdb.Searcher` is documented as **not thread-safe**: this package solves concurrency with the searcher pool, but `BufferCache` shares a single searcher across all queries — under high concurrency prefer a pooled policy (`NoCache` / `VIndexCache`; `NewClient` uses `VIndexCache` by default).
- Querying a disabled (nil-config) IP version silently returns an empty result without an error; handle this in your caller.
- The xdb data format and generator tools live in the upstream ip2region project; this package ships no data-making tools.

## Acknowledgements

- [ip2region](https://github.com/lionsoul2014/ip2region) — source of the data and xdb searcher implementation (by Lion, Apache-2.0)
- [ip2region.net](https://ip2region.net/)
