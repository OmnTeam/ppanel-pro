package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyServer holds the schema definition for the ProxyServer entity
// 服务器管理表 - 对应原项目的 servers 表和 Server 结构体
type ProxyServer struct {
	ent.Schema
}

// Annotations of the ProxyServer
func (ProxyServer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_server"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyServer
func (ProxyServer) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("服务器ID"),
		field.Int64("tenant_id").
			Default(0).
			Comment("租户ID"),
		field.String("name").
			MaxLen(100).
			Default("").
			Comment("服务器名称"),
		field.String("country").
			MaxLen(128).
			Default("").
			Comment("国家"),
		field.String("city").
			MaxLen(128).
			Default("").
			Comment("城市"),
		field.String("server_addr").
			MaxLen(100).
			Default("").
			Comment("服务器地址"),
		field.Int("sort").
			Default(0).
			Comment("排序"),
		field.Text("protocol").
			Optional().
			Comment("协议配置JSON"),
		field.Time("last_reported_at").
			Optional().
			Nillable().
			Comment("最后报告时间"),
		field.Time("created_at").
			Default(time.Now).
			Optional().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Optional().
			Comment("更新时间"),
	}
}

// Edges of the ProxyServer
func (ProxyServer) Edges() []ent.Edge {
	return nil
}
