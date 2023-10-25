package entgo

import (
	"entgo.io/ent/dialect/sql"

	paging "github.com/tx7do/kratos-utils/pagination"
)

func BuildPaginationSelector(page, pageSize int32, noPaging bool) func(*sql.Selector) {
	if noPaging {
		return nil
	}

	if page == 0 {
		page = DefaultPage
	}

	if pageSize == 0 {
		pageSize = DefaultPageSize
	}

	return func(s *sql.Selector) {
		s.Offset(paging.GetPageOffset(page, pageSize)).
			Limit(int(pageSize))
	}
}
