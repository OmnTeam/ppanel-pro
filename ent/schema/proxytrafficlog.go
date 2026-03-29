package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTrafficLog holds the schema definition for the ProxyTrafficLog entity.
type ProxyTrafficLog struct {
	ent.Schema
}

func (ProxyTrafficLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "traffic_log"},
		entsql.WithComments(true),
	}
}

func (ProxyTrafficLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("ID"),
		field.Int64("server_id").Comment("服务器ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.Int64("subscribe_id").Comment("订阅ID"),
		field.Int64("download").Default(0).Comment("下载流量"),
		field.Int64("upload").Default(0).Comment("上传流量"),
		field.Time("timestamp").Default(time.Now).Comment("流量日志时间"),
	}
}

func (ProxyTrafficLog) Edges() []ent.Edge {
	return nil
}
