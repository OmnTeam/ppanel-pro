package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyAnnouncement holds the schema definition for the ProxyAnnouncement entity
type ProxyAnnouncement struct {
	ent.Schema
}

// Annotations of the ProxyAnnouncement
func (ProxyAnnouncement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_announcement"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyAnnouncement
func (ProxyAnnouncement) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("公告ID"),
		field.String("title").
			NotEmpty().
			Default("").
			MaxLen(255).
			Comment("公告标题"),
		field.Text("content").
			Optional().
			Comment("公告内容"),
		field.Bool("show").
			Default(false).
			Comment("是否显示"),
		field.Bool("pinned").
			Default(false).
			Comment("是否置顶"),
		field.Bool("popup").
			Default(false).
			Comment("是否弹窗"),
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

// Edges of the ProxyAnnouncement
func (ProxyAnnouncement) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the ProxyAnnouncement
func (ProxyAnnouncement) Indexes() []ent.Index {
	return []ent.Index{}
}
