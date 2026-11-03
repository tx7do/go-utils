# qqwry

English | [简体中文](README.zh-CN.md)

IPv4 geolocation lookup backed by the CZ88 QQWry IP library. This package is a sub-package of the `github.com/tx7do/go-utils/geoip` module (import path `github.com/tx7do/go-utils/geoip/qqwry`). The data file `qqwry.dat` is embedded via Go `embed` — **it works out of the box, no download required**.

The package implements the parent package's `geoip.GeoIP` interface and returns `geoip.Result`.

## Features

- Embedded QQWry data (`qqwry.dat`, ~10 MB, ~530k IP records), ready to use
- Results are decoded from GB18030 automatically, giving Chinese text directly
- Province and city are split out of the address string automatically, and the ISP is parsed
- Built-in private network checker `IsPrivateIP` and address splitter `SpiltAddress`

> **Note**: this package supports **IPv4 only**, and `Country` is always `"中国"` (see Notes).

## Installation

```bash
go get github.com/tx7do/go-utils/geoip@latest
```

```go
import "github.com/tx7do/go-utils/geoip/qqwry"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip/qqwry"
)

func main() {
	c := qqwry.NewClient()

	res, err := c.Query("47.108.149.89")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// Output: {IP:47.108.149.89 Country:中国 Province:四川省 City:成都市 ISP:阿里云}
}
```

## API

### NewClient

```go
func NewClient() *Client
```

Creates a client from the embedded `qqwry.dat` and initializes the index range. It returns no error; the data ships with the binary and `Client` has no `Close`, so there is nothing to release manually.

### Client.IPNum

```go
IPNum int64
```

Total number of IP records in the data file index (~530k for the current embedded data).

### (*Client) Query

```go
func (c *Client) Query(queryIp string) (geoip.Result, error)
```

Queries the location for an IPv4 address, implementing the `geoip.GeoIP` interface.

- Accepts IPv4 only; IPv6 or invalid strings return an `ip is not ipv4` error;
- Private/loopback IPv4 (including `127.0.0.1`) skips the data file and returns `Province = City = "局域网"`;
- Returns an `ip not found` error when the index has no matching record;
- `Result.IP` is filled with the queried IP; `Country` is always `"中国"`.

### IsPrivateIP

```go
func IsPrivateIP(ipStr string) bool
```

Reports whether an IP string is a private address. **Note this differs from the same-named function in [geolite](../geolite)**: this one takes a `string`, and any input containing `:` (IPv6) or failing to parse returns `true` (treated as private). IPv4 private ranges: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `127.0.0.0/8`, `169.254.0.0/16`.

### SpiltAddress

```go
func SpiltAddress(addr string) []string
```

Splits an address string by administrative suffixes (省, 市, 自治区, 自治州, 盟, 县, 区, 管委会, 街道, 镇, 乡), e.g. `"中国广东省深圳市南山区"` → `["中国广东省", "深圳市", "南山区"]`. `Query` uses it internally to extract province and city.

## Notes

- **IPv4 only**: IPv6 input returns an error immediately; use [geolite](../geolite) or [ip2region](../ip2region) for IPv6 lookups.
- **`Country` is always `"中国"`**: `Query` hardcodes `Country = "中国"` up front. When querying a non-Chinese IP, its country name (e.g. `"美国"`) ends up in the `City` field and the detailed location in the `ISP` field — interpret results accordingly.
- **Private-address placeholder**: private/loopback IPv4 never touches the data file; `Province` and `City` return the `"局域网"` placeholder. Because the private check runs before the record lookup, `127.0.0.1` also returns the `"局域网"` placeholder rather than a record from the data file.
- **Result language**: the data source is the Chinese QQWry library, so results are always Chinese; the placeholder ISP `"CZ88.NET"` is cleared automatically.

## Data updates

The packaged data is a build-time snapshot. To get fresh data, obtain a new `qqwry.dat` from the sources below (the package does not yet offer an external data-loading API; to use new data, replace `assets/qqwry.dat` and rebuild):

- [CZ88 community IP library](https://update.cz88.net/geo-public)
- [qqwry.dat - out0fmemory - GitHub](https://github.com/out0fmemory/qqwry.dat/blob/master/qqwry_lastest.dat)
- [qqwry - xiaoqidun - GitHub](https://github.com/xiaoqidun/qqwry)
- [qqwry - freshcn - GitHub](https://github.com/freshcn/qqwry)

## Acknowledgements

- [CZ88 (纯真网络)](https://update.cz88.net/geo-public) — source of the qqwry.dat data
