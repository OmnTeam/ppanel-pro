package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProxyAuthMethod holds the schema definition for the ProxyAuthMethod entity.
type ProxyAuthMethod struct {
	ent.Schema
}

func (ProxyAuthMethod) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "auth_method"},
		entsql.WithComments(true),
	}
}

func (ProxyAuthMethod) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Comment("认证方法ID"),
		field.String("method").MaxLen(255).NotEmpty().Default("").Comment("认证方法"),
		field.Text("config").NotEmpty().Comment("OAuth配置"),
		field.Bool("enabled").Default(false).Comment("是否启用"),
		field.Time("created_at").Optional().Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Optional().Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyAuthMethod) Edges() []ent.Edge {
	return nil
}

func (ProxyAuthMethod) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("method").Unique(),
	}
}
