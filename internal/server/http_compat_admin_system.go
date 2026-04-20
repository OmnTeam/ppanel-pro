package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	adminsystemv1 "github.com/OmnTeam/ppanel-pro/api/admin/system/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	authmodel "github.com/OmnTeam/ppanel-pro/internal/model/auth"
	adminsystemservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/system"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type compatLegacyRegisterConfig struct {
	StopRegister            bool   `json:"stop_register"`
	EnableTrial             bool   `json:"enable_trial"`
	TrialSubscribe          int64  `json:"trial_subscribe,string"`
	TrialTime               int64  `json:"trial_time"`
	TrialTimeUnit           string `json:"trial_time_unit"`
	EnableIpRegisterLimit   bool   `json:"enable_ip_register_limit"`
	IpRegisterLimit         int64  `json:"ip_register_limit"`
	IpRegisterLimitDuration int64  `json:"ip_register_limit_duration"`
	DeviceLimit             int64  `json:"device_limit"`
}

type compatLegacyNodeMultiplierPreview struct {
	CurrentTime string  `json:"current_time"`
	Ratio       float32 `json:"ratio"`
}

type compatLegacyNodeMultiplierPeriod struct {
	StartTime  string
	EndTime    string
	Multiplier float32
}

var compatTelegramPolling struct {
	mu  sync.Mutex
	bot *tgbotapi.BotAPI
}

func registerLegacyAdminSystemCompatRoutes(r *khttp.Router, dataLayer *data.Data, adminSystem *adminsystemservice.SystemService) {
	if r == nil {
		return
	}

	registerLegacyAdminSystemMutationCompatRoutes(r, adminSystem)

	r.GET("/v1/admin/system/register_config", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			return compatLoadLegacyRegisterConfig(inner, dataLayer)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.PUT("/v1/admin/system/register_config", func(ctx khttp.Context) error {
		var req compatLegacyRegisterConfig
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatUpdateLegacyRegisterConfig(inner, dataLayer, request.(*compatLegacyRegisterConfig))
		}); err != nil {
			return compatJSONError(ctx, err)
		}

		return compatJSON(ctx, nil)
	})

	r.GET("/v1/admin/system/node_multiplier/preview", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			return compatPreviewLegacyNodeMultiplier(inner, dataLayer)
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/admin/system/setting_telegram_bot", func(ctx khttp.Context) error {
		_, _ = compatMiddleware(ctx, nil, func(inner context.Context, req interface{}) (interface{}, error) {
			compatSettingTelegramBot(inner, dataLayer)
			return nil, nil
		})
		return compatJSON(ctx, nil)
	})
}

func registerLegacyAdminSystemMutationCompatRoutes(r *khttp.Router, adminSystem *adminsystemservice.SystemService) {
	if r == nil || adminSystem == nil {
		return
	}

	r.PUT("/v1/admin/system/currency_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateCurrencyConfigRequest) (interface{}, error) {
		return adminSystem.UpdateCurrencyConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/invite_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateInviteConfigRequest) (interface{}, error) {
		return adminSystem.UpdateInviteConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/node_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateNodeConfigRequest) (interface{}, error) {
		return adminSystem.UpdateNodeConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/privacy", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdatePrivacyPolicyConfigRequest) (interface{}, error) {
		return adminSystem.UpdatePrivacyPolicyConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/site_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateSiteConfigRequest) (interface{}, error) {
		return adminSystem.UpdateSiteConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/subscribe_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateSubscribeConfigRequest) (interface{}, error) {
		return adminSystem.UpdateSubscribeConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/tos_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateTosConfigRequest) (interface{}, error) {
		return adminSystem.UpdateTosConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/verify_code_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateVerifyCodeConfigRequest) (interface{}, error) {
		return adminSystem.UpdateVerifyCodeConfig(ctx, req)
	}))
	r.PUT("/v1/admin/system/verify_config", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.UpdateVerifyConfigRequest) (interface{}, error) {
		return adminSystem.UpdateVerifyConfig(ctx, req)
	}))
	r.POST("/v1/admin/system/set_node_multiplier", compatAdminSystemMutation(func(ctx context.Context, req *adminsystemv1.SetNodeMultiplierRequest) (interface{}, error) {
		return adminSystem.SetNodeMultiplier(ctx, req)
	}))
}

func compatAdminSystemMutation[T any](fn func(context.Context, *T) (interface{}, error)) func(khttp.Context) error {
	return func(ctx khttp.Context) error {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		if _, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return fn(inner, request.(*T))
		}); err != nil {
			return compatJSONError(ctx, err)
		}

		return compatJSON(ctx, nil)
	}
}

func compatLoadLegacyRegisterConfig(ctx context.Context, dataLayer *data.Data) (*compatLegacyRegisterConfig, error) {
	if dataLayer == nil {
		return nil, compatCodeError(500, "data layer unavailable")
	}

	repo := data.NewAdminSystemRepo(dataLayer, log.DefaultLogger)
	configs, err := repo.GetConfigByCategory(ctx, "register")
	if err != nil {
		return nil, err
	}

	result := &compatLegacyRegisterConfig{}
	tool.SystemConfigSliceReflectToStruct(configs, result)
	return result, nil
}

