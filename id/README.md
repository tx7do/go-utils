# id

English | [简体中文](README.zh-CN.md)

`id` is a standalone Go module (`github.com/tx7do/go-utils/id`) providing multiple unique-ID generation capabilities: general-purpose primary keys, distributed time-ordered IDs, order numbers, and machine identifiers.

## Features

- UUID/GUID: `NewGUIDv4`, `NewGUIDv7` (with or without hyphens)
- Short IDs: `NewShortUUID`, `NewKSUID`, `NewXID`
- MongoDB ID: `NewMongoObjectID`
- Distributed IDs: `NewSnowflakeID` / `GenerateSnowflakeID`, `NewSonyflakeID` / `GenerateSonyflakeID`
- Snowflake nodes: `NewSnowflakeNode` returns a reusable node instance with `Generate` / `GenerateString`
- Order numbers: random, auto-incrementing, tenant-scoped, prefixed Snowflake/Sonyflake
- Machine ID: `UnifyMachineID`, `FormatMachineID` (cross-platform, based on the embedded `id/machineid` subpackage)

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`:

```bash
go get github.com/tx7do/go-utils/id
```

```go
import "github.com/tx7do/go-utils/id"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/id"
)

func main() {
	guid := id.NewGUIDv4(false)   // 32 chars, no hyphens
	guidV7 := id.NewGUIDv7(true)  // hyphenated UUIDv7 (falls back to v4 if the clock is unreliable)
	shortID := id.NewShortUUID()
	ksuid := id.NewKSUID()
	xid := id.NewXID()
	mongoID := id.NewMongoObjectID()

	// Snowflake (nodes are cached by workerId and reused within the process)
	sfID, err := id.NewSnowflakeID(1)
	if err != nil {
		log.Fatal(err)
	}

	// Sonyflake (process-wide singleton)
	sonyID, err := id.NewSonyflakeID()
	if err != nil {
		log.Fatal(err)
	}

	// Machine ID (32 lowercase hex chars)
	machineID, err := id.UnifyMachineID()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(guid, guidV7, shortID, ksuid, xid, mongoID, sfID, sonyID, machineID)
}
```

## API Reference

### UUID / Short IDs

```go
func NewGUIDv4(withHyphen bool) string
func NewGUIDv7(withHyphen bool) string   // falls back to v4 if UUIDv7 generation fails
func NewShortUUID() string
func NewKSUID() string
func NewXID() string
func NewMongoObjectID() string
```

### Snowflake / Sonyflake

```go
func NewSnowflakeID(workerId int64) (int64, error)
func GenerateSnowflakeID(workerId int64) int64   // ignores errors, returns 0 on failure

type SnowflakeNode struct{ /* ... */ }

func NewSnowflakeNode(workerId int64) (*SnowflakeNode, error)
func (sfNode *SnowflakeNode) Generate() int64
func (sfNode *SnowflakeNode) GenerateString() string

func NewSonyflakeID() (uint64, error)            // shares a single Sonyflake instance per process
func GenerateSonyflakeID() uint64                // ignores errors, returns 0 on failure
```

- `NewSnowflakeID` / `GenerateSnowflakeID` cache `SnowflakeNode` instances by `workerId` in a process-wide `sync.Map`; the same `workerId` reuses the same node
- Use `NewSnowflakeNode` directly when you need to hold an independent node instance for a long time (e.g. custom serialization scenarios)
- The valid `workerId` range is 0–1023 (a 10-bit node constraint from the upstream `bwmarrin/snowflake` library); out-of-range values make `NewSnowflakeNode` return an error

### Order Numbers

```go
func GenerateOrderIdWithRandom(prefix string, tm *time.Time) string          // prefix + 14-digit timestamp + 4-digit random number
func GenerateOrderIdWithIncreaseIndex(prefix string, tm *time.Time) string   // prefix + 14-digit timestamp + auto-incrementing index
func GenerateOrderIdWithTenantId(tenantID string) string                     // 14-digit timestamp + 5-char tenant ID + 4-digit random number
func GenerateOrderIdWithPrefixSonyflake(prefix string) string
func GenerateOrderIdWithPrefixSnowflake(workerId int64, prefix string) string
```

- Passing `nil` for `tm` uses the current time
- The auto-increment index is an in-process `uint32` counter: it starts at 0 and wraps back to 0 upon reaching 1000 (via `atomic.CompareAndSwap`); it is not persisted
- `GenerateOrderIdWithTenantId`: tenant IDs longer than 5 chars are truncated; shorter ones are right-padded with `'0'`

### Machine ID

```go
type FormatOption struct {
	UpperCase  bool // true: uppercase, false: lowercase
	WithHyphen bool // true: with hyphens, false: without
}

func UnifyMachineID() (string, error)                  // equivalent to FormatMachineID{lowercase, no hyphens}
func FormatMachineID(opt FormatOption) (string, error)
```

- `UnifyMachineID` returns a 32-char lowercase hexadecimal string
- `FormatMachineID` supports upper/lowercase and GUID hyphen format (8-4-4-4-12)
- If the cleaned underlying value is shorter than 32 hex chars, it automatically falls back to the first 32 chars of its SHA-256 hash, keeping the output length stable

## Notes

- The underlying machine ID is read only once per process (`sync.Once` cache) by `UnifyMachineID` / `FormatMachineID`; later calls always return the cached value, so changing the system machine ID at runtime is not reflected
- The `Generate*` helpers swallow underlying errors (returning the zero value on failure); prefer the `New*` functions and handle `error` in business code
- Snowflake/Sonyflake rely on a monotonic clock; in multi-process deployments make sure each process uses a distinct `workerId`
- This package depends on several mature third-party libraries (`google/uuid`, `bwmarrin/snowflake`, `sony/sonyflake`, `segmentio/ksuid`, `rs/xid`, `lithammer/shortuuid`, `mongo-driver`, etc.)

## Platform Support

Machine IDs are backed by the [`id/machineid`](./machineid/README.md) subpackage and work on Windows, Linux, macOS, BSD/Solaris, AIX, and z/OS without admin privileges.

## Testing

```bash
go test ./...
```

## License

MIT
