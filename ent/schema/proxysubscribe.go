package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySubscribe holds the schema definition for the ProxySubscribe entity
// 订阅管理表
type ProxySubscribe struct {
	ent.Schema
}

// Annotations of the ProxySubscribe
func (ProxySubscribe) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_subscribe"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxySubscribe
func (ProxySubscribe) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Comment("订阅套餐ID"),
		field.String("name").
			MaxLen(255).
			Default("").
			Comment("订阅套餐名称"),
		field.String("language").
			MaxLen(255).
			Default("").
			Comment("语言"),
		field.Text("description").
			Optional().
			Nillable().
			Comment("订阅套餐描述"),
		field.Int64("unit_price").
			Default(0).
			Comment("单位价格（单位：分）"),
		field.String("unit_time").
			MaxLen(255).
			Default("").
			Comment("单位时间"),
		field.Text("discount").
			Optional().
			Nillable().
			Comment("折扣配置JSON"),
		field.Int64("replacement").
			Default(0).
			Comment("替换"),
		field.Int64("inventory").
			Default(-1).
			Comment("库存"),
		field.Int64("traffic").
			Default(0).
			Comment("流量（字节）"),
		field.Int64("speed_limit").
			Default(0).
			Comment("速度限制"),
		field.Int64("device_limit").
			Default(0).
			Comment("设备数限制"),
		field.Int64("quota").
			Default(0).
			Comment("配额"),
		field.Bool("show").
			Default(false).
			Comment("是否在门户页面显示"),
		field.Bool("sell").
			Default(false).
			Comment("是否售卖"),
		field.Int64("sort").
			Default(0).
			Comment("排序"),
		field.Int64("deduction_ratio").
			Optional().
			Nillable().
			Default(0).
			Comment("扣除比例"),
		field.Bool("allow_deduction").
			Optional().
			Default(true).
			Comment("允许扣除"),
		field.Int64("reset_cycle").
			Optional().
			Nillable().
			Default(0).
			Comment("重置周期: 0-不重置 1-1号 2-每月 3-每年"),
		field.Bool("renewal_reset").
			Optional().
			Default(false).
			Comment("续费重置"),
		field.String("nodes").
			MaxLen(255).
			Default("").
			Comment("节点IDs"),
		field.String("node_tags").
			MaxLen(255).
			Default("").
			Comment("节点标签"),
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

// Edges of the ProxySubscribe
func (ProxySubscribe) Edges() []ent.Edge {
	return nil
}
