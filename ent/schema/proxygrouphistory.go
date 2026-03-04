package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyGroupHistory 分组历史记录表
type ProxyGroupHistory struct {
	ent.Schema
}

func (ProxyGroupHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_group_history"},
		entsql.WithComments(true),
	}
}

func (ProxyGroupHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("历史记录ID"),
		field.String("group_mode").MaxLen(50).Default("user").Comment("分组模式: user/node"),
		field.String("trigger_type").MaxLen(50).Default("manual").Comment("触发类型: manual/auto"),
		field.String("status").MaxLen(50).Default("pending").Comment("状态: pending/running/completed/failed"),
		field.Int("progress").Default(0).Comment("进度"),
		field.Int("total").Default(0).Comment("总数"),
		field.Text("result").Optional().Comment("结果详情JSON"),
		field.Text("error").Optional().Comment("错误信息"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyGroupHistory) Edges() []ent.Edge {
	return nil
}
