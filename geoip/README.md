# geoip

English | [简体中文](README.zh-CN.md)

IP geolocation query module. It defines a unified `GeoIP` interface and a `Result` struct, and ships three sub-packages with embedded data files that work out of the box: [geolite](./geolite) (MaxMind GeoLite2), [qqwry](./qqwry) (CZ88 QQWry library) and [ip2region](./ip2region) (dual-stack IPv4/IPv6).

## Features

- Unified `GeoIP` interface and `Result` struct — switch data sources without touching business code
- Three sub-package implementations covering domestic and international IP lookup scenarios
- All data files are embedded via Go `embed`: a single compiled binary runs standalone, no external data deployment required
- IPv4 / IPv6 support varies by data source (see the table below)

## Installation

```bash
go get github.com/tx7do/go-utils/geoip@latest
```

This directory is a standalone Go module (module path `github.com/tx7do/go-utils/geoip`). `geolite`, `qqwry` and `ip2region` are sub-packages of this module (they have no `go.mod` of their own); installing the module gives you all of them:

```go
import (
	"github.com/tx7do/go-utils/geoip"
	"github.com/tx7do/go-utils/geoip/geolite"
	"github.com/tx7do/go-utils/geoip/ip2region"
	"github.com/tx7do/go-utils/geoip/qqwry"
)
```

## Sub-package comparison

| Sub-package | Data source | Embedded data | IPv4 | IPv6 | ISP field | Notes |
| -------- | -------- | -------- | :--: | :--: | :--: | -------- |
| [geolite](./geolite) | MaxMind GeoLite2 | GeoLite2-City.mmdb (~60 MB) | ✓ | ✓ | ✗ | Only the City database is embedded; output language configurable (default `zh-CN`) |
| [qqwry](./qqwry) | CZ88 community IP library | qqwry.dat (~10 MB) | ✓ | ✗ | ✓ | Chinese results; `Country` is always `"中国"` |
| [ip2region](./ip2region) | ip2region xdb | ip2region_v4.xdb + ip2region_v6.xdb (~50 MB total) | ✓ | ✓ | ✓ | Dual-stack; v4/v6 can be enabled or disabled independently |

Note: the implementations fill `Result.IP` inconsistently — qqwry writes the queried IP into `Result.IP`, while geolite and ip2region do not.

## Quick Start

Using qqwry as an example (all three sub-packages' `Client` implements `geoip.GeoIP`, so they are directly interchangeable). See each sub-package README for details:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip"
	"github.com/tx7do/go-utils/geoip/qqwry"
)

func main() {
	var client geoip.GeoIP = qqwry.NewClient()

	res, err := client.Query("47.108.149.89")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", res)
	// Output: {IP:47.108.149.89 Country:中国 Province:四川省 City:成都市 ISP:阿里云}
}
```

## API

### Result

```go
type Result struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`  // Country
	Province string `json:"province"` // Province
	City     string `json:"city"`     // City
	ISP      string `json:"isp"`      // Internet service provider
}
```

### GeoIP interface

```go
type GeoIP interface {
	Query(queryIp string) (res Result, err error)
}
```

### Sub-package docs

- [geoip/geolite](./geolite/README.md) ([中文](./geolite/README.zh-CN.md)) — MaxMind GeoLite2-City
- [geoip/qqwry](./qqwry/README.md) ([中文](./qqwry/README.zh-CN.md)) — QQWry library (IPv4)
- [geoip/ip2region](./ip2region/README.md) ([中文](./ip2region/README.zh-CN.md)) — ip2region xdb (IPv4/IPv6)

## Notes

- The embedded data significantly increases binary size (geolite ~60 MB, ip2region ~50 MB, qqwry ~10 MB) and is loaded into memory when the client is initialized — pick sub-packages as needed.
- Private/loopback addresses: geolite fills `Country/Province/City` with `"局域网"`; qqwry fills `Province/City` with `"局域网"`. See each README for details.
- The embedded data is refreshed when this library is released; for the latest data, follow the data source links in each sub-package README.

## Acknowledgements

- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [CZ88 community IP library](https://update.cz88.net/geo-public)
- [ip2region](https://github.com/lionsoul2014/ip2region)
