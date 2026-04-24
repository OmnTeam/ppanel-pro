package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
