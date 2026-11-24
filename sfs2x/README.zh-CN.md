# sfs2x — SFS2X 二进制协议编解码

`sfs2x` 实现 SFS2X 二进制线协议（SmartFoxServer 2X 兼容服务所用的对象序列化与消息帧封装）：类型化键值对象、嵌套对象与数组、长度前缀帧。纯标准库实现——零外部依赖。

[English](README.md) | 简体中文

## 线格式

```
frame  = [header 1B][len][payload]
         header: bit5 (0x20) = 载荷 zlib 压缩
                 bit3 (0x08) = 长度为 uint32，否则 uint16
object = [0x12][u16 count][entries...]
entry  = [u16 keyLen][key utf8][1B type][value]
```

嵌套 `SFS_OBJECT` / `SFS_ARRAY` 值不带独立 marker——类型字节即 marker，解码器在其上回退一字节。

## 类型码

| 码 | 类型 | 码 | 类型 |
|---|---|---|---|
| 0 | NULL | 9 | BOOL_ARRAY |
| 1 | BOOL | 10 | BYTE_ARRAY |
| 2 | BYTE | 11 | SHORT_ARRAY |
| 3 | SHORT | 12 | INT_ARRAY |
| 4 | INT | 13 | LONG_ARRAY |
| 5 | LONG | 14 | FLOAT_ARRAY |
| 6 | FLOAT | 15 | DOUBLE_ARRAY |
| 7 | DOUBLE | 16 | UTF_STRING_ARRAY |
| 8 | UTF_STRING | 17 | SFS_ARRAY |
| | | 18 | SFS_OBJECT |

所有整数均为大端序。

## 安装

```bash
go get github.com/tx7do/go-utils/sfs2x
```

```go
import "github.com/tx7do/go-utils/sfs2x"
```

## 快速开始

```go
o := sfs2x.New()
o.Put("action", sfs2x.TUTFString, "login")
o.Put("retries", sfs2x.TInt, int32(3))
o.Put("ratio", sfs2x.TDouble, 0.75)

inner := sfs2x.New()
inner.Put("token", sfs2x.TUTFString, "abc")
o.Put("params", sfs2x.TSFSObject, inner)

// Object → bytes（含 0x12 marker）
data := o.Encode()

// bytes → Object
decoded, err := sfs2x.Decode(data)

// 帧传输（签名 + uint16 长度 + 载荷）
frame := sfs2x.EncodeFrame(o)
msg, err := sfs2x.DecodeFrame(frame)
```

## API

| 函数 | 说明 |
|---|---|
| `New() *Object` | 创建空对象。 |
| `Put(key string, typ Type, val interface{})` | 写入类型化值。 |
| `Get(key string) interface{}` | 读取原始值（缺失返回 nil）。 |
| `GetType(key string) Type` | 读取条目类型码。 |
| `GetKeys() []string` | 按插入序返回 key。 |
| `GetByte/GetShort/GetInt/GetLong(key)` | 类型化读取（类型不匹配会 panic——不可信输入请先 `GetType`）。 |
| `GetUTFString(key) string` | UTF_STRING 读取；缺失或类型不符返回空串。 |
| `GetSFSObject(key) *Object` | 嵌套对象读取；缺失或类型不符返回 nil。 |
| `GetSFSArray(key) []interface{}` | 嵌套数组读取；缺失或类型不符返回 nil。 |
| `Encode() []byte` | 序列化（含 0x12 marker）。 |
| `Decode(data []byte) (*Object, error)` | 从二进制解析（要求 marker）。 |
| `EncodeFrame(o *Object) []byte` | 封装为帧（签名 + uint16 长度）。 |
| `DecodeFrame(data []byte) (*Object, error)` | 解析帧；兼容 uint16 与 uint32 两种长度形态。 |

## 注意事项

- **确定性编码**：key 在序列化前排序，同内容无论插入序如何都产出逐字节一致的输出（`GetKeys` 仍报告插入序）。
- **帧形态**：`DecodeFrame` 两种长度形态（header bit `0x08`）都接受。`EncodeFrame` 发短形态，载荷超过 65535 字节时自动切换 BIG_SIZE（uint32），与官方客户端一致。
- **解码边界安全**：截断或畸形输入返回错误，绝不 panic。`DecodeFrame` 对压缩（0x20）、加密（0x40）、非二进制帧显式报错拒绝，不做误导性硬解。
- **压缩**：header bit `0x20` 的帧携带 zlib 压缩载荷；本包不做解压——对端若发该形态请在传输层处理。
- **UTF_STRING 上限**：线格式以 uint16 携带字符串长度，超过 65535 字节的值无法编码（长度字段静默截断——与官方客户端同款上限）；更大载荷请用 BYTE_ARRAY（uint32 长度）。
- **类型安全**：类型化读取器（`GetInt` 等）使用无检查类型断言，不匹配即 panic。不可信输入请先查 `GetType` 或用 `Get` 安全断言。

## 协议忠实度

对照官方 SFS2X Java 客户端 API 行为（经忠实移植源码）与生产对端联调捕获核验：

| 行为 | 本包 | 官方客户端 API |
|---|---|---|
| 类型码 0–18 | ✓ 一致 | — |
| 类型化数组（BOOL/SHORT/INT/LONG/FLOAT/DOUBLE/UTF_STRING） | ✓ 编码+解码 | 一致 |
| BYTE_ARRAY 长度字段 | **uint32** | uint32（`writeInt`）——部分第三方移植误用 uint16 |
| 其他数组/字符串长度字段 | uint16 | uint16 |
| 嵌套对象/数组 marker | 类型字节兼任 marker（回退一字节） | 一致 |
| 帧头标志 | `0x80` BINARY、`0x08` BIG_SIZE（解码侧） | 另有 `0x40` ENCRYPTED、`0x10` BLUEBOX |
| 压缩（header `0x20`） | 未实现；`DecodeFrame` 对该标志显式拒绝 | zlib，阈值 1024 |
| 加密（header `0x40`） | 未实现；`DecodeFrame` 对该标志显式拒绝 | AES-128-CBC，随机 IV + PKCS#7 |
| CLASS (19) / TEXT (20) 类型 | 未实现 | CLASS 走服务端注册表序列化 |
| **对象键序（编码侧）** | **排序**（确定性输出） | **插入序**（OrderedHashMap） |

唯一的有意偏离是键序：本包排序使同内容产出逐字节一致编码（测试可复现、载荷可哈希、日志稳定）。任何合规解析器都把 SFSObject 当键值映射处理，解码侧行为不受影响。若需要与官方客户端编码器字节级对齐，请改为按对方的插入序排布——或把 `entriesBin` 里的 `sort.Strings` 换成插入序输出。
