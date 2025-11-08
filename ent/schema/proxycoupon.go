package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyCoupon holds the schema definition for the ProxyCoupon entity
// 优惠券管理表
type ProxyCoupon struct {
	ent.Schema
}

// Annotations of the ProxyCoupon
func (ProxyCoupon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_coupon"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyCoupon
func (ProxyCoupon) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("优惠券ID"),
		field.String("name").
			MaxLen(255).
			NotEmpty().
			Comment("优惠券名称"),
		field.String("code").
			MaxLen(255).
			NotEmpty().
			Unique().
			Comment("优惠券代码"),
		field.Int64("count").
			Default(0).
			Comment("数量限制"),
		field.Int8("type").
			Default(1).
			Comment("优惠券类型：1：百分比 2：固定金额"),
		field.Int64("discount").
			Default(0).
			Comment("优惠券折扣"),
		field.Time("start_time").
			Default(time.Now).
			Comment("开始时间"),
		field.Time("end_time").
			Default(time.Now).
			Comment("结束时间"),
		field.Int8("status").
			Default(0).
			Comment("优惠券状态，0禁用，1启用"),
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

// Edges of the ProxyCoupon
func (ProxyCoupon) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the ProxyCoupon
func (ProxyCoupon) Indexes() []ent.Index {
	return []ent.Index{}
}