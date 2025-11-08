package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTicketFollow holds the schema definition for the ProxyTicketFollow entity
// 工单追踪记录表
type ProxyTicketFollow struct {
	ent.Schema
}

// Annotations of the ProxyTicketFollow
func (ProxyTicketFollow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_ticket_follow"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyTicketFollow
func (ProxyTicketFollow) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.Int64("ticket_id").Default(0).Comment("工单ID"),
		field.String("from").Default("").Comment("来源/操作人"),
		field.Int8("type").Default(1).Comment("类型: 1=文本, 2=图片"),
		field.Text("content").Optional().Comment("跟进内容"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
	}
}

// Edges of the ProxyTicketFollow
func (ProxyTicketFollow) Edges() []ent.Edge {
	return nil
}
