package ddlparser

// ColumnDef 列定义
type ColumnDef struct {
	Name          string
	Type          string // 原始类型（如 "VARCHAR(255)"）
	Nullable      bool
	PrimaryKey    bool
	Default       string
	Comment       string
	AutoIncrement bool
	Unique        bool
}

// TableDef 表定义
type TableDef struct {
	Name      string
	Columns   []ColumnDef
	Indexes   []string // 简化：仅存储索引定义字符串
	Engine    string   // MySQL 特有
	Charset   string   // MySQL 特有
	Comment   string
	Collation string
}
