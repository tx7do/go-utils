package entgo

import (
	"entgo.io/ent/dialect/sql"

	paging "github.com/tx7do/go-utils/pagination"
)

func BuildPaginationSelector(page, pageSize int32, noPaging bool) func(*sql.Selector) {
	if noPaging {
		return nil
	} else {
		return func(s *sql.Selector) {
			BuildPaginationSelect(s, page, pageSize)
		}
	}
}

func BuildPaginationSelect(s *sql.Selector, page, pageSize int32) {
	offset := paging.GetPageOffset(page, pageSize)
	s.Offset(offset).Limit(int(pageSize))
}
