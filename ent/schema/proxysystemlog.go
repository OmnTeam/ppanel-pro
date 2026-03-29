package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySystemLog holds the schema definition for the ProxySystemLog entity.
type ProxySystemLog struct {
	ent.Schema
}

func (ProxySystemLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "system_logs"},
		entsql.WithComments(true),
	}
}

func (ProxySystemLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.Int8("type").Comment("日志类型"),
		field.String("date").MaxLen(20).Optional().Comment("日志日期"),
		field.Int64("object_id").Default(0).Comment("对象ID"),
		field.Text("content").Comment("日志内容"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
	}
}

func (ProxySystemLog) Edges() []ent.Edge {
	return nil
}
