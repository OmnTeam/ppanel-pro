package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTicketFollow holds the schema definition for the ProxyTicketFollow entity.
type ProxyTicketFollow struct {
	ent.Schema
}

func (ProxyTicketFollow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ticket_follow"},
		entsql.WithComments(true),
	}
}

func (ProxyTicketFollow) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.Int64("ticket_id").Default(0).Comment("工单ID"),
		field.String("from").MaxLen(255).Default("").Comment("来源/操作人"),
		field.Int8("type").Default(1).Comment("类型"),
		field.Text("content").Optional().Comment("跟进内容"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
	}
}

func (ProxyTicketFollow) Edges() []ent.Edge {
	return nil
}
