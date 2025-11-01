package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"
)

type SoftDelete struct {
	mixin.Schema
}

func (SoftDelete) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DeletedAt{},
		DeletedBy{},
	}
}
