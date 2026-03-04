package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserGroup 用户组表
type ProxyUserGroup struct {
	ent.Schema
}

func (ProxyUserGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_user_group"},
		entsql.WithComments(true),
	}
}

func (ProxyUserGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("用户组ID"),
		field.String("name").MaxLen(255).NotEmpty().Comment("用户组名称"),
		field.String("description").MaxLen(500).Default("").Comment("用户组描述"),
		field.Int("sort").Default(0).Comment("排序"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyUserGroup) Edges() []ent.Edge {
	return nil
}
