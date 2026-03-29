package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyAds holds the schema definition for the ProxyAds entity.
type ProxyAds struct {
	ent.Schema
}

func (ProxyAds) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ads"},
		entsql.WithComments(true),
	}
}

func (ProxyAds) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Comment("广告ID"),
		field.String("title").MaxLen(255).NotEmpty().Default("").Comment("广告标题"),
		field.String("type").MaxLen(255).NotEmpty().Default("").Comment("广告类型"),
		field.Text("content").Optional().Comment("广告内容"),
		field.Text("description").Optional().Comment("广告描述"),
		field.String("target_url").MaxLen(512).Default("").Comment("广告目标链接"),
		field.Time("start_time").Comment("广告开始时间"),
		field.Time("end_time").Comment("广告结束时间"),
		field.Int("status").Default(0).Comment("广告状态，0禁用，1启用"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyAds) Edges() []ent.Edge {
	return nil
}
