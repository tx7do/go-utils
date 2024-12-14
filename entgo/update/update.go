package entgo

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"

	"github.com/tx7do/go-utils/fieldmaskutil"
	"github.com/tx7do/go-utils/stringcase"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func BuildSetNullUpdate(u *sql.UpdateBuilder, fields []string) {
	if len(fields) > 0 {
		for _, field := range fields {
			field = stringcase.ToSnakeCase(field)
			u.SetNull(field)
		}
	}
}

// BuildSetNullUpdater 构建一个UpdateBuilder，用于清空字段的值
func BuildSetNullUpdater(fields []string) func(u *sql.UpdateBuilder) {
	if len(fields) == 0 {
		return nil
	}

	return func(u *sql.UpdateBuilder) {
		BuildSetNullUpdate(u, fields)
	}
}

// ExtractJsonFieldKeyValues 提取json字段的键值对
func ExtractJsonFieldKeyValues(msg proto.Message, paths []string, needToSnakeCase bool) []string {
	var keyValues []string
	rft := msg.ProtoReflect()
	for _, path := range paths {
		fd := rft.Descriptor().Fields().ByName(protoreflect.Name(path))
		if fd == nil {
			continue
		}
		if !rft.Has(fd) {
			continue
		}

		var k string
		if needToSnakeCase {
			k = stringcase.ToSnakeCase(path)
		} else {
			k = path
		}

		keyValues = append(keyValues, fmt.Sprintf("'%s'", k))

		v := rft.Get(fd)
		switch v.Interface().(type) {
		case int32, int64, uint32, uint64, float32, float64, bool:
			keyValues = append(keyValues, fmt.Sprintf("%d", v.Interface()))
		case string:
			keyValues = append(keyValues, fmt.Sprintf("'%s'", v.Interface()))
		}
	}

	return keyValues
}

// SetJsonNullFieldUpdateBuilder 设置json字段的空值
func SetJsonNullFieldUpdateBuilder(fieldName string, msg proto.Message, paths []string) func(u *sql.UpdateBuilder) {
	nilPaths := fieldmaskutil.NilValuePaths(msg, paths)
	if len(nilPaths) == 0 {
		return nil
	}

	return func(u *sql.UpdateBuilder) {
		u.Set(fieldName,
			sql.Expr(
				fmt.Sprintf("\"%s\" - '{%s}'::text[]", fieldName, strings.Join(nilPaths, ",")),
			),
		)
	}
}

// SetJsonFieldValueUpdateBuilder 设置json字段的值
func SetJsonFieldValueUpdateBuilder(fieldName string, msg proto.Message, paths []string, needToSnakeCase bool) func(u *sql.UpdateBuilder) {
	keyValues := ExtractJsonFieldKeyValues(msg, paths, needToSnakeCase)
	if len(keyValues) == 0 {
		return nil
	}

	return func(u *sql.UpdateBuilder) {
		u.Set(fieldName,
			sql.Expr(
				fmt.Sprintf("\"%s\" || jsonb_build_object(%s)", fieldName, strings.Join(keyValues, ",")),
			),
		)
	}
}
