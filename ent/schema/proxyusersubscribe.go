package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserSubscribe holds the schema definition for the ProxyUserSubscribe entity.
type ProxyUserSubscribe struct {
	ent.Schema
}

func (ProxyUserSubscribe) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_subscribe"},
		entsql.WithComments(true),
	}
}

func (ProxyUserSubscribe) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.Int64("order_id").Comment("订单ID"),
		field.Int64("subscribe_id").Comment("订阅套餐ID"),
		field.Int64("node_group_id").Default(0).Comment("节点组ID"),
		field.Bool("group_locked").Default(false).Comment("分组是否锁定"),
		field.Time("start_time").Comment("订阅开始时间"),
		field.Time("expire_time").Optional().Nillable().Comment("订阅过期时间"),
		field.Time("finished_at").Optional().Nillable().Comment("订阅完成时间"),
		field.Int64("traffic").Optional().Nillable().Default(0).Comment("总流量"),
		field.Int64("download").Optional().Nillable().Default(0).Comment("下载流量"),
		field.Int64("upload").Optional().Nillable().Default(0).Comment("上传流量"),
		field.Int64("expired_download").Optional().Nillable().Default(0).Comment("过期下载流量"),
		field.Int64("expired_upload").Optional().Nillable().Default(0).Comment("过期上传流量"),
		field.String("token").MaxLen(255).Optional().Nillable().Unique().Comment("订阅令牌"),
		field.String("uuid").MaxLen(255).Optional().Nillable().Unique().Comment("订阅UUID"),
		field.Int8("status").Optional().Nillable().Default(0).Comment("订阅状态"),
		field.String("note").MaxLen(500).Optional().Nillable().Comment("用户备注"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyUserSubscribe) Edges() []ent.Edge {
	return nil
}
