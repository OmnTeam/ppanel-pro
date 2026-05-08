package telegramcallback

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/npanel-pro/ent/proxysystem"
	"github.com/OmnTeam/npanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/npanel-pro/internal/data"
	authmodel "github.com/OmnTeam/npanel-pro/internal/model/auth"
	"github.com/OmnTeam/npanel-pro/pkg/constant"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const bindMessageTemplate = "Your Telegram account has been bound successfully.\n\nUser ID: {{.Id}}\nBind Time: {{.Time}}"

func ResolveBotToken(ctx context.Context, dataLayer *data.Data) (string, error) {
	if dataLayer == nil || dataLayer.DB() == nil {
		return "", nil
	}

	method, err := dataLayer.DB().ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("telegram")).
		Only(ctx)
	if err == nil && method != nil && strings.TrimSpace(method.Config) != "" {
		var cfg authmodel.TelegramAuthConfig
		if cfgErr := cfg.Unmarshal(method.Config); cfgErr == nil && strings.TrimSpace(cfg.BotToken) != "" {
			return strings.TrimSpace(cfg.BotToken), nil
		}
	}

	entries, err := dataLayer.DB().ProxySystem.Query().
		Where(proxysystem.CategoryEQ("telegram")).
		Order(ent.Desc(proxysystem.FieldUpdatedAt), ent.Desc(proxysystem.FieldID)).
		All(ctx)
	if err != nil {
		return "", err
	}

	for _, lookupKey := range []string{"bot_token", "BotToken"} {
		for _, entry := range entries {
			if strings.TrimSpace(entry.Key) == lookupKey {
				return strings.TrimSpace(entry.Value), nil
			}
		}
	}
	return "", nil
}

func HandleUpdate(ctx context.Context, dataLayer *data.Data, update *tgbotapi.Update, botToken string, logger log.Logger) {
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
		sendMessage(botToken, chatID, "Please bind account!")
		return
	}

	sessionKey := fmt.Sprintf("%s:%s", constant.SessionIdKey, sessionID)
	userIDText, err := dataLayer.Redis().Get(ctx, sessionKey).Result()
	if err != nil || strings.TrimSpace(userIDText) == "" {
		sendMessage(botToken, chatID, "Bind failed!")
		return
	}

	userID, err := strconv.ParseInt(userIDText, 10, 64)
	if err != nil {
		sendMessage(botToken, chatID, "Bind failed!")
		return
	}

	method, err := dataLayer.DB().ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ("telegram"),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		helper.Errorf("[telegramcallback] query auth method failed: %v", err)
		sendMessage(botToken, chatID, "Bind failed!")
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
			helper.Errorf("[telegramcallback] create auth method failed: %v", err)
			sendMessage(botToken, chatID, "Bind failed!")
			return
		}
	} else {
		if _, err := dataLayer.DB().ProxyUserAuthMethod.UpdateOneID(method.ID).
			SetAuthIdentifier(identifier).
			SetVerified(true).
			Save(ctx); err != nil {
			helper.Errorf("[telegramcallback] update auth method failed: %v", err)
			sendMessage(botToken, chatID, "Bind failed!")
			return
		}
	}

	clearUserCache(ctx, dataLayer, userID)

	text, renderErr := tool.RenderTemplateToString(bindMessageTemplate, map[string]string{
		"Id":   strconv.FormatInt(userID, 10),
		"Time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if renderErr != nil {
		text = "Bind success!"
	}
	sendMessage(botToken, chatID, text)
}

func clearUserCache(ctx context.Context, dataLayer *data.Data, userID int64) {
	if dataLayer == nil || dataLayer.DB() == nil || dataLayer.Redis() == nil {
		return
	}

	keys := []string{fmt.Sprintf("cache:user:id:%d", userID)}
	methods, err := dataLayer.DB().ProxyUserAuthMethod.Query().
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

	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		if key = strings.TrimSpace(key); key != "" {
			filtered = append(filtered, key)
		}
	}
	if len(filtered) == 0 {
		return
	}
	_ = dataLayer.Redis().Del(ctx, filtered...).Err()
}

func sendMessage(botToken string, chatID int64, text string) {
	if strings.TrimSpace(botToken) == "" || chatID == 0 || strings.TrimSpace(text) == "" {
		return
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	_, _ = bot.Send(msg)
}
