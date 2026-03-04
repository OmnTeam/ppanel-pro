package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyAds holds the schema definition for the ProxyAds entity
// 广告管理表
type ProxyAds struct {
	ent.Schema
}

// Annotations of the ProxyAds
func (ProxyAds) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_ads"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyAds
func (ProxyAds) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("广告ID"),
		field.Int64("tenant_id").
			Default(0).
			Comment("租户ID"),
		field.String("title").
			NotEmpty().
			Default("").
			Comment("广告标题"),
		field.String("type").
			NotEmpty().
			Default("").
			Comment("广告类型"),
		field.Text("content").
			Optional().
			Comment("广告内容"),
		field.Text("description").
			Optional().
			Comment("广告描述"),
		field.String("target_url").
			MaxLen(512).
			Default("").
			Comment("广告目标链接"),
		field.Time("start_time").
			Optional().
			Comment("广告开始时间"),
		field.Time("end_time").
			Optional().
			Comment("广告结束时间"),
		field.Int8("status").
			Optional().
			Default(0).
			Comment("广告状态，0禁用，1启用"),
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

// Edges of the ProxyAds
func (ProxyAds) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the ProxyAds
func (ProxyAds) Indexes() []ent.Index {
	return []ent.Index{}
}
