package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySubscribeGroup holds the schema definition for the ProxySubscribeGroup entity.
type ProxySubscribeGroup struct {
	ent.Schema
}

func (ProxySubscribeGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscribe_group"},
		entsql.WithComments(true),
	}
}

func (ProxySubscribeGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("订阅组ID"),
		field.String("name").MaxLen(255).Default("").Comment("订阅组名称"),
		field.Text("description").Optional().Nillable().Comment("订阅组描述"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxySubscribeGroup) Edges() []ent.Edge {
	return nil
}
