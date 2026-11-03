# geoip

[English](README.md) | 简体中文

IP 归属地（地理位置）查询模块。本模块定义了统一的 `GeoIP` 接口与 `Result` 结果结构，并提供三个数据文件内嵌、开箱即用的子包实现：[geolite](./geolite)（MaxMind GeoLite2）、[qqwry](./qqwry)（纯真库）与 [ip2region](./ip2region)（IPv4/IPv6 双栈）。

## Features

- 统一的 `GeoIP` 接口与 `Result` 结果结构，业务代码可无感切换数据源
- 三个子包实现，覆盖国内外 IP 查询场景
- 数据文件全部通过 Go `embed` 内嵌，编译出单个二进制即可运行，无需外部部署数据文件
- IPv4 / IPv6 支持因数据源而异（见下表）

## Installation

```bash
go get github.com/tx7do/go-utils/geoip@latest
```

本目录是独立 Go 模块（module 路径为 `github.com/tx7do/go-utils/geoip`）。`geolite`、`qqwry`、`ip2region` 均为该模块下的子包（没有各自的 `go.mod`），安装本模块即可使用全部子包：

```go
import (
	"github.com/tx7do/go-utils/geoip"
	"github.com/tx7do/go-utils/geoip/geolite"
	"github.com/tx7do/go-utils/geoip/ip2region"
	"github.com/tx7do/go-utils/geoip/qqwry"
)
```

## 子包对比

| 子包 | 数据源 | 内置数据 | IPv4 | IPv6 | ISP 字段 | 特点 |
| -------- | -------- | -------- | :--: | :--: | :--: | -------- |
| [geolite](./geolite) | MaxMind GeoLite2 | GeoLite2-City.mmdb（约 60 MB） | ✓ | ✓ | ✗ | 仅内置 City 库；输出语言可配置（默认 `zh-CN`） |
| [qqwry](./qqwry) | 纯真社区版 IP 库 | qqwry.dat（约 10 MB） | ✓ | ✗ | ✓ | 结果为中文；`Country` 恒为 `"中国"` |
| [ip2region](./ip2region) | ip2region xdb | ip2region_v4.xdb + ip2region_v6.xdb（合计约 50 MB） | ✓ | ✓ | ✓ | 双栈支持；v4/v6 可独立启用或禁用 |

注意：三个实现的 `Result.IP` 回填行为不一致——qqwry 会把查询 IP 写入 `Result.IP`，geolite 与 ip2region 不会。

## Quick Start

以 qqwry 为例（三个子包的 `Client` 均实现 `geoip.GeoIP` 接口，可直接互相替换），其余子包用法详见各自 README：

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
	// 输出: {IP:47.108.149.89 Country:中国 Province:四川省 City:成都市 ISP:阿里云}
}
```

## API

### Result 归属地结果

```go
type Result struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 城市
	ISP      string `json:"isp"`      // 服务提供商
}
```

### GeoIP 接口

```go
type GeoIP interface {
	Query(queryIp string) (res Result, err error)
}
```

### 子包文档

- [geoip/geolite](./geolite/README.md)（[中文](./geolite/README.zh-CN.md)）— MaxMind GeoLite2-City
- [geoip/qqwry](./qqwry/README.md)（[中文](./qqwry/README.zh-CN.md)）— 纯真 IP 库（IPv4）
- [geoip/ip2region](./ip2region/README.md)（[中文](./ip2region/README.zh-CN.md)）— ip2region xdb（IPv4/IPv6）

## Notes

- 内嵌数据会显著增大二进制体积（geolite 约 60 MB、ip2region 约 50 MB、qqwry 约 10 MB），且在客户端初始化时载入内存，请按需选择子包。
- 内网/回环地址的处理：geolite 将 `Country/Province/City` 全部置为 `"局域网"`；qqwry 将 `Province/City` 置为 `"局域网"`。详见各自 README。
- 内嵌数据的时效性随本库发布更新，如需最新数据请关注各子包 README 中的数据源链接。

## Acknowledgements

- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [纯真社区版IP库](https://update.cz88.net/geo-public)
- [ip2region](https://github.com/lionsoul2014/ip2region)
