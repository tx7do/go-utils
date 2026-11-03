# qqwry

[English](README.md) | 简体中文

基于纯真（CZ88）IP 库的 IPv4 归属地查询包，是 `github.com/tx7do/go-utils/geoip` 模块的子包（import 路径 `github.com/tx7do/go-utils/geoip/qqwry`）。数据文件 `qqwry.dat` 已通过 Go `embed` 内嵌，**开箱即用，无需自行下载**。

本包实现父包的 `geoip.GeoIP` 接口，查询结果为 `geoip.Result`。

## Features

- 内置纯真 IP 库数据（`qqwry.dat`，约 10 MB，约 53 万条 IP 记录），开箱即用
- 结果自动做 GB18030 解码，直接得到中文
- 自动从地址串拆分省份与城市，并解析 ISP（运营商）
- 内置私有网段判断工具 `IsPrivateIP` 与地址拆分工具 `SpiltAddress`

> **注意**：本包仅支持 **IPv4**，且 `Country` 恒为 `"中国"`（见 Notes）。

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
	// 输出: {IP:47.108.149.89 Country:中国 Province:四川省 City:成都市 ISP:阿里云}
}
```

## API

### NewClient

```go
func NewClient() *Client
```

从内嵌的 `qqwry.dat` 创建客户端并初始化索引区间。没有 error 返回值；数据随二进制内嵌，`Client` 也不提供 `Close`，无需手动关闭。

### Client.IPNum

```go
IPNum int64
```

数据文件索引中的 IP 记录总条数（当前内置数据约 53 万条）。

### (*Client) Query

```go
func (c *Client) Query(queryIp string) (geoip.Result, error)
```

按 IPv4 地址查询归属地，实现 `geoip.GeoIP` 接口。

- 仅接受 IPv4；IPv6 或非法字符串返回 `ip is not ipv4` 错误；
- 私有/回环 IPv4（含 `127.0.0.1`）不查数据文件，返回 `Province = City = "局域网"`；
- 索引中未找到时返回 `ip not found` 错误；
- `Result.IP` 会回填查询的 IP；`Country` 恒为 `"中国"`。

### IsPrivateIP

```go
func IsPrivateIP(ipStr string) bool
```

判断 IP 字符串是否为内网地址。**注意与 [geolite](../geolite) 包的同名函数不同**：本函数接受 `string` 参数；包含 `:` 的字符串（IPv6）或解析失败的输入一律返回 `true`（视为内网）。IPv4 私有网段：`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`、`127.0.0.0/8`、`169.254.0.0/16`。

### SpiltAddress

```go
func SpiltAddress(addr string) []string
```

按行政区划后缀（省、市、自治区、自治州、盟、县、区、管委会、街道、镇、乡）切分地址字符串，例如 `"中国广东省深圳市南山区"` → `["中国广东省", "深圳市", "南山区"]`。`Query` 内部即用它拆分省份与城市。

## Notes

- **仅支持 IPv4**：传入 IPv6 地址会直接报错；需要 IPv6 查询请使用 [geolite](../geolite) 或 [ip2region](../ip2region)。
- **`Country` 恒为 `"中国"`**：`Query` 一开始就硬编码 `Country = "中国"`。查询海外 IP 时，其国家名（如 `"美国"`）会落在 `City` 字段，详细位置信息落在 `ISP` 字段，解读结果时请注意。
- **私有地址占位结果**：私有/回环 IPv4 不查数据文件，`Province`、`City` 返回 `"局域网"` 占位。由于私有判断先于记录查询，`127.0.0.1` 实际返回的也是 `"局域网"` 占位而非数据文件中的记录。
- **结果语言**：数据源为纯真中文 IP 库，结果固定为中文；`ISP` 中的占位运营商 `"CZ88.NET"` 会被自动置空。

## 数据更新

包内数据为构建时的快照。需要最新数据时，可从以下渠道获取新的 `qqwry.dat`（本包暂未提供外部数据加载接口；如需使用新数据，替换 `assets/qqwry.dat` 后重新编译即可）：

- [纯真社区版IP库](https://update.cz88.net/geo-public)
- [qqwry.dat - out0fmemory - GitHub](https://github.com/out0fmemory/qqwry.dat/blob/master/qqwry_lastest.dat)
- [qqwry - xiaoqidun - GitHub](https://github.com/xiaoqidun/qqwry)
- [qqwry - freshcn - GitHub](https://github.com/freshcn/qqwry)

## Acknowledgements

- [纯真（CZ88）](https://update.cz88.net/geo-public) — qqwry.dat 数据来源
