package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProxyPayment holds the schema definition for the ProxyPayment entity
// 支付方式管理表
type ProxyPayment struct {
	ent.Schema
}

// Annotations of the ProxyPayment
func (ProxyPayment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_payment"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyPayment
func (ProxyPayment) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Comment("支付ID"),
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Default("").
			Comment("支付名称"),
		field.String("platform").
			MaxLen(100).
			NotEmpty().
			Comment("支付平台"),
		field.Text("description").
			Optional().
			Comment("支付描述"),
		field.String("icon").
			MaxLen(255).
			Optional().
			Default("").
			Comment("支付图标"),
		field.String("domain").
			MaxLen(255).
			Optional().
			Default("").
			Comment("通知域名"),
		field.Text("config").
			NotEmpty().
			Comment("支付配置"),
		field.Int("fee_mode").
			Default(0).
			Comment("费用模式：0：无费用 1：百分比 2：固定金额 3：百分比+固定金额"),
		field.Float("fee_percent").
			Optional().
			Default(0).
			Comment("费用百分比"),
		field.Int64("fee_amount").
			Optional().
			Default(0).
			Comment("固定费用金额"),
		field.Bool("enable").
			Default(false).
			Comment("是否启用"),
		field.String("token").
			MaxLen(255).
			Optional().
			Unique().
			Comment("支付令牌"),
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

// Edges of the ProxyPayment
func (ProxyPayment) Edges() []ent.Edge {
	return nil
}

// Indexes of the ProxyPayment
func (ProxyPayment) Indexes() []ent.Index {
	return []ent.Index{
		// 平台索引
		index.Fields("platform"),
		// 启用状态索引
		index.Fields("enable"),
		// 平台和启用状态联合索引
		index.Fields("platform", "enable"),
	}
}