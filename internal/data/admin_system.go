package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/npanel-pro/ent/proxysystem"
	"github.com/OmnTeam/npanel-pro/ent/proxyuserauthmethod"
	systembiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/system"
	authmodel "github.com/OmnTeam/npanel-pro/internal/model/auth"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/OmnTeam/npanel-pro/pkg/constant"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type adminSystemRepo struct {
	data *Data
	log  *log.Helper
}

var adminTelegramPolling struct {
	mu  sync.Mutex
	bot *tgbotapi.BotAPI
}

func LoadNodeConfigForServer(ctx context.Context, data *Data, logger log.Logger) (*systembiz.NodeConfig, error) {
	repo := &adminSystemRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
	configs, err := repo.GetConfigByCategory(ctx, "server")
	if err != nil {
		return nil, err
	}
	result := &systembiz.NodeConfig{}
	tool.SystemConfigSliceReflectToStruct(configs, result)
	repo.log.Infof(
		"[LoadNodeConfigForServer] node_secret=%q node_pull_interval=%d node_push_interval=%d traffic_report_threshold=%d ip_strategy=%q",
		result.NodeSecret,
		result.NodePullInterval,
		result.NodePushInterval,
		result.TrafficReportThreshold,
		result.IPStrategy,
	)
	return result, nil
}

// NewAdminSystemRepo creates a new admin system repository
func NewAdminSystemRepo(data *Data, logger log.Logger) systembiz.SystemRepo {
	return &adminSystemRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetConfigByCategory 根据分类获取配置
func (r *adminSystemRepo) GetConfigByCategory(ctx context.Context, category string) ([]*tool.SystemConfig, error) {
	// Query from database directly. Anonymous server interfaces are sensitive to
	// stale configuration, so system config caching is intentionally disabled.
	systems, err := r.data.db.ProxySystem.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ(s.C(proxysystem.FieldCategory), category))
		}).
		All(ctx)
	if err != nil {
		r.log.Errorf("[GetConfigByCategory] Failed to query system config for category %s: %v", category, err)
		return nil, responsecode.NewDatabaseQueryError()
	}
	// Convert to tool.SystemConfig and normalize any legacy alias keys back to old-project keys.
	configByKey := make(map[string]*tool.SystemConfig, len(systems))
	for _, sys := range systems {
		key := normalizeSystemConfigKey(sys.Key)
		if _, exists := configByKey[key]; exists && sys.Key != key {
			continue
		}
		configByKey[key] = &tool.SystemConfig{
			Key:   key,
			Value: sys.Value,
			Type:  sys.Type,
		}
	}
	configs := make([]*tool.SystemConfig, 0, len(configByKey))
	for _, config := range configByKey {
		configs = append(configs, config)
	}

	return configs, nil
}

// UpdateConfigByCategory 根据分类更新配置
func (r *adminSystemRepo) UpdateConfigByCategory(ctx context.Context, category string, configs map[string]*tool.SystemConfig) error {
	// Start transaction
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		r.log.Errorf("[UpdateConfigByCategory] Failed to start transaction for category %s: %v", category, err)
		return responsecode.NewDatabaseUpdateError()
	}

	// Update each config
	for key, config := range configs {
		// Check if config exists
		exists, err := tx.ProxySystem.Query().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(
					sql.EQ(s.C(proxysystem.FieldCategory), category),
					sql.EQ(s.C(proxysystem.FieldKey), key),
				))
			}).
			Exist(ctx)
		if err != nil {
			r.log.Errorf("[UpdateConfigByCategory] Failed to check config existence for category %s, key %s: %v", category, key, err)
			tx.Rollback()
			return responsecode.NewDatabaseQueryError()
		}

		if exists {
			// Update existing config
			affected, err := tx.ProxySystem.Update().
				Where(func(s *sql.Selector) {
					s.Where(sql.And(
						sql.EQ(s.C(proxysystem.FieldCategory), category),
						sql.EQ(s.C(proxysystem.FieldKey), key),
					))
				}).
				SetValue(config.Value).
				SetType(config.Type).
				Save(ctx)
			if err != nil {
				r.log.Errorf("[UpdateConfigByCategory] Failed to update config for category %s, key %s: %v", category, key, err)
				tx.Rollback()
				return responsecode.NewDatabaseUpdateError()
			}
			if affected == 0 {
				r.log.Warnf("[UpdateConfigByCategory] Config not found for category %s, key %s", category, key)
				tx.Rollback()
				return responsecode.NewSystemNotFoundError()
			}
		} else {
			// Create new config
			_, err := tx.ProxySystem.Create().
				SetCategory(category).
				SetKey(key).
				SetValue(config.Value).
				SetType(config.Type).
				SetDesc("").
				Save(ctx)
			if err != nil {
				r.log.Errorf("[UpdateConfigByCategory] Failed to create config for category %s, key %s: %v", category, key, err)
				tx.Rollback()
				return responsecode.NewDatabaseUpdateError()
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		r.log.Errorf("[UpdateConfigByCategory] Failed to commit transaction for category %s: %v", category, err)
		return responsecode.NewDatabaseUpdateError()
	}

	syncRuntimeAppConfig(ctx, r.data.db, r.data.conf, r.log)

	return nil
}

