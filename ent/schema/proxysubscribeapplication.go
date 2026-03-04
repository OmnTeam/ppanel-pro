package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxySubscribeApplication holds the schema definition for the ProxySubscribeApplication entity
// 订阅应用配置表 - 定义各种客户端应用(Shadowrocket, Clash等)的配置
type ProxySubscribeApplication struct {
	ent.Schema
}

// Annotations of the ProxySubscribeApplication
func (ProxySubscribeApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_subscribe_application"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxySubscribeApplication
func (ProxySubscribeApplication) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Comment("应用配置ID"),
		field.Int64("tenant_id").
			Default(0).
			Comment("租户ID"),
		field.String("name").
			MaxLen(255).
			NotEmpty().
			Comment("应用名称"),
		field.String("icon").
			Optional().
			Nillable().
			Comment("应用图标"),
		field.String("description").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("应用描述"),
		field.String("scheme").
			MaxLen(255).
			Comment("应用Scheme"),
		field.String("user_agent").
			MaxLen(255).
			Comment("User Agent"),
		field.Bool("is_default").
			Default(false).
			Comment("是否默认应用"),
		field.Text("subscribe_template").
			Optional().
			Nillable().
			Comment("订阅模板"),
		field.String("output_format").
			MaxLen(50).
			Default("yaml").
			Comment("输出格式"),
		field.String("download_link").
			NotEmpty().
			Comment("下载链接"),
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

// Edges of the ProxySubscribeApplication
func (ProxySubscribeApplication) Edges() []ent.Edge {
	return nil
}
