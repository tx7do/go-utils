package pagination

const (
	DefaultPage     = 1  // 默认页数
	DefaultPageSize = 10 // 默认每页行数
)

// GetPageOffset 计算偏移量
func GetPageOffset(pageNum, pageSize int32) int {
	return int((pageNum - 1) * pageSize)
}
