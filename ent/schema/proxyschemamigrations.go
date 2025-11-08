package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySchemaMigrations holds the schema definition for the ProxySchemaMigrations entity
// 数据库迁移版本记录表
type ProxySchemaMigrations struct {
	ent.Schema
}

// Annotations of the ProxySchemaMigrations
func (ProxySchemaMigrations) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_schema_migrations"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxySchemaMigrations
func (ProxySchemaMigrations) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").
			Unique().
			Comment("迁移版本号"),
		field.Bool("dirty").
			Default(false).
			Comment("是否脏数据"),
	}
}

// Edges of the ProxySchemaMigrations
func (ProxySchemaMigrations) Edges() []ent.Edge {
	return nil
}