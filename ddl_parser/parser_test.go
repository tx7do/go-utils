package ddlparser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCreateTable_BasicTable(t *testing.T) {
	sql := `CREATE TABLE users (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255)
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "users", table.Name)
	assert.Len(t, table.Columns, 3)

	// 验证 id 字段
	assert.Equal(t, "id", table.Columns[0].Name)
	assert.Equal(t, "int", table.Columns[0].Type)
	assert.True(t, table.Columns[0].PrimaryKey)

	// 验证 name 字段
	assert.Equal(t, "name", table.Columns[1].Name)
	assert.Equal(t, "varchar(255)", table.Columns[1].Type)
	assert.False(t, table.Columns[1].Nullable)

	// 验证 email 字段
	assert.Equal(t, "email", table.Columns[2].Name)
	assert.Equal(t, "varchar(255)", table.Columns[2].Type)
	assert.True(t, table.Columns[2].Nullable)
}

func TestParseCreateTable_MySQLWithEngineAndCharset(t *testing.T) {
	sql := `CREATE TABLE products (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL COMMENT '产品名称',
		price DECIMAL(10,2) DEFAULT 0.00
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "products", table.Name)
	assert.Equal(t, "innodb", table.Engine)
	assert.Equal(t, "utf8mb4", table.Charset)

	// 验证 id 字段（auto_increment）
	assert.Equal(t, "id", table.Columns[0].Name)
	assert.True(t, table.Columns[0].PrimaryKey)

	// 验证 name 字段（带注释）
	assert.Equal(t, "name", table.Columns[1].Name)
	assert.Equal(t, "产品名称", table.Columns[1].Comment)

	// 验证 price 字段（带默认值）
	assert.Equal(t, "price", table.Columns[2].Name)
	assert.Equal(t, "0.00", table.Columns[2].Default)
}

func TestParseCreateTable_IfNotExists(t *testing.T) {
	sql := `CREATE TABLE IF NOT EXISTS orders (
		order_id BIGINT PRIMARY KEY,
		user_id BIGINT NOT NULL
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "orders", table.Name)
	assert.Len(t, table.Columns, 2)
}

func TestParseCreateTable_WithBackticks(t *testing.T) {
	sql := "CREATE TABLE `user_profiles` (\n" +
		"	`user_id` INT NOT NULL,\n" +
		"	`profile_data` TEXT\n" +
		")"

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "user_profiles", table.Name)
	assert.Equal(t, "user_id", table.Columns[0].Name)
	assert.Equal(t, "profile_data", table.Columns[1].Name)
}

func TestParseCreateTable_WithComments(t *testing.T) {
	sql := `/* 创建用户表 */
	CREATE TABLE users (
		id INT PRIMARY KEY, -- 用户ID
		-- 用户名称
		name VARCHAR(100)
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "users", table.Name)
	assert.Len(t, table.Columns, 2)
}

