package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserAuthMethod holds the schema definition for the ProxyUserAuthMethod entity
// 用户认证方法关系表
type ProxyUserAuthMethod struct {
	ent.Schema
}

// Annotations of the ProxyUserAuthMethod
func (ProxyUserAuthMethod) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_user_auth_method"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyUserAuthMethod
func (ProxyUserAuthMethod) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Comment("ID"),
		field.Int("user_id").
			Comment("用户ID"),
		field.Int64("tenant_id").
			Default(0).
			Comment("租户ID"),
		field.String("auth_type").
			MaxLen(255).
			NotEmpty().
			Comment("认证类型: apple, google, github, facebook, telegram, email, mobile"),
		field.String("auth_identifier").
			MaxLen(255).
			NotEmpty().
			Comment("认证标识"),
		field.Bool("verified").
			Default(false).
			Comment("是否已验证"),
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

// Edges of the ProxyUserAuthMethod
func (ProxyUserAuthMethod) Edges() []ent.Edge {
	return nil
}