package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySchemaMigrations holds the schema definition for the ProxySchemaMigrations entity.
type ProxySchemaMigrations struct {
	ent.Schema
}

func (ProxySchemaMigrations) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "schema_migrations"},
		entsql.WithComments(true),
	}
}

func (ProxySchemaMigrations) Fields() []ent.Field {
	incremental := false

	return []ent.Field{
		field.Int64("id").
			StorageKey("version").
			Annotations(entsql.Annotation{Incremental: &incremental}).
			Immutable().
			Comment("迁移版本号"),
		field.Bool("dirty").Comment("是否脏数据"),
	}
}

func (ProxySchemaMigrations) Edges() []ent.Edge {
	return nil
}
