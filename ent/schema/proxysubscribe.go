package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySubscribe holds the schema definition for the ProxySubscribe entity.
type ProxySubscribe struct {
	ent.Schema
}

func (ProxySubscribe) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscribe"},
		entsql.WithComments(true),
	}
}

func (ProxySubscribe) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("订阅套餐ID"),
		field.String("name").MaxLen(255).Default("").Comment("订阅套餐名称"),
		field.String("language").MaxLen(255).Default("").Comment("语言"),
		field.Text("description").Optional().Nillable().Comment("订阅套餐描述"),
		field.Int64("unit_price").Default(0).Comment("单位价格"),
		field.String("unit_time").MaxLen(255).Default("").Comment("单位时间"),
		field.Text("discount").Optional().Nillable().Comment("折扣配置"),
		field.Int64("replacement").Default(0).Comment("替换"),
		field.Int32("inventory").Default(-1).Comment("库存"),
		field.Int64("traffic").Default(0).Comment("流量"),
		field.Int32("speed_limit").Default(0).Comment("速度限制"),
		field.Int32("device_limit").Default(0).Comment("设备数限制"),
		field.Int32("quota").Default(0).Comment("配额"),
		field.String("nodes").MaxLen(255).Default("").Comment("节点IDs"),
		field.String("node_tags").MaxLen(255).Default("").Comment("节点标签"),
		field.JSON("node_group_ids", []int64{}).Optional().Comment("节点组ID列表"),
		field.Int64("node_group_id").Optional().Nillable().Default(0).Comment("默认节点组ID"),
		field.Text("traffic_limit").Optional().Nillable().Comment("流量限制规则"),
		field.Bool("show").Default(false).Comment("是否显示"),
		field.Bool("sell").Default(false).Comment("是否售卖"),
		field.Int32("sort").Default(0).Comment("排序"),
		field.Int32("deduction_ratio").Optional().Nillable().Default(0).Comment("扣除比例"),
		field.Bool("allow_deduction").Default(true).Comment("允许扣除"),
		field.Int32("reset_cycle").Optional().Nillable().Default(0).Comment("重置周期"),
		field.Bool("renewal_reset").Default(false).Comment("续费重置"),
		field.Bool("show_original_price").Default(true).Comment("显示原价"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxySubscribe) Edges() []ent.Edge {
	return nil
}
