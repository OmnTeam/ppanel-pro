package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySystemLog holds the schema definition for the ProxySystemLog entity
// 系统日志表
type ProxySystemLog struct {
	ent.Schema
}

// Annotations of the ProxySystemLog
func (ProxySystemLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_system_log"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxySystemLog
func (ProxySystemLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.Int8("type").Comment("日志类型: 10=Email, 11=Mobile, 20=Subscribe, 21=SubscribeTraffic, 22=ServerTraffic, 23=ResetSubscribe, 30=Login, 31=Register, 32=Balance, 33=Commission, 34=Gift"),
		field.String("date").Optional().MaxLen(20).Comment("日志日期"),
		field.Int64("object_id").Default(0).Comment("对象ID（用户ID或其他关联对象ID）"),
		field.Text("content").Comment("日志内容（JSON格式）"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
	}
}

// Edges of the ProxySystemLog
func (ProxySystemLog) Edges() []ent.Edge {
	return nil
}
