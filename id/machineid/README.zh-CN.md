# machineid

[English](README.md) | 简体中文

`machineid` 是 [github.com/denisbrodbeck/machineid](https://github.com/denisbrodbeck/machineid) 的 fork，用于读取大多数操作系统的唯一机器 ID（无需管理员权限）。import 路径为 `github.com/tx7do/go-utils/id/machineid`，随 `id` 模块一同分发：

```bash
go get github.com/tx7do/go-utils/id
```

> 本仓库中没有上游的 `cmd/machineid` 命令行工具，本包只提供 Go API。

## 功能特性

- 跨平台（Windows、Linux、macOS、BSD/Solaris、AIX、z/OS）
- 无需管理员权限
- 不依赖硬件标识（不使用 MAC、BIOS 或 CPU 序列号——这些在虚拟机环境中不可靠）
- ID 在操作系统安装生命周期内保持稳定
- **fork 增强**：Linux 下扩展了回退链，覆盖云主机与容器环境（见下文）
- 提供加密安全的 `ProtectedID`（HMAC-SHA256），避免直接暴露原始机器 ID

## 用法

获取原始机器 ID：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/id/machineid"
)

func main() {
	id, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
```

更推荐的做法是使用哈希保护后的机器 ID：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/id/machineid"
)

func main() {
	id, err := machineid.ProtectedID("myAppName")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
```

### Function: ID() (string, error)

以 `string` 返回原始机器 ID。请将其视为"机密信息"，优先使用 `ProtectedID()`。

### Function: ProtectedID(appID string) (string, error)

返回机器 ID 的哈希版本：以机器 ID 为密钥，对应用专属的 appID 计算 HMAC-SHA256，输出十六进制字符串。同一台机器上不同应用会得到不同且不可逆推的 ID。

## 各平台取值来源

本包返回操作系统自身的机器 UUID/GUID，通常在系统安装时生成，此后保持不变：

* **Linux**：按顺序尝试 `/var/lib/dbus/machine-id` → `/etc/machine-id` → `/sys/class/dmi/id/product_uuid`（云主机/裸金属）→ `/proc/self/cgroup` → `/proc/self/mountinfo`（Docker 容器 ID）
* **Windows**：注册表 `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography` 的 `MachineGuid` → BIOS UUID（`wmic`/PowerShell）→ 主网卡 MAC 地址
* **macOS**：`IOPlatformUUID`
* **BSD**：OpenBSD 优先 `sysctl -n hw.uuid`；其余使用 `/etc/hostid`，回退 `kenv -q smbios.system.uuid`
* **AIX**：`lsattr` 的 `os_uuid` 属性
* **z/OS**：通过语言环境（LE）运行时服务读取系统标识

### fork 增强：Linux 容器环境回退链

上游实现只读取 dbus machine-id 两个路径，在容器中往往为空或与宿主机相同。本 fork 在两个路径都读取失败时继续回退：

1. `/sys/class/dmi/id/product_uuid`——云主机/虚拟机/裸金属的 DMI 产品 UUID
2. `/proc/self/cgroup`——cgroup 路径片段
3. `/proc/self/mountinfo`——Docker 容器 ID（匹配 `/docker/containers/<id>/hostname`）

因此本 fork 在容器与云主机环境下的成功率显著高于上游。

## 唯一性可靠性

`machine-id` 与 `MachineGuid` 可以被 root/管理员修改（但可能带来系统服务异常等代价）。克隆/镜像/恢复操作系统安装时，大多数 ID 不会被操作系统重新生成——克隆 Windows 安装（不使用官方 sysprep 工具）是一个已知案例。

**Linux** 用户可用 `dbus-uuidgen` 生成新 ID 并写入 `/var/lib/dbus/machine-id` 和 `/etc/machine-id`。
**Windows** 用户可使用 `sysprep` 工具链制作镜像，此类镜像在每次部署时会产生新的唯一机器 ID。

## 安全考量

机器 ID 唯一标识主机，应视为"机密"，不得暴露在不可信环境中。如果需要一个稳定的应用唯一标识，不要直接使用机器 ID：

> 可靠的做法是以固定、应用专属的密钥，对机器 ID 做加密安全的哈希。

这样得到的 ID 既保持唯一性和确定性，又无法反推原始机器 ID。直接使用便捷 API 即可：

```go
hashedID, err := machineid.ProtectedID("myAppName")
```

## 手动查询各平台机器 ID

* **AIX**：`lsattr -l sys0 -a os_uuid -E | cut -f 2 -d ' '`
* **BSD**：`cat /etc/hostid`（或可能为空的 `kenv -q smbios.system.uuid`）
* **Linux**：`cat /var/lib/dbus/machine-id`（或部分系统独有的 `cat /etc/machine-id`）
* **macOS**：`ioreg -rd1 -c IOPlatformExpertDevice | grep IOPlatformUUID`
* **Windows**：`reg query HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography /v MachineGuid`

## 致谢

本包基于 [Denis Brodbeck](https://github.com/denisbrodbeck) 的 [denisbrodbeck/machineid](https://github.com/denisbrodbeck/machineid)（MIT 许可证），保留了上游的 MIT 署名（见 [LICENSE.md](LICENSE.md)，Copyright (c) 2017 Denis Brodbeck）。原 Go gopher 形象由 Denis Brodbeck 使用 [gopherize.me](https://gopherize.me/) 制作，基于 Renee French 的原画。

## 许可证

MIT