func TestParseCreateTable_ComplexTypes(t *testing.T) {
	sql := `CREATE TABLE test_types (
		id BIGINT PRIMARY KEY,
		amount DECIMAL(10, 2),
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN NOT NULL DEFAULT true
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "test_types", table.Name)
	assert.Len(t, table.Columns, 5)

	// 验证 DECIMAL 类型
	assert.Equal(t, "amount", table.Columns[1].Name)
	assert.Contains(t, table.Columns[1].Type, "decimal")

	// 验证 TEXT 类型
	assert.Equal(t, "description", table.Columns[2].Name)
	assert.Equal(t, "text", table.Columns[2].Type)

	// 验证带默认值的 TIMESTAMP
	assert.Equal(t, "created_at", table.Columns[3].Name)
	assert.NotEmpty(t, table.Columns[3].Default)
}

func TestParseCreateTable_WithTableConstraints(t *testing.T) {
	sql := `CREATE TABLE orders (
		id INT,
		user_id INT NOT NULL,
		total DECIMAL(10,2),
		PRIMARY KEY (id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		INDEX idx_user (user_id)
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "orders", table.Name)
	// 表级约束应被跳过，只解析字段定义
	assert.Equal(t, 3, len(table.Columns))
	assert.Equal(t, "id", table.Columns[0].Name)
	assert.Equal(t, "user_id", table.Columns[1].Name)
	assert.Equal(t, "total", table.Columns[2].Name)
}

func TestParseCreateTable_MultipleEngineFormats(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		engine  string
		charset string
	}{
		{
			name:    "ENGINE and CHARSET",
			sql:     "CREATE TABLE t1 (id INT) ENGINE=InnoDB CHARSET=utf8",
			engine:  "innodb",
			charset: "utf8",
		},
		{
			name:    "ENGINE and DEFAULT CHARSET",
			sql:     "CREATE TABLE t2 (id INT) ENGINE=MyISAM DEFAULT CHARSET=latin1",
			engine:  "myisam",
			charset: "latin1",
		},
		{
			name:    "Only ENGINE",
			sql:     "CREATE TABLE t3 (id INT) ENGINE=Memory",
			engine:  "memory",
			charset: "",
		},
		{
			name:    "No attributes",
			sql:     "CREATE TABLE t4 (id INT)",
			engine:  "",
			charset: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			require.NoError(t, err)
			assert.Equal(t, tt.engine, table.Engine)
			assert.Equal(t, tt.charset, table.Charset)
		})
	}
}

