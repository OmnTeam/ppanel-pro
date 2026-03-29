package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyGroupHistory holds the schema definition for the ProxyGroupHistory entity.
type ProxyGroupHistory struct {
	ent.Schema
}

func (ProxyGroupHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_history"},
		entsql.WithComments(true),
	}
}

func (ProxyGroupHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("历史记录ID"),
		field.String("group_mode").MaxLen(50).Default("").Comment("分组模式"),
		field.String("trigger_type").MaxLen(50).Default("").Comment("触发类型"),
		field.String("state").MaxLen(50).Default("").Comment("状态"),
		field.Int("total_users").Default(0).Comment("总用户数"),
		field.Int("success_count").Default(0).Comment("成功数量"),
		field.Int("failed_count").Default(0).Comment("失败数量"),
		field.Time("start_time").Optional().Nillable().Comment("开始时间"),
		field.Time("end_time").Optional().Nillable().Comment("结束时间"),
		field.String("operator").MaxLen(100).Optional().Nillable().Comment("操作人"),
		field.Text("error_message").Optional().Comment("错误信息"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
	}
}

func (ProxyGroupHistory) Edges() []ent.Edge {
	return nil
}
