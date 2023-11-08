package entgo

import (
	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/go-utils/stringcase"
)

func BuildSetNullUpdate(u *sql.UpdateBuilder, fields []string) {
	if len(fields) > 0 {
		for _, field := range fields {
			field = stringcase.ToSnakeCase(field)
			u.SetNull(field)
		}
	}
}

func BuildSetNullUpdater(fields []string) (error, func(u *sql.UpdateBuilder)) {
	if len(fields) > 0 {
		return nil, func(u *sql.UpdateBuilder) {
			BuildSetNullUpdate(u, fields)
		}
	} else {
		return nil, nil
	}
}
