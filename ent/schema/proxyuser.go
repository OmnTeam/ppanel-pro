package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ProxyUser holds the schema definition for the ProxyUser entity
// 用户管理表
type ProxyUser struct {
	ent.Schema
}

// Annotations of the ProxyUser
func (ProxyUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "proxy_user"},
		entsql.WithComments(true),
	}
}

// Fields of the ProxyUser
func (ProxyUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Comment("用户ID"),
		field.String("password").
			MaxLen(100).
			NotEmpty().
			Comment("用户密码"),
		field.String("algo").
			Default("default").
			MaxLen(20).
			Comment("加密算法"),
		field.String("salt").
			Optional().
			Nillable().
			MaxLen(20).
			Comment("密码盐值"),
		field.Text("avatar").
			Optional().
			Nillable().
			Comment("用户头像"),
		field.Int64("tenant_id").
			Default(0).
			Comment("租户ID"),
		field.Int64("balance").
			Optional().
			Nillable().
			Default(0).
			Comment("用户余额（单位：分）"),
		field.Int64("telegram").
			Optional().
			Nillable().
			Comment("Telegram账号"),
		field.String("refer_code").
			MaxLen(20).
			Optional().
			Nillable().
			Comment("推荐码"),
		field.Int("referer_id").
			Optional().
			Nillable().
			Comment("推荐人ID"),
		field.Int64("commission").
			Optional().
			Nillable().
			Default(0).
			Comment("佣金（单位：分）"),
		field.Int("referral_percentage").
			Default(0).
			Comment("推荐百分比"),
		field.Bool("only_first_purchase").
			Default(true).
			Comment("仅首次购买"),
		field.Int64("gift_amount").
			Optional().
			Nillable().
			Default(0).
			Comment("用户礼品金额（单位：分）"),
		field.Bool("enable").
			Default(true).
			Comment("账号是否启用"),
		field.Bool("is_admin").
			Default(false).
			Comment("是否管理员"),
		field.Bool("valid_email").
			Default(false).
			Comment("邮箱是否验证"),
		field.Bool("enable_email_notify").
			Default(false).
			Comment("启用邮件通知"),
		field.Bool("enable_telegram_notify").
			Default(false).
			Comment("启用Telegram通知"),
		field.Bool("enable_balance_notify").
			Default(false).
			Comment("启用余额变动通知"),
		field.Bool("enable_login_notify").
			Default(false).
			Comment("启用登录通知"),
		field.Bool("enable_subscribe_notify").
			Default(false).
			Comment("启用订阅通知"),
		field.Bool("enable_trade_notify").
			Default(false).
			Comment("启用交易通知"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
		field.Bool("is_del").
			Optional().
			Nillable().
			Default(false).
			Comment("1：正常 0：已删除"),
	}
}

// Edges of the ProxyUser
func (ProxyUser) Edges() []ent.Edge {
	return nil
}