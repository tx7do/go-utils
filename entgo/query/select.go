package entgo

import (
	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/go-utils/stringcase"
)

func BuildFieldSelect(s *sql.Selector, fields []string) {
	if len(fields) > 0 {
		for i, field := range fields {
			fields[i] = stringcase.ToSnakeCase(field)
		}
		s.Select(fields...)
	}
}

func BuildFieldSelector(fields []string) (error, func(s *sql.Selector)) {
	if len(fields) > 0 {
		return nil, func(s *sql.Selector) {
			BuildFieldSelect(s, fields)
		}
	} else {
		return nil, nil
	}
}
