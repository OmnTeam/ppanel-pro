package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTicket holds the schema definition for the ProxyTicket entity.
type ProxyTicket struct {
	ent.Schema
}

func (ProxyTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ticket"},
		entsql.WithComments(true),
	}
}

func (ProxyTicket) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("title").MaxLen(255).Default("").Comment("工单标题"),
		field.Text("description").Optional().Comment("工单描述"),
		field.Int64("user_id").Default(0).Comment("用户ID"),
		field.Int8("status").Default(1).Comment("工单状态"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyTicket) Edges() []ent.Edge {
	return nil
}
