package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserDeviceOnlineRecord holds the schema definition for the ProxyUserDeviceOnlineRecord entity.
type ProxyUserDeviceOnlineRecord struct {
	ent.Schema
}

func (ProxyUserDeviceOnlineRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_device_online_record"},
		entsql.WithComments(true),
	}
}

func (ProxyUserDeviceOnlineRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.String("identifier").MaxLen(255).NotEmpty().Comment("设备标识符"),
		field.Time("online_time").Optional().Nillable().Comment("上线时间"),
		field.Time("offline_time").Optional().Nillable().Comment("下线时间"),
		field.Int64("online_seconds").Optional().Nillable().Comment("在线秒数"),
		field.Int64("duration_days").Optional().Nillable().Comment("持续天数"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
	}
}

func (ProxyUserDeviceOnlineRecord) Edges() []ent.Edge {
	return nil
}
