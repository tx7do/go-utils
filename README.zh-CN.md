# go-utils

一套模块化的 Go 工具库合集，覆盖 ID 生成、加解密、通知投递、IP 地理位置等常用后端场景。

[English](README.md) | 简体中文

## 特性

- **多模块设计**：仓库内含 30+ 个独立 Go module（各自拥有 `go.mod`、独立版本化），按需引入单个模块即可，不必拉起整个仓库。
- **轻依赖**：多个模块坚持纯标准库实现（如 `sqlutil`、`netutil`）；`notify` 全部渠道为裸 HTTP 实现，零厂商 SDK。
- **KRatos 生态友好**：`authorizer` 对接 [kratos-authz](https://github.com/tx7do/kratos-authz) 引擎（noop / casbin / opa），`jwtutil` 复用 KRatos 编码层，`converter` 直接解析 KRatos operationID。
- **Protobuf 原生**：`fieldperm` / `fieldmaskutil` 基于 protoreflect 做字段级消息裁剪；`trans` / `timeutil` 打通 protobuf well-known 类型。
- **开箱即用**：每个模块聚焦单一职责，API 面小，多数模块附带单元测试与使用示例。

## 模块索引

独立模块的 import 路径即表中所示完整子路径。带 📦 标注的行属于其父模块的子包（无独立 `go.mod`）。

### ID 与生成

| Module | Import Path | Description | Docs |
|---|---|---|---|
| id | `github.com/tx7do/go-utils/id` | ID 生成器：UUID v4/v7、Snowflake、Sonyflake、订单号、机器标识。 | [docs](./id/README.md) |
| 📦 id/machineid | （`id` 模块子包） | 读取各操作系统主机唯一机器 ID（无需管理员权限）。 | [docs](./id/machineid/README.md) |
| name_generator | `github.com/tx7do/go-utils/name_generator` | 名字生成器：中文、英文、日文人名，以及游戏昵称、道具名。 | [docs](./name_generator/README.md) |
| code_generator | `github.com/tx7do/go-utils/code_generator` | 模板驱动代码生成器，支持内嵌与文件系统两种模板引擎。 | — |
| rand | `github.com/tx7do/go-utils/rand` | 随机数工具：全局函数 + 可配置 Randomizer 实例，含 Seeder 种子生成与 SHA-256 随机源。 | [docs](./rand/README.md) |

### 安全与加密

| Module | Import Path | Description | Docs |
|---|---|---|---|
| crypto | `github.com/tx7do/go-utils/crypto` | 加解密与签名：统一 `Cipher`/`Signer`/`Encryptor` 接口，覆盖 AES、AES-GCM、RSA、ECDSA/ECDH、HMAC、国密 SM 系列。 | [docs](./crypto/README.md) |
| password | `github.com/tx7do/go-utils/password` | 密码哈希与校验：Bcrypt、Argon2、PBKDF2、SHA 系列，附 RSA/HMAC/ECDSA 辅助与密码强度策略。 | [docs](./password/README.md) |
| jwtutil | `github.com/tx7do/go-utils/jwtutil` | JWT 工具（基于 golang-jwt/v5）：解析、签发、刷新、校验、短时效 token 与 HTTP 请求提取。 | — |
| captcha | `github.com/tx7do/go-utils/captcha` | 高度可定制的验证码：数字、字符串、算术、中文、滑动拼图、点击文字、旋转图片。 | [docs](./captcha/README.md) |
| authorizer | `github.com/tx7do/go-utils/authorizer` | 授权引擎管理器：按配置装载 kratos-authz 引擎（noop / casbin / opa）并注入角色→(path, method) 策略。 | [docs](./authorizer/README.md) |
| converter | `github.com/tx7do/go-utils/converter` | 权限码转换器：把 KRatos operationID 或 REST method+path 映射为 `resource:action` 权限码。 | [docs](./converter/README.md) |
| fieldperm | `github.com/tx7do/go-utils/fieldperm` | 字段级权限裁剪原语：基于 protoreflect 按令牌 claim 对 protobuf 消息做读写两向裁剪。 | [docs](./fieldperm/README.md) |

### 通知与消息

| Module | Import Path | Description | Docs |
|---|---|---|---|
| notify | `github.com/tx7do/go-utils/notify` | 多渠道通知投递统一抽象：40+ 渠道（Telegram、Slack、Discord、钉钉、飞书、企业微信、邮件、短信、推送等），裸 HTTP 实现、零厂商 SDK。 | [docs](./notify/README.md) |
| mailer | `github.com/tx7do/go-utils/mailer` | 基于 `net/smtp` 的邮件发送：支持 STARTTLS 与隐式 SSL/TLS，建连贯穿 `ctx` 超时控制。 | [docs](./mailer/README.md) |
| eventbus | `github.com/tx7do/go-utils/eventbus` | 进程内发布/订阅事件总线：同步/异步/一次性订阅、handler 组合子。 | [docs](./eventbus/README.md) |
| eventloop | `github.com/tx7do/go-utils/eventloop` | 单 goroutine 顺序事件循环：按优先级调度，支持帧驱动模式与 Metrics 统计。 | [docs](./eventloop/README.md) |

### 网络与地理

| Module | Import Path | Description | Docs |
|---|---|---|---|
| netutil | `github.com/tx7do/go-utils/netutil` | 网络安全工具：SSRF 防护（禁投网段、URL 校验、解析校验）与客户端真实 IP 提取。 | [docs](./netutil/README.md) |
| geoip | `github.com/tx7do/go-utils/geoip` | IP 查询地理位置，支持 MaxMind GeoLite2、纯真（QQWry）、ip2region 三种数据库。 | [docs](./geoip/README.md) |
| 📦 geoip/geolite | （`geoip` 模块子包） | MaxMind GeoLite2 mmdb 数据库查询。 | — |
| 📦 geoip/qqwry | （`geoip` 模块子包） | 纯真 IP 库查询。 | — |
| 📦 geoip/ip2region | （`geoip` 模块子包） | ip2region xdb 离线 IP 库查询（v4/v6，支持缓存策略与搜索器池）。 | — |
| auditutil | `github.com/tx7do/go-utils/auditutil` | HTTP 审计日志信息采集与防篡改辅助：客户端 IP 归一化、请求元数据、签名辅助。 | [docs](./auditutil/README.md) |

### 数据与查询

| Module | Import Path | Description | Docs |
|---|---|---|---|
| aggregator | `github.com/tx7do/go-utils/aggregator` | 数据聚合工具包：并行 Fetch 与强类型泛型 Populate，解决微服务 N+1 查询与 DTO 关联回填。 | [docs](./aggregator/README.md) |
| query_parser | `github.com/tx7do/go-utils/query_parser` | 查询解析器：把 API 排序与过滤表达式解析为 SQL 条件。 | [docs](./query_parser/README.md) |
| ddl_parser | `github.com/tx7do/go-utils/ddl_parser` | 轻量 SQL DDL 解析器，专注提取 CREATE TABLE 定义。 | [docs](./ddl_parser/README.md) |
| sqlutil | `github.com/tx7do/go-utils/sqlutil` | PostgreSQL 词法级 SQL 审计：字面量脱敏与表提取，纯标准库。 | [docs](./sqlutil/README.md) |
| mapper | `github.com/tx7do/go-utils/mapper` | 类型映射器：结构体字段映射与枚举转换，用于实体与 DTO 间的数据转换。 | [docs](./mapper/README.md) |
| copierutil | `github.com/tx7do/go-utils/copierutil` | [jinzhu/copier](https://github.com/jinzhu/copier) 类型转换器集合：时间↔字符串、protobuf Timestamp/Duration 等常用转换。 | — |

### 文本与转换

| Module | Import Path | Description | Docs |
|---|---|---|---|
| doctext | `github.com/tx7do/go-utils/doctext` | 从常见文档格式抽取纯文本（txt/md/csv/json/xml/yaml/html、docx、pdf），供 RAG 入库与全文搜索使用。 | [docs](./doctext/README.md) |
| slug | `github.com/tx7do/go-utils/slug` | URL 短链/别名生成（基于 gosimple/slug），提供小写、大小写敏感、英文等变体。 | — |

### 翻译

| Module | Import Path | Description | Docs |
|---|---|---|---|
| translator | `github.com/tx7do/go-utils/translator` | 翻译器统一接口 `Translator` 与目标语言定义；各厂商实现见子模块。 | [docs](./translator/README.md) |
| translator/alibaba | `github.com/tx7do/go-utils/translator/alibaba` | 阿里云机器翻译客户端（200+ 语种组合）。 | [docs](./translator/alibaba/README.md) |
| translator/baidu | `github.com/tx7do/go-utils/translator/baidu` | 百度翻译 API 客户端。 | [docs](./translator/baidu/README.md) |
| translator/google | `github.com/tx7do/go-utils/translator/google` | Google 翻译 API 客户端（三个 API 版本）。 | [docs](./translator/google/README.md) |
| translator/volc | `github.com/tx7do/go-utils/translator/volc` | 火山引擎（字节跳动）翻译客户端。 | [docs](./translator/volc/README.md) |

### 存储与基础设施

| Module | Import Path | Description | Docs |
|---|---|---|---|
| ossutil | `github.com/tx7do/go-utils/ossutil` | 对象存储通用工具（与 SDK 解耦）：MIME/扩展名到分类桶映射、魔术头文件类型嗅探。 | [docs](./ossutil/README.md) |
| distlock | `github.com/tx7do/go-utils/distlock` | 统一分布式锁抽象，屏蔽 Redis / Etcd 底层实现差异。 | [docs](./distlock/README.md) |
| bank_card | `github.com/tx7do/go-utils/bank_card` | 银行卡 BIN 查询与校验：卡号归属行识别、Luhn 校验。 | [docs](./bank_card/README.md) |

## 根模块工具包

以下包随根模块 `github.com/tx7do/go-utils` 分发（无独立 `go.mod`）：

| Package | Description | Docs |
|---|---|---|
| trans | 指针/类型转换辅助（`String`、`Int` 等返回指针的便捷函数）。 | — |
| timeutil | 时间布局、格式化、时差、区间与 protobuf Timestamp/Duration 互转。 | — |
| fieldmaskutil | Protobuf FieldMask 工具：按 paths 保留/清除字段，支持嵌套掩码。 | [docs](./fieldmaskutil/README.md) |
| stringcase | 命名法转换：camelCase、PascalCase、snake_case、kebab-case。 | [docs](./stringcase/README.md) |
| stringutil | 字符串工具：类型互转、随机与加密安全随机字符串、JSON 字段替换。 | [docs](./stringutil/README.md) |
| sliceutil | 泛型切片工具（Filter、Map 等，借鉴 JavaScript/Python）。 | — |
| maputils | 泛型映射工具（Keys、Filter 等）。 | — |
| structutil | 反射遍历结构体字段（名称、值、tag）。 | — |
| byteutil | 整数与字节切片互转、二进制读写辅助。 | — |
| math | 泛型数学辅助（符号函数、高斯函数等）。 | — |
| pagination | 分页辅助（页码偏移量计算等）。 | — |
| dateutil | 日期辅助：对齐到当天起止边界等。 | — |
| cryptocurrency | 加密货币地址格式校验（BTC、BTG、DASH 等）。 | — |
| tls | TLS 证书加载辅助，用于双向认证配置。 | — |
| ioutil | 简单的文件与路径辅助函数。 | — |

## 许可证

[MIT](./LICENSE)
