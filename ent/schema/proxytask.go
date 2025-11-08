package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyTask holds the schema definition for the ProxyTask entity
// 任务管理表
type ProxyTask struct {
	ent.Schema
}

// Annotations of the ProxyTask
func (ProxyTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_task"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyTask
func (ProxyTask) Fields() []ent.Field {
	return []ent.Field{
		// 基础字段
		field.Int64("id").Unique().Comment("ID"),

		// 任务字段 - 严格按照ppanel.sql定义
		field.Int8("type").Comment("任务类型: 0:Email, 1:Quota"),
		field.Text("scope").Optional().Comment("任务范围（JSON格式）"),
		field.Text("content").Optional().Comment("任务内容（JSON格式）"),
		field.Int8("status").Default(0).Comment("任务状态: 0:Pending, 1:In Progress, 2:Completed, 3:Failed"),
		field.Text("errors").Optional().Comment("任务错误信息"),
		field.Uint64("total").Default(0).Comment("总数"),
		field.Uint64("current").Default(0).Comment("当前数量"),

		// 时间戳字段
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

// Edges of the ProxyTask
func (ProxyTask) Edges() []ent.Edge {
	return nil
}
