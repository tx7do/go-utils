package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*Tag)(nil)

// Tag holds the schema definition for the tags
type Tag struct {
	mixin.Schema
}

// Fields of the Tag.
func (t Tag) Fields() []ent.Field {
	return []ent.Field{
		field.Strings("tags").
			Comment("tags associated with the object").
			Default([]string{}).
			Optional(),
	}
}
