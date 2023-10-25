package entgo

import (
	"entgo.io/ent/dialect/sql"

	paging "restroom-system/pkg/util/pagination"
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
