package ddlparser

import (
	"fmt"
	"regexp"
	"strings"
)

// normalizeSQL 标准化：转小写（保留引号内内容）、移除注释
func normalizeSQL(sql string) string {
	// 移除多行注释 /* ... */
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, " ")

	// 移除单行注释 -- ...
	sql = regexp.MustCompile(`--.*?$`).ReplaceAllString(sql, " ")

	// 保留引号内内容，其余转小写
	var result strings.Builder
	inSingle, inDouble, inBacktick := false, false, false
	for _, ch := range sql {
		// 处理引号切换（简化版，不处理转义）
		if ch == '\'' && !inDouble && !inBacktick {
			inSingle = !inSingle
		} else if ch == '"' && !inSingle && !inBacktick {
			inDouble = !inDouble
		} else if ch == '`' && !inSingle && !inDouble {
			inBacktick = !inBacktick
		}

		if inSingle || inDouble || inBacktick {
			result.WriteRune(ch)
		} else {
			result.WriteRune(rune(strings.ToLower(string(ch))[0]))
		}
	}

	return strings.Join(strings.Fields(result.String()), " ")
}

// extractTableName 提取表名（支持多种格式）
func extractTableName(sql string) (string, error) {
	// 匹配: CREATE TABLE [IF NOT EXISTS] table_name
	re := regexp.MustCompile(`(?i)create\s+table\s+(if\s+not\s+exists\s+)?([` + "`" + `"\w.]+)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 3 {
		// 安全截取 SQL 字符串用于错误消息
		maxLen := 50
		if len(sql) < maxLen {
			maxLen = len(sql)
		}
		return "", fmt.Errorf("无法提取表名: %s", sql[:maxLen])
	}

	name := matches[2]
	// 移除引号/反引号
	name = strings.Trim(name, "`\"")
	return name, nil
}

// extractColumnBlock 提取括号内的字段定义
func extractColumnBlock(sql string) (string, error) {
	// 找到第一个左括号和匹配的右括号（处理嵌套括号）
	leftIdx := strings.Index(sql, "(")
	if leftIdx == -1 {
		return "", fmt.Errorf("未找到字段定义块")
	}

	// 简化：假设第一层括号即为字段定义（CREATE TABLE 通常如此）
	rightIdx := findMatchingParen(sql, leftIdx)
	if rightIdx == -1 {
		return "", fmt.Errorf("括号不匹配")
	}

	return sql[leftIdx+1 : rightIdx], nil
}

// findMatchingParen 寻找匹配的右括号（处理嵌套）
func findMatchingParen(sql string, leftIdx int) int {
	level := 0
	for i := leftIdx; i < len(sql); i++ {
		if sql[i] == '(' {
			level++
		} else if sql[i] == ')' {
			level--
			if level == 0 {
				return i
			}
		}
	}
	return -1
}

// parseColumns 解析字段定义
func parseColumns(block string) ([]ColumnDef, error) {
	// 分割字段（逗号分隔,但不在括号内）
	var columns []ColumnDef
	var primaryKeyColumns []string
	parts := splitColumns(block)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否为表级约束（更精确的匹配）
		partLower := strings.ToLower(part)

		// 检查是否为表级 PRIMARY KEY 约束
		if strings.HasPrefix(partLower, "primary key") {
			// 提取主键列名：PRIMARY KEY (col1, col2, ...)
			pkCols := extractPrimaryKeyColumns(part)
			primaryKeyColumns = append(primaryKeyColumns, pkCols...)
			continue
		}

		isConstraint := strings.HasPrefix(partLower, "foreign key") ||
			strings.HasPrefix(partLower, "constraint") ||
			strings.HasPrefix(partLower, "fulltext") ||
			strings.HasPrefix(partLower, "spatial") ||
			strings.HasPrefix(partLower, "unique key") ||
			strings.HasPrefix(partLower, "unique index") ||
			// 匹配 KEY/INDEX 后面跟着名称和括号的模式（表级约束）
			regexp.MustCompile(`^(key|index)\s+\w+\s*\(`).MatchString(partLower) ||
			// 匹配单独的 KEY/INDEX 后面跟着括号（匿名索引）
			regexp.MustCompile(`^(key|index)\s*\(`).MatchString(partLower)

		if isConstraint {
			// 表级约束，暂不处理
			continue
		}

		col, err := parseColumn(part)
		if err != nil {
			// 容错：跳过无法解析的字段
			continue
		}
		columns = append(columns, col)
	}

	// 标记表级主键约束引用的列
	if len(primaryKeyColumns) > 0 {
		for i := range columns {
			for _, pkCol := range primaryKeyColumns {
				if columns[i].Name == pkCol {
					columns[i].PrimaryKey = true
					break
				}
			}
		}
	}

	return columns, nil
}