// GetNodeMultiplier 获取节点倍率配置
func (r *adminSystemRepo) GetNodeMultiplier(ctx context.Context) (string, error) {
	values, err := loadSystemConfigMap(ctx, r.data.db, "server")
	if err != nil {
		r.log.Errorf("[GetNodeMultiplier] Failed to query node multiplier: %v", err)
		return "", responsecode.NewDatabaseQueryError()
	}

	return systemConfigString(values, "NodeMultiplierConfig", "NodeMultiplier"), nil
}

// UpdateNodeMultiplier 更新节点倍率配置
func (r *adminSystemRepo) UpdateNodeMultiplier(ctx context.Context, value string) error {
	exists, err := r.data.db.ProxySystem.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.And(
				sql.EQ(s.C(proxysystem.FieldCategory), "server"),
				sql.EQ(s.C(proxysystem.FieldKey), "NodeMultiplierConfig"),
			))
		}).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("[UpdateNodeMultiplier] Failed to check node multiplier existence: %v", err)
		return responsecode.NewDatabaseQueryError()
	}

	if exists {
		// Update existing config
		affected, err := r.data.db.ProxySystem.Update().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(
					sql.EQ(s.C(proxysystem.FieldCategory), "server"),
					sql.EQ(s.C(proxysystem.FieldKey), "NodeMultiplierConfig"),
				))
			}).
			SetValue(value).
			Save(ctx)
		if err != nil {
			r.log.Errorf("[UpdateNodeMultiplier] Failed to update node multiplier: %v", err)
			return responsecode.NewDatabaseUpdateError()
		}
		if affected == 0 {
			r.log.Warnf("[UpdateNodeMultiplier] Node multiplier not found")
			return responsecode.NewSystemNotFoundError()
		}
	} else {
		// Create new config
		_, err := r.data.db.ProxySystem.Create().
			SetCategory("server").
			SetKey("NodeMultiplierConfig").
			SetValue(value).
			SetType("string").
			SetDesc("node multiplier config").
			Save(ctx)
		if err != nil {
			r.log.Errorf("[UpdateNodeMultiplier] Failed to create node multiplier: %v", err)
			return responsecode.NewDatabaseUpdateError()
		}
	}

	return nil
}

