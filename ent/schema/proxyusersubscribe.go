package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserSubscribe holds the schema definition for the ProxyUserSubscribe entity
// 用户订阅关系表
type ProxyUserSubscribe struct {
	ent.Schema
}

// Annotations of the ProxyUserSubscribe
func (ProxyUserSubscribe) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_user_subscribe"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyUserSubscribe
func (ProxyUserSubscribe) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Comment("ID"),
		field.Int64("user_id").
			Comment("用户ID"),
		field.Int64("order_id").
			Comment("订单ID"),
		field.Int64("subscribe_id").
			Comment("订阅套餐ID"),
		field.Time("start_time").
			Comment("订阅开始时间"),
		field.Time("expire_time").
			Optional().
			Nillable().
			Comment("订阅过期时间"),
		field.Time("finished_at").
			Optional().
			Nillable().
			Comment("订阅完成时间"),
		field.Int64("traffic").
			Optional().
			Nillable().
			Comment("总流量（字节）"),
		field.Int64("download").
			Optional().
			Nillable().
			Comment("下载流量（字节）"),
		field.Int64("upload").
			Optional().
			Nillable().
			Comment("上传流量（字节）"),
		field.String("token").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("订阅令牌"),
		field.String("uuid").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("订阅UUID"),
		field.Int8("status").
			Optional().
			Nillable().
			Comment("订阅状态: 0-待激活 1-激活 2-完成 3-过期 4-已扣除"),
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

// Edges of the ProxyUserSubscribe
func (ProxyUserSubscribe) Edges() []ent.Edge {
	return nil
}
