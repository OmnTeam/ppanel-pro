package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyCoupon holds the schema definition for the ProxyCoupon entity.
type ProxyCoupon struct {
	ent.Schema
}

func (ProxyCoupon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "coupon"},
		entsql.WithComments(true),
	}
}

func (ProxyCoupon) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Comment("优惠券ID"),
		field.String("name").MaxLen(255).NotEmpty().Default("").Comment("优惠券名称"),
		field.String("code").MaxLen(255).NotEmpty().Unique().Comment("优惠券代码"),
		field.Int32("count").Default(0).Comment("数量限制"),
		field.Int8("type").Default(1).Comment("优惠券类型：1：百分比 2：固定金额"),
		field.Int64("discount").Default(0).Comment("优惠券折扣"),
		field.Int64("start_time").Default(0).Comment("开始时间"),
		field.Int64("expire_time").Default(0).Comment("结束时间"),
		field.Int64("user_limit").Default(0).Comment("用户限制"),
		field.String("subscribe").MaxLen(255).Default("").Comment("订阅限制"),
		field.Int8("used_count").Default(0).Comment("已使用次数"),
		field.Bool("enable").Default(true).Comment("是否启用"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyCoupon) Edges() []ent.Edge {
	return nil
}
