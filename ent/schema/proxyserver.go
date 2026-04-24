package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyServer holds the schema definition for the ProxyServer entity.
type ProxyServer struct {
	ent.Schema
}

func (ProxyServer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "servers"},
		entsql.WithComments(true),
	}
}

func (ProxyServer) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Comment("服务器ID"),
		field.String("name").MaxLen(100).Default("").Comment("服务器名称"),
		field.String("country").MaxLen(128).Default("").Comment("国家"),
		field.String("city").MaxLen(128).Default("").Comment("城市"),
		field.String("server_addr").StorageKey("address").MaxLen(100).Default("").Comment("服务器地址"),
		field.Int32("sort").Default(0).Comment("排序"),
		field.Text("protocol").StorageKey("protocols").Optional().Comment("协议配置JSON"),
		field.Time("last_reported_at").Optional().Nillable().Comment("最后报告时间"),
		field.String("longitude").MaxLen(50).Default("0.0").Comment("经度"),
		field.String("latitude").MaxLen(50).Default("0.0").Comment("纬度"),
		field.String("longitude_center").MaxLen(50).Default("0.0").Comment("中心经度"),
		field.String("latitude_center").MaxLen(50).Default("0.0").Comment("中心纬度"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

func (ProxyServer) Edges() []ent.Edge {
	return nil
}
