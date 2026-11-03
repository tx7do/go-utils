# go-utils

A modular collection of Go utility libraries covering ID generation, cryptography, notification delivery, IP geolocation, and other common backend needs.

English | [简体中文](README.zh-CN.md)

## Features

- **Multi-module design**: 30+ independent Go modules (each with its own `go.mod`, independently versioned) — import only the single module you need instead of the whole repository.
- **Lightweight dependencies**: several modules stick to the pure standard library (`sqlutil`, `netutil`); every `notify` channel is implemented over bare HTTP with zero vendor SDKs.
- **KRatos ecosystem friendly**: `authorizer` loads [kratos-authz](https://github.com/tx7do/kratos-authz) engines (noop / casbin / opa), `jwtutil` reuses the KRatos codec layer, and `converter` parses KRatos operation IDs directly.
- **Protobuf-native**: `fieldperm` / `fieldmaskutil` trim protobuf messages field-by-field via protoreflect; `trans` / `timeutil` bridge protobuf well-known types.
- **Batteries included**: every module focuses on a single responsibility with a small API surface; most come with unit tests and usage examples.

## Module Index

Import paths are shown in full. Rows marked with 📦 are subpackages of their parent module (no standalone `go.mod`).

### ID & Generation

| Module | Import Path | Description | Docs |
|---|---|---|---|
| id | `github.com/tx7do/go-utils/id` | ID generators: UUID v4/v7, Snowflake, Sonyflake, order IDs, and host machine IDs. | [docs](./id/README.md) |
| 📦 id/machineid | (subpackage of `id`) | Read the unique machine ID of most host OSes without admin privileges. | [docs](./id/machineid/README.md) |
| name_generator | `github.com/tx7do/go-utils/name_generator` | Name generator producing Chinese, English, and Japanese person names, plus game nicknames and item names. | [docs](./name_generator/README.md) |
| code_generator | `github.com/tx7do/go-utils/code_generator` | Template-driven code generator with embedded and file-system template engines. | — |
| rand | `github.com/tx7do/go-utils/rand` | Random utilities: global functions plus configurable `Randomizer` instances, with `Seeder` and a SHA-256 random source. | [docs](./rand/README.md) |

### Security & Crypto

| Module | Import Path | Description | Docs |
|---|---|---|---|
| crypto | `github.com/tx7do/go-utils/crypto` | Encryption and signing: unified `Cipher`/`Signer`/`Encryptor` interfaces over AES, AES-GCM, RSA, ECDSA/ECDH, HMAC, and Chinese-national-standard SM algorithms. | [docs](./crypto/README.md) |
| password | `github.com/tx7do/go-utils/password` | Password hashing and verification: Bcrypt, Argon2, PBKDF2, SHA family, plus RSA/HMAC/ECDSA helpers and password policies. | [docs](./password/README.md) |
| jwtutil | `github.com/tx7do/go-utils/jwtutil` | JWT toolkit (built on golang-jwt/v5): parse, issue, refresh, verify, short-lived tokens, and extraction from HTTP requests. | — |
| captcha | `github.com/tx7do/go-utils/captcha` | Highly customizable CAPTCHAs: digits, strings, arithmetic, Chinese characters, slider puzzles, click-text, and image rotation. | [docs](./captcha/README.md) |
| authorizer | `github.com/tx7do/go-utils/authorizer` | Authorization engine manager: loads kratos-authz engines (noop / casbin / opa) by config and injects role→(path, method) policy providers. | [docs](./authorizer/README.md) |
| converter | `github.com/tx7do/go-utils/converter` | Permission-code converter mapping KRatos operation IDs or REST method+path to `resource:action` permission codes. | [docs](./converter/README.md) |
| fieldperm | `github.com/tx7do/go-utils/fieldperm` | Field-level permission primitives: read/write trimming of protobuf messages from token claims, via protoreflect. | [docs](./fieldperm/README.md) |

### Notification & Messaging

| Module | Import Path | Description | Docs |
|---|---|---|---|
| notify | `github.com/tx7do/go-utils/notify` | Unified multi-channel notification delivery: 40+ channels (Telegram, Slack, Discord, DingTalk, Feishu, WeCom, email, SMS, push, and more), all bare-HTTP with zero vendor SDKs. | [docs](./notify/README.md) |
| mailer | `github.com/tx7do/go-utils/mailer` | Email sending over `net/smtp`: STARTTLS and implicit SSL/TLS, with `ctx`-aware connection timeouts. | [docs](./mailer/README.md) |
| eventbus | `github.com/tx7do/go-utils/eventbus` | In-process publish/subscribe event bus: sync/async/once subscriptions and composable handlers. | [docs](./eventbus/README.md) |
| eventloop | `github.com/tx7do/go-utils/eventloop` | Single-goroutine sequential event loop: priority scheduling, frame-driven mode, and optional metrics. | [docs](./eventloop/README.md) |

### Network & Geo

| Module | Import Path | Description | Docs |
|---|---|---|---|
| netutil | `github.com/tx7do/go-utils/netutil` | Network security utilities: SSRF protection (blocked ranges, URL validation, resolution checks) and real client IP extraction. | [docs](./netutil/README.md) |
| geoip | `github.com/tx7do/go-utils/geoip` | IP geolocation backed by MaxMind GeoLite2, QQWry (纯真), and ip2region databases. | [docs](./geoip/README.md) |
| 📦 geoip/geolite | (subpackage of `geoip`) | Queries against MaxMind GeoLite2 mmdb databases. | — |
| 📦 geoip/qqwry | (subpackage of `geoip`) | Queries against the QQWry IP database. | — |
| 📦 geoip/ip2region | (subpackage of `geoip`) | Queries against the offline ip2region xdb database (v4/v6, with cache policies and a searcher pool). | — |
| auditutil | `github.com/tx7do/go-utils/auditutil` | HTTP audit-log collection and tamper-proofing helpers: normalized client IPs, request metadata, signature helpers. | [docs](./auditutil/README.md) |

### Data & Query

| Module | Import Path | Description | Docs |
|---|---|---|---|
| aggregator | `github.com/tx7do/go-utils/aggregator` | Data aggregation toolkit: parallel fetch and strongly-typed generic population, solving the N+1 query problem and DTO association backfill in microservices. | [docs](./aggregator/README.md) |
| query_parser | `github.com/tx7do/go-utils/query_parser` | Query parser: turns API sort and filter expressions into SQL conditions. | [docs](./query_parser/README.md) |
| ddl_parser | `github.com/tx7do/go-utils/ddl_parser` | Small SQL DDL parser focused on extracting CREATE TABLE definitions. | [docs](./ddl_parser/README.md) |
| sqlutil | `github.com/tx7do/go-utils/sqlutil` | PostgreSQL lexical-level SQL auditing: literal masking and table extraction, pure standard library. | [docs](./sqlutil/README.md) |
| mapper | `github.com/tx7do/go-utils/mapper` | Type mapper: struct field mapping with enum conversion, for moving data between entities and DTOs. | [docs](./mapper/README.md) |
| copierutil | `github.com/tx7do/go-utils/copierutil` | Type converters for [jinzhu/copier](https://github.com/jinzhu/copier): time↔string, protobuf Timestamp/Duration, and other common conversions. | — |

### Text & Conversion

| Module | Import Path | Description | Docs |
|---|---|---|---|
| doctext | `github.com/tx7do/go-utils/doctext` | Plain-text extraction from common document formats (txt/md/csv/json/xml/yaml/html, docx, pdf) for RAG ingestion and full-text search. | [docs](./doctext/README.md) |
| slug | `github.com/tx7do/go-utils/slug` | URL slug generation (built on gosimple/slug) with lowercase, case-sensitive, and English variants. | — |

### Translation

| Module | Import Path | Description | Docs |
|---|---|---|---|
| translator | `github.com/tx7do/go-utils/translator` | Unified `Translator` interface and target-language definitions; vendor implementations live in submodules. | [docs](./translator/README.md) |
| translator/alibaba | `github.com/tx7do/go-utils/translator/alibaba` | Alibaba Cloud machine translation client (200+ language pairs). | [docs](./translator/alibaba/README.md) |
| translator/baidu | `github.com/tx7do/go-utils/translator/baidu` | Baidu Translate API client. | [docs](./translator/baidu/README.md) |
| translator/google | `github.com/tx7do/go-utils/translator/google` | Google Translate API client (three API versions). | [docs](./translator/google/README.md) |
| translator/volc | `github.com/tx7do/go-utils/translator/volc` | Volcengine (ByteDance) translation client. | [docs](./translator/volc/README.md) |

### Storage & Infrastructure

| Module | Import Path | Description | Docs |
|---|---|---|---|
| ossutil | `github.com/tx7do/go-utils/ossutil` | SDK-agnostic object-storage helpers: MIME/extension-to-bucket mapping and magic-number file type detection. | [docs](./ossutil/README.md) |
| distlock | `github.com/tx7do/go-utils/distlock` | Unified distributed-lock abstraction over Redis and Etcd backends. | [docs](./distlock/README.md) |
| bank_card | `github.com/tx7do/go-utils/bank_card` | Bank card BIN lookup and validation: issuer identification and Luhn checksum checking. | [docs](./bank_card/README.md) |

## Root Module Packages

The following packages ship with the root module `github.com/tx7do/go-utils` (no standalone `go.mod`):

| Package | Description | Docs |
|---|---|---|
| trans | Pointer/type conversion helpers (`String`, `Int`, etc. return pointers). | — |
| timeutil | Time layouts, formatting, diff, ranges, and protobuf Timestamp/Duration conversion. | — |
| fieldmaskutil | Protobuf FieldMask utilities: keep/clear fields by mask paths, with nested masking. | [docs](./fieldmaskutil/README.md) |
| stringcase | Naming-convention conversion: camelCase, PascalCase, snake_case, kebab-case. | [docs](./stringcase/README.md) |
| stringutil | String utilities: type conversion, random and crypto-secure random strings, JSON field replacement. | [docs](./stringutil/README.md) |
| sliceutil | Generic slice utilities (Filter, Map, etc.), inspired by JavaScript/Python. | — |
| maputils | Generic map utilities (Keys, Filter, etc.). | — |
| structutil | Iterate struct fields via reflection (name, value, tag). | — |
| byteutil | Integer/byte-slice conversions and binary read/write helpers. | — |
| math | Generic math helpers (sign function, Gaussian, etc.). | — |
| pagination | Pagination helpers (page offset calculation). | — |
| dateutil | Date helpers: align to day start/end boundaries and similar operations. | — |
| cryptocurrency | Cryptocurrency address format validation (BTC, BTG, DASH, and more). | — |
| tls | TLS certificate loading helpers for mutual-TLS setups. | — |
| ioutil | Simple file and path helpers. | — |

## License

[MIT](./LICENSE)