func TestParseCreateTable_CaseInsensitive(t *testing.T) {
	sql := `create table Users (
		ID int primary key,
		NAME varchar(100) NOT NULL
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "users", table.Name)
	assert.Equal(t, "id", table.Columns[0].Name)
	assert.Equal(t, "name", table.Columns[1].Name)
}

func TestParseCreateTable_NullableFields(t *testing.T) {
	sql := `CREATE TABLE test (
		field1 INT NULL,
		field2 INT NOT NULL,
		field3 INT
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	// field1 explicitly NULL
	assert.True(t, table.Columns[0].Nullable)

	// field2 explicitly NOT NULL
	assert.False(t, table.Columns[1].Nullable)

	// field3 默认可为空
	assert.True(t, table.Columns[2].Nullable)
}

func TestParseCreateTable_DefaultValues(t *testing.T) {
	sql := `CREATE TABLE settings (
		id INT PRIMARY KEY,
		timeout INT DEFAULT 30,
		name VARCHAR(50) DEFAULT 'unknown',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_enabled BOOLEAN DEFAULT false
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "", table.Columns[0].Default) // 主键无默认值
	assert.Equal(t, "30", table.Columns[1].Default)
	assert.Equal(t, "'unknown'", table.Columns[2].Default)
	assert.NotEmpty(t, table.Columns[3].Default)
	assert.Equal(t, "false", table.Columns[4].Default)
}

func TestParseCreateTable_InvalidSQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "No table name",
			sql:  "CREATE TABLE (id INT)",
		},
		{
			name: "No parentheses",
			sql:  "CREATE TABLE users",
		},
		{
			name: "Mismatched parentheses",
			sql:  "CREATE TABLE users (id INT",
		},
		{
			name: "Not a CREATE TABLE",
			sql:  "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			assert.Error(t, err)
			assert.Nil(t, table)
		})
	}
}

func TestParseCreateTable_EmptyTable(t *testing.T) {
	sql := `CREATE TABLE empty_table ()`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)
	assert.Equal(t, "empty_table", table.Name)
	assert.Len(t, table.Columns, 0)
}

func TestParseCreateTable_ComplexRealWorld(t *testing.T) {
	sql := `
	/* 用户订单表 */
	CREATE TABLE IF NOT EXISTS user_orders (
		order_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '订单ID',
		user_id BIGINT NOT NULL COMMENT '用户ID',
		order_number VARCHAR(50) NOT NULL COMMENT '订单号',
		total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '总金额',
		status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		deleted_at TIMESTAMP NULL COMMENT '删除时间',
		-- 表级约束
		INDEX idx_user_id (user_id),
		INDEX idx_order_number (order_number),
		CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户订单表';
	`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)

	assert.Equal(t, "user_orders", table.Name)
	assert.Equal(t, "innodb", table.Engine)
	assert.Equal(t, "utf8mb4", table.Charset)

	// 验证字段数量（表级约束被跳过）
	assert.GreaterOrEqual(t, len(table.Columns), 6)

	// 验证主键字段
	orderIdCol := table.Columns[0]
	assert.Equal(t, "order_id", orderIdCol.Name)
	assert.True(t, orderIdCol.PrimaryKey)
	assert.Equal(t, "订单ID", orderIdCol.Comment)

	// 验证带默认值的字段
	var totalAmountCol *ColumnDef
	for i := range table.Columns {
		if table.Columns[i].Name == "total_amount" {
			totalAmountCol = &table.Columns[i]
			break
		}
	}
	if totalAmountCol != nil {
		assert.Equal(t, "0.00", totalAmountCol.Default)
		assert.False(t, totalAmountCol.Nullable)
	}
}

// ============================================================================
// MySQL 方言特性测试
// ============================================================================

func TestMySQL_AutoIncrement(t *testing.T) {
	sql := `CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100)
	) ENGINE=InnoDB`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "users", table.Name)
	assert.True(t, table.Columns[0].PrimaryKey, "AUTO_INCREMENT should imply PRIMARY KEY")
	assert.Equal(t, "innodb", table.Engine)
}

func TestMySQL_EngineTypes(t *testing.T) {
	engines := []string{"InnoDB", "MyISAM", "Memory", "Archive", "CSV"}

	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			sql := fmt.Sprintf("CREATE TABLE test (id INT) ENGINE=%s", engine)
			table, err := ParseCreateTable(sql)
			require.NoError(t, err)
			assert.Equal(t, strings.ToLower(engine), table.Engine)
		})
	}
}

func TestMySQL_CharsetAndCollation(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		charset string
	}{
		{
			name:    "UTF8",
			sql:     "CREATE TABLE t (id INT) CHARSET=utf8",
			charset: "utf8",
		},
		{
			name:    "UTF8MB4",
			sql:     "CREATE TABLE t (id INT) DEFAULT CHARSET=utf8mb4",
			charset: "utf8mb4",
		},
		{
			name:    "Latin1",
			sql:     "CREATE TABLE t (id INT) CHARACTER SET latin1",
			charset: "latin1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			require.NoError(t, err)
			assert.Equal(t, tt.charset, table.Charset)
		})
	}
}

func TestMySQL_DataTypes(t *testing.T) {
	sql := `CREATE TABLE mysql_types (
		tiny_col TINYINT,
		small_col SMALLINT,
		medium_col MEDIUMINT,
		int_col INT,
		big_col BIGINT,
		decimal_col DECIMAL(10,2),
		float_col FLOAT,
		double_col DOUBLE,
		char_col CHAR(10),
		varchar_col VARCHAR(255),
		text_col TEXT,
		mediumtext_col MEDIUMTEXT,
		longtext_col LONGTEXT,
		blob_col BLOB,
		date_col DATE,
		datetime_col DATETIME,
		timestamp_col TIMESTAMP,
		year_col YEAR,
		enum_col ENUM('a','b','c'),
		set_col SET('x','y','z'),
		json_col JSON
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)
	assert.Equal(t, "mysql_types", table.Name)

	// 验证一些关键类型
	typeMap := make(map[string]string)
	for _, col := range table.Columns {
		typeMap[col.Name] = col.Type
	}

	assert.Contains(t, typeMap["tiny_col"], "tinyint")
	assert.Contains(t, typeMap["varchar_col"], "varchar")
	assert.Contains(t, typeMap["decimal_col"], "decimal")
	assert.Contains(t, typeMap["json_col"], "json")
}

