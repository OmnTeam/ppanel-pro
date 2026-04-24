package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTask holds the schema definition for the ProxyTask entity.
type ProxyTask struct {
	ent.Schema
}

func (ProxyTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task"},
		entsql.WithComments(true),
	}
}

func (ProxyTask) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("ID"),
		field.Int8("type").Comment("任务类型"),
		field.Text("scope").Optional().Comment("任务范围"),
		field.Text("content").Optional().Comment("任务内容"),
		field.Int8("status").Default(0).Comment("任务状态"),
		field.Text("errors").Optional().Comment("任务错误信息"),
		field.Uint32("total").Default(0).Comment("总数"),
		field.Uint32("current").Default(0).Comment("当前数量"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyTask) Edges() []ent.Edge {
	return nil
}
