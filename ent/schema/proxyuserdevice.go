package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUserDevice holds the schema definition for the ProxyUserDevice entity.
type ProxyUserDevice struct {
	ent.Schema
}

func (ProxyUserDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_device"},
		entsql.WithComments(true),
	}
}

func (ProxyUserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("ID"),
		field.Int64("user_id").Comment("用户ID"),
		field.Int64("subscribe_id").Optional().Nillable().Comment("订阅ID"),
		field.String("ip").MaxLen(191).Optional().Nillable().Comment("设备IP"),
		field.String("user_agent").MaxLen(64).Optional().Nillable().Comment("设备User Agent"),
		field.String("identifier").StorageKey("Identifier").MaxLen(191).Optional().Nillable().Comment("设备标识符"),
		field.String("short_code").MaxLen(255).Default("").Comment("短码"),
		field.Bool("online").Default(false).Comment("是否在线"),
		field.Bool("enabled").Default(true).Comment("是否启用"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyUserDevice) Edges() []ent.Edge {
	return nil
}
