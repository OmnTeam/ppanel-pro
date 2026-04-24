package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ProxyRedemptionCode holds the schema definition for the ProxyRedemptionCode entity.
type ProxyRedemptionCode struct {
	ent.Schema
}

func (ProxyRedemptionCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "redemption_code"},
		entsql.WithComments(true),
	}
}

func (ProxyRedemptionCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("兑换码ID"),
		field.String("code").Unique().MaxLen(255).NotEmpty().Comment("兑换码"),
		field.Int32("total_count").Default(0).Comment("总兑换次数"),
		field.Int32("used_count").Default(0).Comment("已使用次数"),
		field.Int64("subscribe_plan").Default(0).Comment("订阅套餐ID"),
		field.String("unit_time").Default("month").MaxLen(50).Comment("时间单位"),
		field.Int32("quantity").Default(1).Comment("数量"),
		field.Int8("status").Default(1).Comment("状态"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
		field.Time("deleted_at").Optional().Nillable().Comment("删除时间"),
	}
}

func (ProxyRedemptionCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("records", ProxyRedemptionRecord.Type),
	}
}
