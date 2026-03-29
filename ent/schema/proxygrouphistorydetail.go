package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyGroupHistoryDetail holds the schema definition for the ProxyGroupHistoryDetail entity.
type ProxyGroupHistoryDetail struct {
	ent.Schema
}

func (ProxyGroupHistoryDetail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_history_detail"},
		entsql.WithComments(true),
	}
}

func (ProxyGroupHistoryDetail) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("详情记录ID"),
		field.Int64("history_id").Comment("历史记录ID"),
		field.Int64("node_group_id").Comment("节点组ID"),
		field.Int("user_count").Default(0).Comment("用户数量"),
		field.Int("node_count").Default(0).Comment("节点数量"),
		field.Text("user_data").Optional().Comment("用户数据"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
	}
}

func (ProxyGroupHistoryDetail) Edges() []ent.Edge {
	return nil
}
