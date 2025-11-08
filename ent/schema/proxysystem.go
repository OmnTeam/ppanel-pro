package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProxySystem holds the schema definition for the ProxySystem entity
// 系统配置表
type ProxySystem struct {
	ent.Schema
}

// Annotations of the ProxySystem
func (ProxySystem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_system"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxySystem
func (ProxySystem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("category").MaxLen(100).Default("").Comment("分类"),
		field.String("key").MaxLen(100).NotEmpty().Comment("键名"),
		field.Text("value").Comment("键值"),
		field.String("type").MaxLen(50).Default("").Comment("类型"),
		field.Text("desc").Comment("描述"),
		field.Time("created_at").Optional().Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Optional().Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

// Indexes of the ProxySystem
func (ProxySystem) Indexes() []ent.Index {
	return []ent.Index{
		// 确保category+key组合是唯一的
		index.Fields("category", "key").Unique(),
	}
}

// Edges of the ProxySystem
func (ProxySystem) Edges() []ent.Edge {
	return nil
}
