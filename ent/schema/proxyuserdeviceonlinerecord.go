package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserDeviceOnlineRecord holds the schema definition for the ProxyUserDeviceOnlineRecord entity
// 用户设备在线记录表
type ProxyUserDeviceOnlineRecord struct {
	ent.Schema
}

// Annotations of the ProxyUserDeviceOnlineRecord
func (ProxyUserDeviceOnlineRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_user_device_online_record"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyUserDeviceOnlineRecord
func (ProxyUserDeviceOnlineRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Comment("ID"),
		field.Int("user_id").
			Comment("用户ID"),
		field.String("identifier").
			MaxLen(255).
			NotEmpty().
			Comment("设备标识符"),
		field.Time("online_time").
			Optional().
			Nillable().
			Comment("上线时间"),
		field.Time("offline_time").
			Optional().
			Nillable().
			Comment("下线时间"),
		field.Int("online_seconds").
			Optional().
			Nillable().
			Comment("在线秒数"),
		field.Int("duration_days").
			Optional().
			Nillable().
			Comment("持续天数"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

// Edges of the ProxyUserDeviceOnlineRecord
func (ProxyUserDeviceOnlineRecord) Edges() []ent.Edge {
	return nil
}