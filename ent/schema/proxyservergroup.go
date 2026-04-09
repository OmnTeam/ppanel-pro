package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyServerGroup holds the schema definition for the ProxyServerGroup entity.
type ProxyServerGroup struct {
	ent.Schema
}

func (ProxyServerGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "node_group"},
		entsql.WithComments(true),
	}
}

func (ProxyServerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Positive().Comment("ID"),
		field.String("name").MaxLen(255).NotEmpty().Comment("Group Name"),
		field.String("group_type").MaxLen(32).Default("common").Comment("Node Group Type"),
		field.String("description").MaxLen(500).Optional().Comment("Group Description"),
		field.Int("sort").Default(0).Comment("Sort Order"),
		field.Bool("for_calculation").Default(true).Comment("For Calculation"),
		field.Bool("is_expired_group").Default(false).Comment("Is Expired Group"),
		field.Int("expired_days_limit").Default(7).Comment("Expired Days Limit"),
		field.Int64("max_traffic_gb_expired").Optional().Nillable().Comment("Max Traffic GB for Expired Users"),
		field.Int("speed_limit").Default(0).Comment("Speed Limit"),
		field.Int64("min_traffic_gb").Optional().Nillable().Comment("Minimum Traffic"),
		field.Int64("max_traffic_gb").Optional().Nillable().Comment("Maximum Traffic"),
		field.Time("created_at").Default(time.Now).Comment("Creation Time"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("Update Time"),
	}
}

func (ProxyServerGroup) Edges() []ent.Edge {
	return nil
}