func compatUpdateLegacyRegisterConfig(ctx context.Context, dataLayer *data.Data, req *compatLegacyRegisterConfig) error {
	if dataLayer == nil {
		return compatCodeError(500, "data layer unavailable")
	}
	if req == nil {
		return compatParamError("invalid request")
	}

	repo := data.NewAdminSystemRepo(dataLayer, log.DefaultLogger)
	return repo.UpdateConfigByCategory(ctx, "register", compatStructToSystemConfigs(req))
}

func compatPreviewLegacyNodeMultiplier(ctx context.Context, dataLayer *data.Data) (*compatLegacyNodeMultiplierPreview, error) {
	now := time.Now()
	ratio := float32(1)

	if dataLayer != nil {
		repo := data.NewAdminSystemRepo(dataLayer, log.DefaultLogger)
		value, err := repo.GetNodeMultiplier(ctx)
		if err != nil {
			return nil, err
		}
		for _, period := range compatParseLegacyNodeMultiplierPeriods(value) {
			if compatLegacyTimeWithinPeriod(now, period.StartTime, period.EndTime) {
				ratio = period.Multiplier
				break
			}
		}
	}

	return &compatLegacyNodeMultiplierPreview{
		CurrentTime: now.Format("2006-01-02 15:04:05"),
		Ratio:       ratio,
	}, nil
}

func compatParseLegacyNodeMultiplierPeriods(raw string) []compatLegacyNodeMultiplierPeriod {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var objects []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &objects); err != nil {
		return nil
	}

	periods := make([]compatLegacyNodeMultiplierPeriod, 0, len(objects))
	for _, item := range objects {
		startTime := compatMapString(item, "start_time", "StartTime")
		endTime := compatMapString(item, "end_time", "EndTime")
		multiplier := compatMapFloat32(item, "multiplier", "Multiplier")
		periods = append(periods, compatLegacyNodeMultiplierPeriod{
			StartTime:  startTime,
			EndTime:    endTime,
			Multiplier: multiplier,
		})
	}
	return periods
}

func compatMapString(item map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := item[key]; ok {
			switch v := value.(type) {
			case string:
				return strings.TrimSpace(v)
			case fmt.Stringer:
				return strings.TrimSpace(v.String())
			default:
				return strings.TrimSpace(fmt.Sprint(v))
			}
		}
	}
	return ""
}

func compatMapFloat32(item map[string]interface{}, keys ...string) float32 {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case float32:
			return v
		case float64:
			return float32(v)
		case int:
			return float32(v)
		case int32:
			return float32(v)
		case int64:
			return float32(v)
		case json.Number:
			if parsed, err := v.Float64(); err == nil {
				return float32(parsed)
			}
		case string:
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 32); err == nil {
				return float32(parsed)
			}
		}
	}
	return 0
}

func compatLegacyTimeWithinPeriod(current time.Time, start, end string) bool {
	startTime, err := time.Parse("15:04.000", start)
	if err != nil {
		return false
	}
	endTime, err := time.Parse("15:04.000", end)
	if err != nil {
		return false
	}

	currentTime := time.Date(0, 1, 1, current.Hour(), current.Minute(), 0, 0, time.UTC)
	startFormatted := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endFormatted := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	if startFormatted.Before(endFormatted) {
		return currentTime.After(startFormatted) && currentTime.Before(endFormatted)
	}
	return currentTime.After(startFormatted) || currentTime.Before(endFormatted)
}

func compatStructToSystemConfigs(value interface{}) map[string]*tool.SystemConfig {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	result := make(map[string]*tool.SystemConfig, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		result[fieldType.Name] = &tool.SystemConfig{
			Key:   fieldType.Name,
			Value: tool.ConvertValueToString(field),
			Type:  compatConfigFieldType(field),
		}
	}
	return result
}

func compatConfigFieldType(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "int"
	case reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int64"
	case reflect.Float32, reflect.Float64:
		return "float"
	default:
		return "string"
	}
}