func TestMySQL_Timestamps(t *testing.T) {
	sql := `CREATE TABLE events (
		id INT PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "events", table.Name)
	assert.Len(t, table.Columns, 3)

	// created_at 应该有默认值
	assert.NotEmpty(t, table.Columns[1].Default)
	assert.Contains(t, strings.ToLower(table.Columns[1].Default), "current_timestamp")
}

func TestMySQL_UnsignedAndZerofill(t *testing.T) {
	sql := `CREATE TABLE test (
		id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		amount DECIMAL(10,2) UNSIGNED,
		code INT ZEROFILL
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "test", table.Name)
	assert.Len(t, table.Columns, 3)
}

func TestMySQL_FullTextIndex(t *testing.T) {
	sql := `CREATE TABLE articles (
		id INT PRIMARY KEY,
		title VARCHAR(200),
		content TEXT,
		FULLTEXT INDEX ft_content (content),
		FULLTEXT INDEX ft_title_content (title, content)
	) ENGINE=InnoDB`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "articles", table.Name)
	// 字段定义应该被正确解析，FULLTEXT 索引被跳过
	assert.Equal(t, 3, len(table.Columns))
}

func TestMySQL_PartitionedTable(t *testing.T) {
	sql := `CREATE TABLE sales (
		id INT,
		sale_date DATE,
		amount DECIMAL(10,2)
	) ENGINE=InnoDB
	PARTITION BY RANGE(YEAR(sale_date)) (
		PARTITION p0 VALUES LESS THAN (2020),
		PARTITION p1 VALUES LESS THAN (2021)
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "sales", table.Name)
	assert.Equal(t, "innodb", table.Engine)
}

// ============================================================================
// PostgreSQL 方言特性测试
// ============================================================================

func TestPostgreSQL_SerialTypes(t *testing.T) {
	sql := `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		big_id BIGSERIAL,
		small_id SMALLSERIAL
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "users", table.Name)
	assert.Len(t, table.Columns, 3)

	// SERIAL 类型应该被识别
	assert.Contains(t, strings.ToLower(table.Columns[0].Type), "serial")
}

func TestPostgreSQL_DataTypes(t *testing.T) {
	sql := `CREATE TABLE pg_types (
		uuid_col UUID,
		json_col JSON,
		jsonb_col JSONB,
		array_col INTEGER[],
		text_array_col TEXT[],
		hstore_col HSTORE,
		inet_col INET,
		cidr_col CIDR,
		macaddr_col MACADDR,
		money_col MONEY,
		interval_col INTERVAL,
		point_col POINT,
		line_col LINE,
		box_col BOX,
		path_col PATH,
		polygon_col POLYGON,
		circle_col CIRCLE
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.NotNil(t, table)
	assert.Equal(t, "pg_types", table.Name)

	typeMap := make(map[string]string)
	for _, col := range table.Columns {
		typeMap[col.Name] = col.Type
	}

	assert.Contains(t, typeMap["uuid_col"], "uuid")
	assert.Contains(t, typeMap["jsonb_col"], "jsonb")
	assert.Contains(t, typeMap["inet_col"], "inet")
}

func TestPostgreSQL_ArrayTypes(t *testing.T) {
	sql := `CREATE TABLE test (
		id SERIAL PRIMARY KEY,
		tags TEXT[],
		numbers INTEGER[],
		matrix INTEGER[][]
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "test", table.Name)
	assert.Len(t, table.Columns, 4)
}

func TestPostgreSQL_CheckConstraint(t *testing.T) {
	sql := `CREATE TABLE products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10,2) CHECK (price > 0),
		quantity INT CHECK (quantity >= 0)
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "products", table.Name)
	assert.Len(t, table.Columns, 4)
}

func TestPostgreSQL_DefaultValues(t *testing.T) {
	sql := `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		uuid UUID DEFAULT gen_random_uuid()
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "users", table.Name)
	// 默认值应该被提取
	assert.NotEmpty(t, table.Columns[1].Default)
}

