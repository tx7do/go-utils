# Protobuf FieldMask Utility

## Filter

保留 msg 中 paths 列表指定的字段,清除所有其他字段。

```go
// 假设有消息: {Name: "Alice", Age: 30, Email: "alice@example.com"}
Filter(msg, []string{"name", "email"})
// 结果: {Name: "Alice", Age: 0, Email: "alice@example.com"}
// Age 字段被清除
```

## Prune

清除 msg 中 paths 列表指定的字段,保留所有其他字段。

```go
// 假设有消息: {Name: "Alice", Age: 30, Email: "alice@example.com"}
Prune(msg, []string{"age", "email"})
// 结果: {Name: "Alice", Age: 0, Email: ""}
// Age 和 Email 字段被清除,Name 保留
```

## Overwrite

使用 src 消息中的值覆盖 dest 消息中 paths 列表指定的字段。

```go
// src: {Name: "Alice", Age: 30, Email: "alice@example.com"}
// dest: {Name: "Bob", Age: 25, Email: "bob@example.com"}
Overwrite(src, dest, []string{"name", "age"})
// 结果 dest: {Name: "Alice", Age: 30, Email: "bob@example.com"}
// Name 和 Age 被 src 的值覆盖,Email 保持不变
```

## Validate

检查所有 paths 对于指定的 validationModel 消息是否有效。

```go
// 验证字段路径是否存在
err := Validate(userMsg, []string{"name", "profile.bio"})
// 如果 userMsg 没有 profile.bio 字段,返回错误
// 如果所有字段都存在,返回 nil
```

## PathsFromFieldNumbers

从给定的字段编号列表中,返回对应的字段路径列表。

```go
// 假设消息定义: message User {
//   string name = 1;
//   int32 age = 2;
//   string email = 3;
// }
paths := PathsFromFieldNumbers(userMsg, []int{1, 3})
// 返回: ["name", "email"]
```

## NilValuePaths

从给定的字段路径列表中,返回那些在消息中不存在或未设置(nil)的字段路径。

```go
// 假设消息: {Name: "Alice", Email: "alice@example.com"}
// Age 字段未设置(为零值)
nilPaths := NilValuePaths(msg, []string{"name", "age", "email"})
// 返回: ["age"]
// 因为只有 age 字段未设置
```
