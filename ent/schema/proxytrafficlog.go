package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTrafficLog holds the schema definition for the ProxyTrafficLog entity
// 流量日志表
type ProxyTrafficLog struct {
	ent.Schema
}

// Annotations of the ProxyTrafficLog
func (ProxyTrafficLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_traffic_log"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyTrafficLog
func (ProxyTrafficLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("ID"),
		field.Int64("server_id").Comment("服务器ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.Int64("subscribe_id").Comment("订阅ID"),
		field.Int("download").Default(0).Comment("下载流量"),
		field.Int("upload").Default(0).Comment("上传流量"),
		field.Time("timestamp").Default(time.Now).Comment("流量日志时间"),
	}
}

// Edges of the ProxyTrafficLog
func (ProxyTrafficLog) Edges() []ent.Edge {
	return nil
}