// extractPrimaryKeyColumns 从表级 PRIMARY KEY 约束中提取列名
func extractPrimaryKeyColumns(constraintDef string) []string {
	// 匹配 PRIMARY KEY (col1, col2, ...)
	re := regexp.MustCompile(`primary\s+key\s*\(([^)]+)\)`)
	matches := re.FindStringSubmatch(strings.ToLower(constraintDef))
	if len(matches) < 2 {
		return nil
	}

	// 分割列名
	colsPart := matches[1]
	var cols []string
	for _, col := range strings.Split(colsPart, ",") {
		col = strings.TrimSpace(col)
		col = strings.Trim(col, "`\"") // 移除引号
		if col != "" {
			cols = append(cols, col)
		}
	}
	return cols
}

// splitColumns 智能分割字段（跳过括号内的逗号）
func splitColumns(block string) []string {
	var parts []string
	var current strings.Builder
	parenLevel := 0
	inSingle, inDouble := false, false

	for _, ch := range block {
		switch ch {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				parenLevel++
			}
		case ')':
			if !inSingle && !inDouble {
				parenLevel--
			}
		case ',':
			if parenLevel == 0 && !inSingle && !inDouble {
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseColumn 解析单个字段
func parseColumn(def string) (ColumnDef, error) {
	// 基础模式：`name type [constraints]`
	// 示例: "id int auto_increment primary key"
	//       "name varchar(255) not null comment '用户名'"

	// 首先提取 COMMENT（因为它可能包含空格）
	var comment string
	defLower := strings.ToLower(def)
	commentIdx := strings.Index(defLower, "comment")
	if commentIdx != -1 {
		// 提取 COMMENT 之后的内容
		commentPart := def[commentIdx+7:] // "comment" 长度为 7
		commentPart = strings.TrimSpace(commentPart)

		// 匹配引号内的内容
		if len(commentPart) > 0 {
			quote := commentPart[0]
			if quote == '\'' || quote == '"' {
				// 找到匹配的结束引号
				endIdx := strings.IndexRune(commentPart[1:], rune(quote))
				if endIdx != -1 {
					comment = commentPart[1 : endIdx+1]
					// 从定义中移除 COMMENT 部分
					def = def[:commentIdx]
				}
			}
		}
	}

	// 提取字段名（第一个单词，可能带引号或反引号）
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return ColumnDef{}, fmt.Errorf("字段定义过短: %s", def)
	}

	col := ColumnDef{
		Name:     strings.Trim(parts[0], "`\""), // 移除字段名的引号/反引号
		Nullable: true,                          // 默认可为空
		Comment:  comment,                       // 设置已提取的注释
	}

	// 提取类型（合并可能带括号的类型，如 varchar(255) 或 decimal(10, 2)）
	typeParts := []string{strings.Trim(parts[1], "`\"")} // 也移除类型的引号/反引号
	i := 2
	// 继续收集类型定义直到遇到完整的括号或非类型关键字
	for i < len(parts) && strings.Contains(parts[i-1], "(") && !strings.Contains(parts[i-1], ")") {
		typeParts = append(typeParts, strings.Trim(parts[i], "`\""))
		i++
	}
	col.Type = strings.Join(typeParts, " ")

	// 解析约束
	for j := i; j < len(parts); j++ {
		token := strings.ToLower(parts[j])

		switch token {
		case "not", "null":
			// 处理 "not null"
			if j > 0 && strings.ToLower(parts[j-1]) == "not" {
				col.Nullable = false
			}
		case "primary", "key":
			// 处理 "primary key"
			if j > 0 && strings.ToLower(parts[j-1]) == "primary" {
				col.PrimaryKey = true
			}
		case "default":
			if j+1 < len(parts) {
				col.Default = parts[j+1]
				j++ // 跳过值
			}
		}
	}

	// 特殊处理：MySQL 的 auto_increment 隐含主键
	if strings.Contains(strings.ToLower(def), "auto_increment") {
		col.PrimaryKey = true
	}

	return col, nil
}

// extractTableAttributes 提取表级属性（MySQL 特有）
func extractTableAttributes(sql string) map[string]string {
	attrs := make(map[string]string)

	attrsSQL := sql
	leftIdx := strings.Index(sql, "(")
	if leftIdx != -1 {
		rightIdx := findMatchingParen(sql, leftIdx)
		if rightIdx != -1 && rightIdx+1 < len(sql) {
			// Only parse table attributes after the column block.
			attrsSQL = sql[rightIdx+1:]
		}
	}

	// ENGINE=InnoDB
	if matches := regexp.MustCompile(`engine\s*=\s*(\w+)`).FindStringSubmatch(attrsSQL); len(matches) > 1 {
		attrs["engine"] = matches[1]
	}

	// DEFAULT CHARSET=utf8mb4 或 CHARACTER SET latin1
	if matches := regexp.MustCompile(`(?:default\s+)?(?:charset|character\s+set)\s*=?\s*(\w+)`).FindStringSubmatch(attrsSQL); len(matches) > 1 {
		attrs["charset"] = matches[1]
	}

	// COLLATE=utf8mb4_unicode_ci 或 COLLATION=utf8mb4_unicode_ci
	if matches := regexp.MustCompile(`(?:collate|collation)\s*=\s*([a-zA-Z0-9_]+)`).FindStringSubmatch(attrsSQL); len(matches) > 1 {
		attrs["collation"] = matches[1]
	}

	// COMMENT='...' 或 COMMENT="..." 或 COMMENT '...' 或 COMMENT "..."
	// 修复：支持带等号和不带等号的语法
	if matches := regexp.MustCompile(`comment\s*=?\s*['"]([^'"]*)['"]`).FindStringSubmatch(attrsSQL); len(matches) > 1 {
		attrs["comment"] = matches[1]
	}

	return attrs
}

// splitSQLStatements splits SQL by semicolons, respecting quotes and parentheses.
func splitSQLStatements(sql string) []string {
	var parts []string
	var current strings.Builder

	inSingle, inDouble, inBacktick := false, false, false
	parenLevel := 0

	for _, ch := range sql {
		switch ch {
		case '\'':
			if !inDouble && !inBacktick {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !inBacktick {
				inDouble = !inDouble
			}
		case '`':
			if !inSingle && !inDouble {
				inBacktick = !inBacktick
			}
		case '(':
			if !inSingle && !inDouble && !inBacktick {
				parenLevel++
			}
		case ')':
			if !inSingle && !inDouble && !inBacktick && parenLevel > 0 {
				parenLevel--
			}
		case ';':
			if !inSingle && !inDouble && !inBacktick && parenLevel == 0 {
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
		}

		current.WriteRune(ch)
	}

	if strings.TrimSpace(current.String()) != "" {
		parts = append(parts, current.String())
	}

	return parts
}
