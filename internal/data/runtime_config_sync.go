package data

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	authmodel "github.com/OmnTeam/ppanel-pro/internal/model/auth"
	"github.com/go-kratos/kratos/v2/log"
)

// syncRuntimeAppConfig keeps the in-memory application config aligned with the
// old project's startup/update behavior. This is important because several
// public routes and middleware still read from the runtime config snapshot.
func syncRuntimeAppConfig(ctx context.Context, client *ent.Client, appConf *conf.Application, logger *log.Helper) {
	if client == nil || appConf == nil {
		return
	}

	syncRuntimeSystemConfig(ctx, client, appConf, logger)
	syncRuntimeAuthMethodConfig(ctx, client, appConf, logger)
}

func syncRuntimeSystemConfig(ctx context.Context, client *ent.Client, appConf *conf.Application, logger *log.Helper) {
	if values, err := loadSystemConfigMap(ctx, client, "site"); err != nil {
		logger.Warnw("sync runtime site config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Site = &conf.Site{
			Host:       getStringConfigWithDefault(values, "", "Host", "host"),
			SiteName:   getStringConfigWithDefault(values, "", "SiteName", "site_name"),
			SiteDesc:   getStringConfigWithDefault(values, "", "SiteDesc", "site_desc"),
			SiteLogo:   getStringConfigWithDefault(values, "", "SiteLogo", "site_logo"),
			Keywords:   getStringConfigWithDefault(values, "", "Keywords", "keywords"),
			CustomHtml: getStringConfigWithDefault(values, "", "CustomHTML", "custom_html"),
			CustomData: getStringConfigWithDefault(values, "", "CustomData", "custom_data"),
		}
	}

	if values, err := loadSystemConfigMap(ctx, client, "subscribe"); err != nil {
		logger.Warnw("sync runtime subscribe config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Subscribe = &conf.Subscribe{
			SingleModel:     getBoolConfig(values, false, "SingleModel", "single_model"),
			SubscribePath:   getStringConfigWithDefault(values, "", "SubscribePath", "subscribe_path"),
			SubscribeDomain: getStringConfigWithDefault(values, "", "SubscribeDomain", "subscribe_domain"),
			PanDomain:       getBoolConfig(values, false, "PanDomain", "pan_domain"),
			UserAgentLimit:  getBoolConfig(values, false, "UserAgentLimit", "user_agent_limit"),
			UserAgentList:   getStringConfigWithDefault(values, "", "UserAgentList", "user_agent_list"),
		}
	}

	if values, err := loadSystemConfigMap(ctx, client, "register"); err != nil {
		logger.Warnw("sync runtime register config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Register = &conf.Register{
			StopRegister:            getBoolConfig(values, false, "StopRegister", "stop_register"),
			EnableIpRegisterLimit:   getBoolConfig(values, false, "EnableIpRegisterLimit", "enable_ip_register_limit"),
			IpRegisterLimit:         getInt64Config(values, 0, "IpRegisterLimit", "ip_register_limit"),
			IpRegisterLimitDuration: getInt64Config(values, 0, "IpRegisterLimitDuration", "ip_register_limit_duration"),
			EnableTrial:             getBoolConfig(values, false, "EnableTrial", "enable_trial"),
			TrialSubscribe:          getInt64Config(values, 0, "TrialSubscribe", "trial_subscribe"),
			TrialTimeUnit:           getStringConfigWithDefault(values, "", "TrialTimeUnit", "trial_time_unit"),
			TrialTime:               getInt64Config(values, 0, "TrialTime", "trial_time"),
		}
	}

	if values, err := loadSystemConfigMap(ctx, client, "invite"); err != nil {
		logger.Warnw("sync runtime invite config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Invite = &conf.Invite{
			ForcedInvite:       getBoolConfig(values, false, "ForcedInvite", "forced_invite"),
			ReferralPercentage: getInt64Config(values, 0, "ReferralPercentage", "referral_percentage"),
			OnlyFirstPurchase:  getBoolConfig(values, false, "OnlyFirstPurchase", "only_first_purchase"),
		}
	}

	if values, err := loadSystemConfigMap(ctx, client, "verify"); err != nil {
		logger.Warnw("sync runtime verify config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Verify = &conf.Verify{
			TurnstileSiteKey:          getStringConfigWithDefault(values, "", "TurnstileSiteKey", "turnstile_site_key"),
			EnableLoginVerify:         getBoolConfig(values, false, "EnableUserLoginCaptcha", "enable_user_login_captcha", "EnableLoginVerify", "enable_login_verify"),
			EnableRegisterVerify:      getBoolConfig(values, false, "EnableUserRegisterCaptcha", "enable_user_register_captcha", "EnableRegisterVerify", "enable_register_verify"),
			EnableResetPasswordVerify: getBoolConfig(values, false, "EnableUserResetPasswordCaptcha", "enable_user_reset_password_captcha", "EnableResetPasswordVerify", "enable_reset_password_verify"),
		}
	}

	if values, err := loadSystemConfigMap(ctx, client, "server"); err != nil {
		logger.Warnw("sync runtime node config failed", "error", err)
	} else if len(values) > 0 {
		appConf.Node = &conf.Node{
			NodeSecret: getStringConfigWithDefault(values, "", "NodeSecret", "node_secret"),
		}
	}
}

func syncRuntimeAuthMethodConfig(ctx context.Context, client *ent.Client, appConf *conf.Application, logger *log.Helper) {
	methods, err := client.ProxyAuthMethod.Query().
		Order(ent.Asc(proxyauthmethod.FieldID)).
		All(ctx)
	if err != nil {
		logger.Warnw("sync runtime auth methods failed", "error", err)
		return
	}

	for _, method := range methods {
		switch method.Method {
		case "email":
			var raw authmodel.EmailAuthConfig
			raw.Unmarshal(method.Config)
			appConf.Email = &conf.EmailAuth{
				Enable:             method.Enabled,
				EnableVerify:       raw.EnableVerify,
				EnableDomainSuffix: raw.EnableDomainSuffix,
				DomainSuffixList:   raw.DomainSuffixList,
			}
		case "mobile":
			var raw authmodel.MobileAuthConfig
			raw.Unmarshal(method.Config)
			appConf.Mobile = &conf.MobileAuth{
				Enable:          method.Enabled,
				EnableWhitelist: raw.EnableWhitelist,
				Whitelist:       append([]string(nil), raw.Whitelist...),
			}
		case "device":
			var raw authmodel.DeviceConfig
			if err := raw.Unmarshal(method.Config); err != nil {
				logger.Warnw("sync runtime device auth config failed", "error", err)
				continue
			}
			appConf.Device = &conf.Device{
				Enable:         method.Enabled,
				SecuritySecret: raw.SecuritySecret,
			}
		}
	}
}
