# geolite

[English](README.md) | 简体中文

基于 MaxMind GeoLite2 数据库的 IP 归属地查询包，是 `github.com/tx7do/go-utils/geoip` 模块的子包（import 路径 `github.com/tx7do/go-utils/geoip/geolite`）。GeoLite2-City 数据库已通过 Go `embed` 内嵌，开箱即用，无需自备数据文件。

本包实现父包的 `geoip.GeoIP` 接口，查询结果为 `geoip.Result`。

## Features

- 内置 GeoLite2-City 数据库（`GeoLite2-City.mmdb`，约 60 MB），开箱即用
- 支持 IPv4 与 IPv6 查询
- 查询结果语言可配置（`SetLanguage`），默认 `zh-CN`
- 内置 IPv4/IPv6 私有网段判断工具 `IsPrivateIP`

> **关键澄清**：本包只内置 **GeoLite2-City（城市级）** 一个数据库，**不包含** Country / ASN 版本。City 库的查询结果包含国家、省、城市信息，但**没有 ISP（运营商）数据**，`geoip.Result.ISP` 恒为空。如需 ISP 信息，请使用 [qqwry](../qqwry) 或 [ip2region](../ip2region)。

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
	// 输出: {IP: Country:中国 Province:四川 City:成都 ISP:}
}
```

## API

### NewClient

```go
func NewClient() (*Client, error)
```

从内嵌的 GeoLite2-City 数据库创建客户端；数据库解析失败时返回错误。

### (*Client) Close

```go
func (g *Client) Close() error
```

关闭客户端，释放数据库句柄。

### (*Client) SetLanguage

```go
func (g *Client) SetLanguage(code string)
```

设置查询结果的语言代码，默认 `zh-CN`。语言代码对应 MaxMind 数据库内置的名称表（如 `en`、`zh-CN`）；若某字段没有对应语言的名称，该字段为空。

### (*Client) Query

```go
func (g *Client) Query(rawIP string) (geoip.Result, error)
```

按 IP 查询归属地，实现 `geoip.GeoIP` 接口。

- `rawIP` 不是合法 IP 时返回 `invalid ip address` 错误；
- 私有/回环地址不查库，直接返回占位结果（见 Notes）；
- `Result.ISP` 恒为空，`Result.IP` 不回填。

### IsPrivateIP

```go
func IsPrivateIP(ip net.IP) bool
```

判断 IP 是否为内网地址，支持 IPv4 与 IPv6 私有网段（`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`、`127.0.0.0/8`、`169.254.0.0/16`、`fc00::/7`、`fe80::/10`、`::1/128`、`fec0::/10`）。

## Notes

- **私有地址占位结果**：内网/回环 IP（IPv4 与 IPv6）的查询不访问数据库，`Country`、`Province`、`City` 三个字段均返回占位字符串 `"局域网"`，`err` 为 `nil`。
- **查询失败会终止进程**：当底层数据库查询出错时，本包调用 kratos log 的 `log.Fatal(err)` **直接退出进程**。请确保传入合法且非私有的 IP。
- **仅内置 City 库**：不支持 GeoLite2-Country / GeoLite2-ASN，也不含 ISP 信息。
- **数据时效**：内嵌数据随本库发布更新。GeoLite2 上游每周更新两次，需要最新数据时可参考下方数据源自行获取。

## Acknowledgements

- [MaxMind GeoLite2 免费库](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- 数据下载/更新参考：
  1. <https://github.com/P3TERX/GeoLite.mmdb/releases>
  2. <https://dev.maxmind.com/geoip/geolite2-free-geolocation-data>
