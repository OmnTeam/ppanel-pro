package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyServerGroup holds the schema definition for the ProxyServerGroup entity
// 服务器组管理表
type ProxyServerGroup struct {
	ent.Schema
}

// Annotations of the ProxyServerGroup
func (ProxyServerGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_server_group"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyServerGroup
func (ProxyServerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Comment("ID"),
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Comment("Group Name"),
		field.String("description").
			MaxLen(255).
			Default("").
			Comment("Group Description"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation Time"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Update Time"),
	}
}

// Edges of the ProxyServerGroup
func (ProxyServerGroup) Edges() []ent.Edge {
	return nil
}