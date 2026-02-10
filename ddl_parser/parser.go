package ddlparser

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseCreateTable 解析 CREATE TABLE 语句（支持多数据库基础语法）
func ParseCreateTable(sql string) (*TableDef, error) {
	sql = normalizeSQL(sql)

	// 1. 提取表名
	tableName, err := extractTableName(sql)
	if err != nil {
		return nil, err
	}

	// 2. 提取字段定义块（位于括号内）
	columnBlock, err := extractColumnBlock(sql)
	if err != nil {
		return nil, err
	}

	// 3. 解析字段
	columns, err := parseColumns(columnBlock)
	if err != nil {
		return nil, err
	}

	// 4. 提取表级属性（ENGINE/CHARSET 等）
	tableAttrs := extractTableAttributes(sql)

	return &TableDef{
		Name:      tableName,
		Columns:   columns,
		Engine:    tableAttrs["engine"],
		Charset:   tableAttrs["charset"],
		Comment:   tableAttrs["comment"],
		Collation: tableAttrs["collation"],
	}, nil
}

// ParseCreateTables parses multiple CREATE TABLE statements in one SQL string.
// Non-CREATE TABLE statements are ignored.
func ParseCreateTables(sql string) ([]*TableDef, error) {
	statements := splitSQLStatements(sql)
	var tables []*TableDef

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Only handle CREATE TABLE statements.
		if !regexp.MustCompile(`(?i)\bcreate\s+table\b`).MatchString(stmt) {
			continue
		}

		table, err := ParseCreateTable(stmt)
		if err != nil {
			return nil, fmt.Errorf("parse create table failed: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}