func TestPostgreSQL_GeneratedColumns(t *testing.T) {
	sql := `CREATE TABLE people (
		id SERIAL PRIMARY KEY,
		first_name TEXT,
		last_name TEXT,
		full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "people", table.Name)
	// 生成列应该被识别为字段
	assert.GreaterOrEqual(t, len(table.Columns), 3)
}

func TestPostgreSQL_InheritedTables(t *testing.T) {
	sql := `CREATE TABLE employees (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		salary DECIMAL(10,2)
	) INHERITS (persons)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "employees", table.Name)
	assert.Len(t, table.Columns, 3)
}

// ============================================================================
// SQLite 方言特性测试
// ============================================================================

func TestSQLite_AutoIncrement(t *testing.T) {
	sql := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "users", table.Name)
	assert.True(t, table.Columns[0].PrimaryKey)
}

func TestSQLite_DataTypes(t *testing.T) {
	sql := `CREATE TABLE sqlite_types (
		int_col INTEGER,
		text_col TEXT,
		real_col REAL,
		blob_col BLOB,
		numeric_col NUMERIC
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "sqlite_types", table.Name)
	assert.Len(t, table.Columns, 5)

	typeMap := make(map[string]string)
	for _, col := range table.Columns {
		typeMap[col.Name] = col.Type
	}

	assert.Contains(t, typeMap["int_col"], "integer")
	assert.Contains(t, typeMap["text_col"], "text")
	assert.Contains(t, typeMap["real_col"], "real")
	assert.Contains(t, typeMap["blob_col"], "blob")
}

func TestSQLite_WithoutRowID(t *testing.T) {
	sql := `CREATE TABLE config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	) WITHOUT ROWID`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "config", table.Name)
	assert.Len(t, table.Columns, 2)
	assert.True(t, table.Columns[0].PrimaryKey)
}

func TestSQLite_DefaultValues(t *testing.T) {
	sql := `CREATE TABLE logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message TEXT NOT NULL,
		created_at TEXT DEFAULT (datetime('now')),
		level TEXT DEFAULT 'INFO'
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "logs", table.Name)
	assert.Len(t, table.Columns, 4)
}

func TestSQLite_StrictTables(t *testing.T) {
	sql := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		age INTEGER
	) STRICT`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "users", table.Name)
	assert.Len(t, table.Columns, 3)
}

func TestSQLite_GeneratedColumns(t *testing.T) {
	sql := `CREATE TABLE rectangle (
		width REAL,
		height REAL,
		area REAL GENERATED ALWAYS AS (width * height) STORED
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "rectangle", table.Name)
	assert.GreaterOrEqual(t, len(table.Columns), 2)
}

// ============================================================================
// 跨数据库兼容性测试
// ============================================================================

func TestCrossDB_CommonDataTypes(t *testing.T) {
	sql := `CREATE TABLE universal (
		id INTEGER PRIMARY KEY,
		name VARCHAR(100),
		amount DECIMAL(10,2),
		description TEXT,
		created_at TIMESTAMP,
		is_active BOOLEAN
	)`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)

	assert.Equal(t, "universal", table.Name)
	assert.Len(t, table.Columns, 6)

	// 验证通用类型被正确解析
	typeMap := make(map[string]string)
	for _, col := range table.Columns {
		typeMap[col.Name] = col.Type
	}

	assert.Contains(t, typeMap["id"], "integer")
	assert.Contains(t, typeMap["name"], "varchar")
	assert.Contains(t, typeMap["amount"], "decimal")
}

