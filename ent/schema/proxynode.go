package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyNode holds the schema definition for the ProxyNode entity
// 代理节点管理表 - 对应原项目的 nodes 表
type ProxyNode struct {
	ent.Schema
}

// Annotations of the ProxyNode
func (ProxyNode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_nodes"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyNode
func (ProxyNode) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("节点ID"),
		field.String("name").
			MaxLen(100).
			Default("").
			Comment("节点名称"),
		field.String("tags").
			MaxLen(255).
			Default("").
			Comment("标签"),
		field.Int("port").
			Default(0).
			Comment("连接端口"),
		field.String("address").
			MaxLen(255).
			Default("").
			Comment("连接地址"),
		field.Int64("server_id").
			Default(0).
			Comment("服务器ID"),
		field.String("protocol").
			MaxLen(100).
			Default("").
			Comment("协议"),
		field.Bool("enabled").
			Default(true).
			Comment("启用"),
		field.Int("sort").
			Default(0).
			Comment("排序"),
		field.Int64("group_id").
			Optional().
			Nillable().
			Default(0).
			Comment("节点分组ID"),
		field.Bool("group_locked").
			Default(false).
			Comment("是否锁定分组"),
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

// Edges of the ProxyNode
func (ProxyNode) Edges() []ent.Edge {
	return nil
}
