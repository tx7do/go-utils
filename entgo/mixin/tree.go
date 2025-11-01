package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*ParentID)(nil)

type ParentID struct {
	mixin.Schema
}

func (t ParentID) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("parent_id").
			Comment("父节点ID").
			Optional().
			Nillable(),
	}
}

type TableInterface interface {
	Type()
}

type Tree[T TableInterface] struct {
	mixin.Schema
}

func (Tree[T]) Fields() []ent.Field {
	var fields []ent.Field
	fields = append(fields, ParentID{}.Fields()...)
	return fields
}

// Edges of the Tree.
func (Tree[T]) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			To("children", T.Type).
			From("parent").Unique().Field("parent_id"),
	}
}