func TestCrossDB_NullableConstraints(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "MySQL style",
			sql:  "CREATE TABLE t (id INT NOT NULL, name VARCHAR(50) NULL)",
		},
		{
			name: "PostgreSQL style",
			sql:  "CREATE TABLE t (id INTEGER NOT NULL, name TEXT)",
		},
		{
			name: "SQLite style",
			sql:  "CREATE TABLE t (id INTEGER NOT NULL, name TEXT)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			require.NoError(t, err)
			assert.NotNil(t, table)

			// id 应该是 NOT NULL
			assert.False(t, table.Columns[0].Nullable)
		})
	}
}

func TestCrossDB_PrimaryKeyVariants(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		pkName string
	}{
		{
			name:   "Inline PRIMARY KEY",
			sql:    "CREATE TABLE t (id INT PRIMARY KEY, name TEXT)",
			pkName: "id",
		},
		{
			name:   "Composite via inline",
			sql:    "CREATE TABLE t (id1 INT, id2 INT, name TEXT, PRIMARY KEY(id1, id2))",
			pkName: "id1", // 第一个字段
		},
		{
			name:   "MySQL AUTO_INCREMENT",
			sql:    "CREATE TABLE t (id INT AUTO_INCREMENT PRIMARY KEY)",
			pkName: "id",
		},
		{
			name:   "PostgreSQL SERIAL",
			sql:    "CREATE TABLE t (id SERIAL PRIMARY KEY)",
			pkName: "id",
		},
		{
			name:   "SQLite AUTOINCREMENT",
			sql:    "CREATE TABLE t (id INTEGER PRIMARY KEY AUTOINCREMENT)",
			pkName: "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			require.NoError(t, err)

			// 至少应该有一个主键
			hasPK := false
			for _, col := range table.Columns {
				if col.PrimaryKey {
					hasPK = true
					break
				}
			}
			assert.True(t, hasPK, "Should have at least one primary key column")
		})
	}
}

func TestCrossDB_QuotedIdentifiers(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		tableName string
		colName   string
	}{
		{
			name:      "MySQL backticks",
			sql:       "CREATE TABLE `my_table` (`my_column` INT)",
			tableName: "my_table",
			colName:   "my_column",
		},
		{
			name:      "PostgreSQL double quotes",
			sql:       `CREATE TABLE "my_table" ("my_column" INT)`,
			tableName: "my_table",
			colName:   "my_column",
		},
		{
			name:      "SQLite double quotes",
			sql:       `CREATE TABLE "my_table" ("my_column" INTEGER)`,
			tableName: "my_table",
			colName:   "my_column",
		},
		{
			name:      "Mixed quotes",
			sql:       "CREATE TABLE `table1` (\"col1\" INT, `col2` TEXT)",
			tableName: "table1",
			colName:   "col1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := ParseCreateTable(tt.sql)
			require.NoError(t, err)
			assert.Equal(t, tt.tableName, table.Name)
			if len(table.Columns) > 0 {
				assert.Equal(t, tt.colName, table.Columns[0].Name)
			}
		})
	}
}

func TestParseCreateTables_MultipleCreate(t *testing.T) {
	sql := `
	CREATE TABLE users (
		id INT PRIMARY KEY,
		name VARCHAR(100)
	);
	CREATE TABLE orders (
		order_id INT PRIMARY KEY,
		user_id INT NOT NULL
	);
	`

	tables, err := ParseCreateTables(sql)
	require.NoError(t, err)
	require.Len(t, tables, 2)
	assert.Equal(t, "users", tables[0].Name)
	assert.Equal(t, "orders", tables[1].Name)
}

func TestParseCreateTables_IgnoresNonCreate(t *testing.T) {
	sql := `
	CREATE TABLE users (id INT PRIMARY KEY);
	SELECT * FROM users;
	CREATE TABLE logs (id INT, msg TEXT);
	`

	tables, err := ParseCreateTables(sql)
	require.NoError(t, err)
	require.Len(t, tables, 2)
	assert.Equal(t, "users", tables[0].Name)
	assert.Equal(t, "logs", tables[1].Name)
}

