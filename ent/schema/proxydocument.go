package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProxyDocument holds the schema definition for the ProxyDocument entity.
type ProxyDocument struct {
	ent.Schema
}

func (ProxyDocument) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "document"},
		entsql.WithComments(true),
	}
}

func (ProxyDocument) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Comment("文档ID"),
		field.String("title").MaxLen(255).NotEmpty().Default("").Comment("文档标题"),
		field.Text("content").Optional().Comment("文档内容"),
		field.String("tags").MaxLen(255).NotEmpty().Default("").Comment("文档标签"),
		field.Bool("show").Default(true).Comment("显示"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyDocument) Edges() []ent.Edge {
	return nil
}

func (ProxyDocument) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("show"),
	}
}
