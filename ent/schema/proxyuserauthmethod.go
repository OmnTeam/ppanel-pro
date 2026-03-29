package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserAuthMethod holds the schema definition for the ProxyUserAuthMethod entity.
type ProxyUserAuthMethod struct {
	ent.Schema
}

func (ProxyUserAuthMethod) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_auth_methods"},
		entsql.WithComments(true),
	}
}

func (ProxyUserAuthMethod) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.String("auth_type").MaxLen(255).NotEmpty().Comment("认证类型"),
		field.String("auth_identifier").MaxLen(255).NotEmpty().Unique().Comment("认证标识"),
		field.Bool("verified").Default(false).Comment("是否已验证"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyUserAuthMethod) Edges() []ent.Edge {
	return nil
}
