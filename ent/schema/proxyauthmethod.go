package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProxyAuthMethod holds the schema definition for the ProxyAuthMethod entity
// 认证方法配置表
type ProxyAuthMethod struct {
	ent.Schema
}

// Annotations of the ProxyAuthMethod
func (ProxyAuthMethod) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_auth_method"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyAuthMethod
func (ProxyAuthMethod) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Comment("认证方法ID"),
		field.String("method").
			MaxLen(255).
			NotEmpty().
			Comment("认证方法"),
		field.Text("config").
			NotEmpty().
			Comment("OAuth配置"),
		field.Bool("enabled").
			Default(false).
			Comment("是否启用"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Edges of the ProxyAuthMethod
func (ProxyAuthMethod) Edges() []ent.Edge {
	return nil
}

// Indexes of the ProxyAuthMethod
func (ProxyAuthMethod) Indexes() []ent.Index {
	return []ent.Index{
		// 认证方法唯一索引
		index.Fields("method").Unique(),
	}
}