func compatSettingTelegramBot(ctx context.Context, dataLayer *data.Data) {
	if dataLayer == nil || dataLayer.DB() == nil {
		return
	}

	method, err := dataLayer.DB().ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("telegram")).
		Only(ctx)
	if err != nil || method == nil || strings.TrimSpace(method.Config) == "" {
		return
	}

	var cfg authmodel.TelegramAuthConfig
	if err := cfg.Unmarshal(method.Config); err != nil || strings.TrimSpace(cfg.BotToken) == "" {
		return
	}

	bot, err := tgbotapi.NewBotAPI(strings.TrimSpace(cfg.BotToken))
	if err != nil {
		return
	}

	webhookDomain := strings.TrimSpace(cfg.WebHookDomain)
	if webhookDomain != "" {
		if wh, err := tgbotapi.NewWebhook(fmt.Sprintf("%s/v1/telegram/webhook?secret=%s", webhookDomain, tool.Md5Encode(cfg.BotToken, false))); err == nil {
			_, _ = bot.Request(wh)
		}
		compatStopTelegramLongPolling()
	} else {
		compatStartTelegramLongPolling(dataLayer, bot, cfg.BotToken)
	}

	me, err := bot.GetMe()
	if err != nil {
		return
	}

	repo := data.NewAdminSystemRepo(dataLayer, log.DefaultLogger)
	_ = repo.UpdateConfigByCategory(ctx, "telegram", map[string]*tool.SystemConfig{
		"bot_token": {
			Key:   "bot_token",
			Value: strings.TrimSpace(cfg.BotToken),
			Type:  "string",
		},
		"bot_name": {
			Key:   "bot_name",
			Value: strings.TrimPrefix(strings.TrimSpace(me.UserName), "@"),
			Type:  "string",
		},
		"bot_id": {
			Key:   "bot_id",
			Value: strconv.FormatInt(int64(me.ID), 10),
			Type:  "int64",
		},
		"enable_notify": {
			Key:   "enable_notify",
			Value: strconv.FormatBool(cfg.EnableNotify),
			Type:  "bool",
		},
		"webhook_domain": {
			Key:   "webhook_domain",
			Value: webhookDomain,
			Type:  "string",
		},
	})
}

func compatStartTelegramLongPolling(dataLayer *data.Data, bot *tgbotapi.BotAPI, botToken string) {
	if bot == nil || strings.TrimSpace(botToken) == "" {
		return
	}

	compatTelegramPolling.mu.Lock()
	oldBot := compatTelegramPolling.bot
	compatTelegramPolling.bot = bot
	compatTelegramPolling.mu.Unlock()

	if oldBot != nil && oldBot != bot {
		oldBot.StopReceivingUpdates()
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	go func(currentBot *tgbotapi.BotAPI) {
		for update := range updates {
			compatHandleTelegramUpdate(context.Background(), dataLayer, &update, botToken, log.DefaultLogger)
		}

		compatTelegramPolling.mu.Lock()
		if compatTelegramPolling.bot == currentBot {
			compatTelegramPolling.bot = nil
		}
		compatTelegramPolling.mu.Unlock()
	}(bot)
}

func compatStopTelegramLongPolling() {
	compatTelegramPolling.mu.Lock()
	bot := compatTelegramPolling.bot
	compatTelegramPolling.bot = nil
	compatTelegramPolling.mu.Unlock()

	if bot != nil {
		bot.StopReceivingUpdates()
	}
}

func compatHandleTelegramUpdate(ctx context.Context, dataLayer *data.Data, update *tgbotapi.Update, botToken string, logger log.Logger) {
	helper := log.NewHelper(logger)
	if update == nil || update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
		return
	}
	if update.Message.Command() != "start" {
		return
	}
	if dataLayer == nil || dataLayer.DB() == nil || dataLayer.Redis() == nil {
		return
	}

	chatID := update.Message.Chat.ID
	sessionID := strings.TrimSpace(update.Message.CommandArguments())
	if sessionID == "" {
		compatSendTelegramMessage(botToken, chatID, "Please bind account!")
		return
	}

	sessionKey := fmt.Sprintf("%s:%s", constant.SessionIdKey, sessionID)
	userIDText, err := dataLayer.Redis().Get(ctx, sessionKey).Result()
	if err != nil || strings.TrimSpace(userIDText) == "" {
		compatSendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	userID, err := strconv.ParseInt(userIDText, 10, 64)
	if err != nil {
		compatSendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	method, err := dataLayer.DB().ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ("telegram"),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		helper.Errorf("[compatTelegramUpdate] query auth method failed: %v", err)
		compatSendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	identifier := strconv.FormatInt(chatID, 10)
	if ent.IsNotFound(err) {
		if _, err := dataLayer.DB().ProxyUserAuthMethod.Create().
			SetUserID(userID).
			SetAuthType("telegram").
			SetAuthIdentifier(identifier).
			SetVerified(true).
			Save(ctx); err != nil {
			helper.Errorf("[compatTelegramUpdate] create auth method failed: %v", err)
			compatSendTelegramMessage(botToken, chatID, "Bind failed!")
			return
		}
	} else {
		if _, err := dataLayer.DB().ProxyUserAuthMethod.UpdateOneID(method.ID).
			SetAuthIdentifier(identifier).
			SetVerified(true).
			Save(ctx); err != nil {
			helper.Errorf("[compatTelegramUpdate] update auth method failed: %v", err)
			compatSendTelegramMessage(botToken, chatID, "Bind failed!")
			return
		}
	}

	emails, _ := compatUserEmails(ctx, dataLayer, userID)
	compatClearUserCache(ctx, dataLayer.Redis(), userID, emails...)

	text, renderErr := tool.RenderTemplateToString(compatTelegramBindMessage, map[string]string{
		"Id":   strconv.FormatInt(userID, 10),
		"Time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if renderErr != nil {
		text = "Bind success!"
	}
	compatSendTelegramMessage(botToken, chatID, text)
}
