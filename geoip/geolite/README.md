# geolite

English | [简体中文](README.zh-CN.md)

IP geolocation lookup backed by the MaxMind GeoLite2 database. This package is a sub-package of the `github.com/tx7do/go-utils/geoip` module (import path `github.com/tx7do/go-utils/geoip/geolite`). The GeoLite2-City database is embedded via Go `embed`, so it works out of the box with no data file to provision.

The package implements the parent package's `geoip.GeoIP` interface and returns `geoip.Result`.

## Features

- Embedded GeoLite2-City database (`GeoLite2-City.mmdb`, ~60 MB), ready to use
- IPv4 and IPv6 queries supported
- Output language configurable via `SetLanguage` (default `zh-CN`)
- Built-in private network checker `IsPrivateIP` for IPv4 and IPv6

> **Key clarification**: this package embeds only the **GeoLite2-City (city-level)** database — **not** the Country or ASN databases. City queries return country, province and city information, but there is **no ISP data**, so `geoip.Result.ISP` is always empty. For ISP information use [qqwry](../qqwry) or [ip2region](../ip2region).

## Installation

```bash
go get github.com/tx7do/go-utils/geoip@latest
```

```go
import "github.com/tx7do/go-utils/geoip/geolite"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/geoip/geolite"
)

func main() {
	g, err := geolite.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = g.Close() }()

	ret, err := g.Query("47.108.149.89")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", ret)
	// Output: {IP: Country:中国 Province:四川 City:成都 ISP:}
}
```

## API

### NewClient

```go
func NewClient() (*Client, error)
```

Creates a client from the embedded GeoLite2-City database; returns an error if the database fails to load.

### (*Client) Close

```go
func (g *Client) Close() error
```

Closes the client and releases the database handle.

### (*Client) SetLanguage

```go
func (g *Client) SetLanguage(code string)
```

Sets the language code used for query results; the default is `zh-CN`. The code maps to the name table built into the MaxMind database (e.g. `en`, `zh-CN`); if a field has no name in the requested language, it is left empty.

### (*Client) Query

```go
func (g *Client) Query(rawIP string) (geoip.Result, error)
```

Queries the location for an IP, implementing the `geoip.GeoIP` interface.

- Returns an `invalid ip address` error if `rawIP` is not a valid IP;
- Private/loopback addresses skip the database and return a placeholder result (see Notes);
- `Result.ISP` is always empty and `Result.IP` is not filled in.

### IsPrivateIP

```go
func IsPrivateIP(ip net.IP) bool
```

Reports whether the IP is a private address. Supports IPv4 and IPv6 private ranges: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `127.0.0.0/8`, `169.254.0.0/16`, `fc00::/7`, `fe80::/10`, `::1/128`, `fec0::/10`.

## Notes

- **Private-address placeholder**: queries for private/loopback IPs (IPv4 and IPv6) never touch the database; `Country`, `Province` and `City` are all set to the placeholder string `"局域网"` and `err` is `nil`.
- **Query failure terminates the process**: when the underlying database query fails, this package calls kratos log's `log.Fatal(err)`, which **exits the process immediately**. Make sure you pass valid, non-private IPs.
- **City database only**: GeoLite2-Country / GeoLite2-ASN are not embedded, and there is no ISP information.
- **Data freshness**: the embedded data is updated with releases of this library. The upstream GeoLite2 data is refreshed twice a week; use the data sources below if you need the latest data.

## Acknowledgements

- [MaxMind GeoLite2 free data](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- Data download/update references:
  1. <https://github.com/P3TERX/GeoLite.mmdb/releases>
  2. <https://dev.maxmind.com/geoip/geolite2-free-geolocation-data>
