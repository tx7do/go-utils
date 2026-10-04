# sqlutil

PostgreSQL 词法级的 SQL 审计工具，纯标准库：

- **`MaskSQL`**：字面量级脱敏——字符串字面量（`'…'`、`E'…'`、`$tag$…$tag$` 美元引用）
  与数值字面量整体替换为 `***`，保留 SQL 结构（关键字、双引号标识符、`$n` 占位符、
  注释、`::` 类型转换）。按 PG 词法规则实现：单引号串内 `\'` 与 `''` 均按转义处理
  （宁多掩不漏掩）、美元引用可跨行、块注释支持嵌套、注释与标识符中的引号不误判。
- **`ExtractTables`**：从（已脱敏的）SQL 文本提取被访问的表名，覆盖 FROM /
  INSERT INTO / UPDATE / DELETE FROM / JOIN 及逗号多表，跳过子查询括号，
  schema 限定名保留全名，按首个出现顺序去重。字符串与注释不拆词。

```go
import "github.com/tx7do/go-utils/sqlutil"

masked := sqlutil.MaskSQL(rawSQL)      // 审计落库前脱敏
tables := sqlutil.ExtractTables(masked) // 数据分类/访问面分析
```