func TestParseCreateTables_SemicolonInString(t *testing.T) {
	sql := `
	CREATE TABLE messages (
		id INT PRIMARY KEY,
		content VARCHAR(100) DEFAULT 'a; b; c'
	);
	CREATE TABLE audit (
		id INT PRIMARY KEY,
		note TEXT
	);
	`

	tables, err := ParseCreateTables(sql)
	require.NoError(t, err)
	require.Len(t, tables, 2)
	assert.Equal(t, "messages", tables[0].Name)
	assert.Equal(t, "audit", tables[1].Name)
}

func ExampleParseCreateTables() {
	sql := `
	CREATE TABLE users (id INT PRIMARY KEY, name TEXT);
	CREATE TABLE orders (id INT PRIMARY KEY, user_id INT);
	`

	tables, err := ParseCreateTables(sql)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, t := range tables {
		fmt.Println(t.Name)
	}
	// Output:
	// users
	// orders
}

func TestParseCreateTable_Comment(t *testing.T) {
	sql := `CREATE TABLE users (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	) COMMENT='用户表'`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "users", table.Name)
	assert.Equal(t, "用户表", table.Comment)
}

func TestParseCreateTable_CommentWithDoubleQuotes(t *testing.T) {
	sql := `CREATE TABLE products (
		id INT PRIMARY KEY,
		name VARCHAR(255)
	) COMMENT="产品表"`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "products", table.Name)
	assert.Equal(t, "产品表", table.Comment)
}

func TestParseCreateTable_Collation(t *testing.T) {
	sql := `CREATE TABLE orders (
		id INT PRIMARY KEY,
		content TEXT
	) COLLATION=utf8mb4_unicode_ci`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "orders", table.Name)
	assert.Equal(t, "utf8mb4_unicode_ci", table.Collation)
}

func TestParseCreateTable_CollateKeyword(t *testing.T) {
	sql := `CREATE TABLE articles (
		id INT PRIMARY KEY,
		title VARCHAR(255)
	) COLLATE=utf8mb4_general_ci`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "articles", table.Name)
	assert.Equal(t, "utf8mb4_general_ci", table.Collation)
}

func TestParseCreateTable_CommentAndCollation(t *testing.T) {
	sql := `CREATE TABLE users_v2 (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表'`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "users_v2", table.Name)
	assert.Equal(t, "utf8mb4", table.Charset)
	assert.Equal(t, "utf8mb4_unicode_ci", table.Collation)
	assert.Equal(t, "用户表", table.Comment)
	assert.Equal(t, "innodb", table.Engine)
}

func TestParseCreateTable_MultipleAttributes(t *testing.T) {
	sql := `CREATE TABLE logs (
		id BIGINT PRIMARY KEY,
		message TEXT NOT NULL,
		created_at TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATION=utf8mb4_bin COMMENT="系统日志表"`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "logs", table.Name)
	assert.Equal(t, "innodb", table.Engine)
	assert.Equal(t, "utf8mb4", table.Charset)
	assert.Equal(t, "utf8mb4_bin", table.Collation)
	assert.Equal(t, "系统日志表", table.Comment)
}

func TestParseCreateTable_CommentWithSpecialChars(t *testing.T) {
	sql := `CREATE TABLE config (
		id INT PRIMARY KEY,
		value TEXT
	) COMMENT='配置表: 用于存储系统配置'`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "config", table.Name)
	assert.Equal(t, "配置表: 用于存储系统配置", table.Comment)
}

func TestParseCreateTable_Latin1Collation(t *testing.T) {
	sql := `CREATE TABLE archive (
		id INT PRIMARY KEY,
		data VARCHAR(255)
	) DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci`

	table, err := ParseCreateTable(sql)
	require.NoError(t, err)
	assert.Equal(t, "archive", table.Name)
	assert.Equal(t, "latin1", table.Charset)
	assert.Equal(t, "latin1_general_ci", table.Collation)
}
