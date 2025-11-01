package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// 注意：使用此 Mixin 时，需要在 Ent 客户端配置 Hook 或 Interceptor
// 来在更新操作前自动检查并递增版本号。

var _ ent.Mixin = (*Version)(nil)

type Version struct{ mixin.Schema }

func (Version) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("version").
			Comment("版本号/乐观锁").
			Default(1), // 初始版本为 1
	}
}