func (r *adminSystemRepo) ApplyTelegramBot(ctx context.Context) error {
	if r.data == nil || r.data.db == nil {
		return nil
	}

	method, err := r.data.db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("telegram")).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		r.log.Errorf("[ApplyTelegramBot] query telegram auth method failed: %v", err)
		return responsecode.NewDatabaseQueryError()
	}
	if strings.TrimSpace(method.Config) == "" {
		return nil
	}

	var cfg authmodel.TelegramAuthConfig
	if err := cfg.Unmarshal(method.Config); err != nil {
		r.log.Errorf("[ApplyTelegramBot] unmarshal telegram config failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrConfigurationError)
	}
	botToken := strings.TrimSpace(cfg.BotToken)
	if botToken == "" {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		r.log.Errorf("[ApplyTelegramBot] create telegram bot failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrConfigurationError)
	}

	webhookDomain := strings.TrimSpace(cfg.WebHookDomain)
	if webhookDomain != "" {
		webhookURL := fmt.Sprintf("%s/v1/telegram/webhook?secret=%s", webhookDomain, tool.Md5Encode(botToken, false))
		webhook, err := tgbotapi.NewWebhook(webhookURL)
		if err != nil {
			r.log.Errorf("[ApplyTelegramBot] create telegram webhook failed: %v", err)
			return responsecode.NewKratosError(responsecode.ErrConfigurationError)
		}
		if _, err := bot.Request(webhook); err != nil {
			r.log.Errorf("[ApplyTelegramBot] apply telegram webhook failed: %v", err)
			return responsecode.NewKratosError(responsecode.ErrConfigurationError)
		}
		r.stopTelegramLongPolling()
	} else {
		r.startTelegramLongPolling(bot, botToken)
	}

	me, err := bot.GetMe()
	if err != nil {
		r.log.Errorf("[ApplyTelegramBot] get telegram bot info failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrConfigurationError)
	}

	return r.UpdateConfigByCategory(ctx, "telegram", map[string]*tool.SystemConfig{
		"bot_token": {
			Key:   "bot_token",
			Value: botToken,
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

func (r *adminSystemRepo) startTelegramLongPolling(bot *tgbotapi.BotAPI, botToken string) {
	if bot == nil || strings.TrimSpace(botToken) == "" {
		return
	}

	adminTelegramPolling.mu.Lock()
	oldBot := adminTelegramPolling.bot
	adminTelegramPolling.bot = bot
	adminTelegramPolling.mu.Unlock()

	if oldBot != nil && oldBot != bot {
		oldBot.StopReceivingUpdates()
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	go func(currentBot *tgbotapi.BotAPI) {
		for update := range updates {
			r.handleTelegramUpdate(context.Background(), &update, botToken)
		}

		adminTelegramPolling.mu.Lock()
		if adminTelegramPolling.bot == currentBot {
			adminTelegramPolling.bot = nil
		}
		adminTelegramPolling.mu.Unlock()
	}(bot)
}

func (r *adminSystemRepo) stopTelegramLongPolling() {
	adminTelegramPolling.mu.Lock()
	bot := adminTelegramPolling.bot
	adminTelegramPolling.bot = nil
	adminTelegramPolling.mu.Unlock()

	if bot != nil {
		bot.StopReceivingUpdates()
	}
}

func (r *adminSystemRepo) handleTelegramUpdate(ctx context.Context, update *tgbotapi.Update, botToken string) {
	if update == nil || update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
		return
	}
	if update.Message.Command() != "start" {
		return
	}
	if r.data == nil || r.data.db == nil || r.data.rdb == nil {
		return
	}

	chatID := update.Message.Chat.ID
	sessionID := strings.TrimSpace(update.Message.CommandArguments())
	if sessionID == "" {
		r.sendTelegramMessage(botToken, chatID, "Please bind account!")
		return
	}

	sessionKey := fmt.Sprintf("%s:%s", constant.SessionIdKey, sessionID)
	userIDText, err := r.data.rdb.Get(ctx, sessionKey).Result()
	if err != nil || strings.TrimSpace(userIDText) == "" {
		r.sendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	userID, err := strconv.ParseInt(userIDText, 10, 64)
	if err != nil {
		r.sendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	method, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ("telegram"),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		r.log.Errorf("[handleTelegramUpdate] query auth method failed: %v", err)
		r.sendTelegramMessage(botToken, chatID, "Bind failed!")
		return
	}

	identifier := strconv.FormatInt(chatID, 10)
	if ent.IsNotFound(err) {
		if _, err := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(userID).
			SetAuthType("telegram").
			SetAuthIdentifier(identifier).
			SetVerified(true).
			Save(ctx); err != nil {
			r.log.Errorf("[handleTelegramUpdate] create auth method failed: %v", err)
			r.sendTelegramMessage(botToken, chatID, "Bind failed!")
			return
		}
	} else {
		if _, err := r.data.db.ProxyUserAuthMethod.UpdateOneID(method.ID).
			SetAuthIdentifier(identifier).
			SetVerified(true).
			Save(ctx); err != nil {
			r.log.Errorf("[handleTelegramUpdate] update auth method failed: %v", err)
			r.sendTelegramMessage(botToken, chatID, "Bind failed!")
			return
		}
	}

	r.clearTelegramUserCache(ctx, userID)
	text, renderErr := tool.RenderTemplateToString("Bind success!\nAccount ID: {{.Id}}\nTime: {{.Time}}", map[string]string{
		"Id":   strconv.FormatInt(userID, 10),
		"Time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if renderErr != nil {
		text = "Bind success!"
	}
	r.sendTelegramMessage(botToken, chatID, text)
}

func (r *adminSystemRepo) sendTelegramMessage(botToken string, chatID int64, text string) {
	botToken = strings.TrimSpace(botToken)
	if botToken == "" || chatID == 0 || strings.TrimSpace(text) == "" {
		return
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	_, _ = bot.Send(msg)
}

func (r *adminSystemRepo) clearTelegramUserCache(ctx context.Context, userID int64) {
	if r.data == nil || r.data.rdb == nil {
		return
	}
	keys := []string{fmt.Sprintf("cache:user:id:%d", userID)}

	methods, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ("email"),
		).
		All(ctx)
	if err == nil {
		for _, item := range methods {
			if email := strings.TrimSpace(item.AuthIdentifier); email != "" {
				keys = append(keys, fmt.Sprintf("cache:user:email:%s", email))
			}
		}
	}
	r.deleteRedisKeys(ctx, keys...)
}

func (r *adminSystemRepo) deleteRedisKeys(ctx context.Context, keys ...string) {
	if r.data == nil || r.data.rdb == nil || len(keys) == 0 {
		return
	}
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		if key = strings.TrimSpace(key); key != "" {
			filtered = append(filtered, key)
		}
	}
	if len(filtered) == 0 {
		return
	}
	_ = r.data.rdb.Del(ctx, filtered...).Err()
}
