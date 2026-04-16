package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyorder"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdeviceonlinerecord"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	userBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/user"
	authmodel "github.com/OmnTeam/ppanel-pro/internal/model/auth"
	systemlog "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/OmnTeam/ppanel-pro/pkg/deduction"
	appleoauth "github.com/OmnTeam/ppanel-pro/pkg/oauth/apple"
	googleoauth "github.com/OmnTeam/ppanel-pro/pkg/oauth/google"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
	"github.com/OmnTeam/ppanel-pro/pkg/random"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/oauth2"
)

const (
	legacyCacheUserIDPrefix             = "cache:user:id:"
	legacyCacheUserEmailPrefix          = "cache:user:email:"
	legacyCacheUserSubscribeTokenPrefix = "cache:user:subscribe:token:"
	legacyCacheUserSubscribeUserPrefix  = "cache:user:subscribe:user:"
	legacyCacheUserSubscribeIDPrefix    = "cache:user:subscribe:id:"
	legacyCacheSubscribeIDPrefix        = "cache:subscribe:id:"
	legacyCacheSubscribeServersPrefix   = "cache:subscribe:servers:"
	legacyDeviceIdentifierSessionPrefix = "auth:device_identifier"
	legacyTelegramUnbindMessageTemplate = "Your account has been unbound.\n\nUser ID: {{.Id}}\nTime: {{.Time}}\n"
)

var _ userBiz.UserRepo = (*publicUserRepo)(nil)

type publicUserRepo struct {
	data   *Data
	logger *kratoslog.Helper
}

type verifyCodePayload struct {
	Code   string `json:"code"`
	LastAt int64  `json:"lastAt"`
}

type googleBindCallback struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type appleBindCallback struct {
	Code    string `json:"code"`
	IDToken string `json:"id_token"`
	State   string `json:"state"`
}

func NewPublicUserRepo(d *Data, logger kratoslog.Logger) userBiz.UserRepo {
	return &publicUserRepo{
		data:   d,
		logger: kratoslog.NewHelper(logger),
	}
}

func getAuthTypePriority(authType string) int {
	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "email":
		return 1
	case "mobile":
		return 2
	default:
		return 100
	}
}

func maskOpenID(openID string) string {
	if len(openID) <= 6 {
		return "***"
	}
	maskLen := len(openID) - 6
	mask := strings.Repeat("*", maskLen)
	return openID[:3] + mask + openID[len(openID)-3:]
}

func legacyUnixMillis(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	if value.Unix() == 0 {
		return 0
	}
	return value.UnixMilli()
}

func legacyTimeOrZero(value *time.Time) time.Time {
	if value == nil {
		return time.UnixMilli(0)
	}
	return *value
}

func isLegacyUnlimitedTime(value *time.Time) bool {
	if value == nil {
		return true
	}
	return value.Unix() == 0
}

func parseUserSubscribeDiscounts(raw *string) []*userBiz.SubscribeDiscount {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	var items []struct {
		Quantity   int32 `json:"quantity"`
		Percentage int32 `json:"percentage"`
	}
	if err := json.Unmarshal([]byte(*raw), &items); err != nil {
		return nil
	}
	result := make([]*userBiz.SubscribeDiscount, 0, len(items))
	for _, item := range items {
		result = append(result, &userBiz.SubscribeDiscount{
			Quantity:   item.Quantity,
			Percentage: item.Percentage,
		})
	}
	return result
}

func parseLegacyUnitTime(unitTime string) int32 {
	switch strings.ToLower(strings.TrimSpace(unitTime)) {
	case "", "nolimit":
		return 0
	case "day":
		return 1
	case "month":
		return 2
	case "year":
		return 3
	case "hour":
		return 4
	case "minute":
		return 5
	default:
		if value, err := strconv.Atoi(unitTime); err == nil {
			return int32(value)
		}
		return 0
	}
}

func calculateNextResetTimeLegacy(expireTime int64, resetCycle int64) int64 {
	if expireTime == 0 || resetCycle <= 0 {
		return 0
	}
	return tool.CalculateNextResetTime(expireTime, int32(resetCycle))
}

func pickAffiliateIdentifier(methods []*ent.ProxyUserAuthMethod) string {
	if len(methods) == 0 {
		return ""
	}

	selected := methods[0]
	for _, item := range methods {
		switch strings.ToLower(strings.TrimSpace(item.AuthType)) {
		case "6", "7", "email", "mobile":
			selected = item
			goto MASK
		}
	}

MASK:
	identifier := selected.AuthIdentifier
	hideTextLength := len(identifier) / 3
	if hideTextLength > 0 {
		return identifier[:hideTextLength] + "***" + identifier[hideTextLength*2:]
	}
	return identifier
}

func shouldKeepLegacyUserSubscribe(item *ent.ProxyUserSubscribe, now time.Time) bool {
	if item == nil {
		return false
	}
	if item.Status == nil {
		return false
	}
	switch *item.Status {
	case 0, 1, 2, 3:
	default:
		return false
	}

	if isLegacyUnlimitedTime(item.ExpireTime) {
		return true
	}
	if item.ExpireTime != nil && item.ExpireTime.After(now) {
		return true
	}
	if item.FinishedAt != nil && !item.FinishedAt.Before(now.Add(-7*24*time.Hour)) {
		return true
	}
	return false
}

func (r *publicUserRepo) deleteRedisKeys(ctx context.Context, keys ...string) {
	if r == nil || r.data == nil || r.data.rdb == nil || len(keys) == 0 {
		return
	}
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			filtered = append(filtered, key)
		}
	}
	if len(filtered) == 0 {
		return
	}
	if err := r.data.rdb.Del(ctx, filtered...).Err(); err != nil {
		r.logger.Warnw("delete redis keys failed", "error", err, "keys", filtered)
	}
}

func (r *publicUserRepo) clearLegacyUserCaches(ctx context.Context, userID int64, extraEmails ...string) {
	keys := []string{
		fmt.Sprintf("%s%d", legacyCacheUserIDPrefix, userID),
	}

	methods, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(userID)).
		All(ctx)
	if err == nil {
		for _, item := range methods {
			if strings.EqualFold(item.AuthType, "email") {
				keys = append(keys, fmt.Sprintf("%s%s", legacyCacheUserEmailPrefix, item.AuthIdentifier))
			}
		}
	}
	for _, email := range extraEmails {
		email = strings.TrimSpace(email)
		if email != "" {
			keys = append(keys, fmt.Sprintf("%s%s", legacyCacheUserEmailPrefix, email))
		}
	}
	r.deleteRedisKeys(ctx, keys...)
}

func (r *publicUserRepo) clearLegacyUserSubscribeCaches(ctx context.Context, userSub *ent.ProxyUserSubscribe) {
	if userSub == nil {
		return
	}
	keys := []string{
		fmt.Sprintf("%s%d", legacyCacheUserSubscribeUserPrefix, userSub.UserID),
		fmt.Sprintf("%s%d", legacyCacheUserSubscribeIDPrefix, userSub.ID),
	}
	if userSub.Token != nil && strings.TrimSpace(*userSub.Token) != "" {
		keys = append(keys, fmt.Sprintf("%s%s", legacyCacheUserSubscribeTokenPrefix, *userSub.Token))
	}
	r.deleteRedisKeys(ctx, keys...)
}

func (r *publicUserRepo) clearLegacySubscribeCaches(ctx context.Context, subscribeID int64) {
	r.deleteRedisKeys(ctx,
		fmt.Sprintf("%s%d", legacyCacheSubscribeIDPrefix, subscribeID),
		fmt.Sprintf("%s%d", legacyCacheSubscribeServersPrefix, subscribeID),
	)
}

func (r *publicUserRepo) clearLegacyDeviceSessionCaches(ctx context.Context, identifier string) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return
	}
	deviceKey := fmt.Sprintf("%s:%s", legacyDeviceIdentifierSessionPrefix, identifier)
	sessionID, err := r.data.rdb.Get(ctx, deviceKey).Result()
	if err == nil && sessionID != "" {
		r.deleteRedisKeys(ctx, deviceKey, fmt.Sprintf("%s:%s", constant.SessionIdKey, sessionID))
		return
	}
	r.deleteRedisKeys(ctx, deviceKey)
}

func (r *publicUserRepo) clearLegacyServerUserCaches(ctx context.Context) {
	if r == nil || r.data == nil {
		return
	}
	if err := ClearLegacyServerAllCaches(ctx, r.data.rdb); err != nil {
		r.logger.Warnw("delete legacy server caches failed", "error", err)
	}
}

func (r *publicUserRepo) loadVerifyCodePayload(ctx context.Context, keys ...string) (*verifyCodePayload, string, error) {
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value, err := r.data.rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		var payload verifyCodePayload
		if jsonErr := json.Unmarshal([]byte(value), &payload); jsonErr != nil {
			return nil, key, responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
		}
		return &payload, key, nil
	}
	return nil, "", responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
}

func (r *publicUserRepo) resolveTelegramBotToken(ctx context.Context) string {
	method, err := r.data.db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("telegram")).
		Only(ctx)
	if err == nil && method != nil && strings.TrimSpace(method.Config) != "" {
		var cfg authmodel.TelegramAuthConfig
		if cfgErr := cfg.Unmarshal(method.Config); cfgErr == nil && strings.TrimSpace(cfg.BotToken) != "" {
			return strings.TrimSpace(cfg.BotToken)
		}
	}

	configs, err := loadSystemConfigMap(ctx, r.data.db, "telegram")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(systemConfigString(configs, "bot_token", "BotToken", "telegram_bot_token"))
}

func (r *publicUserRepo) resolveTelegramBotName(ctx context.Context) string {
	configs, err := loadSystemConfigMap(ctx, r.data.db, "telegram")
	if err == nil {
		name := strings.TrimSpace(systemConfigString(configs, "bot_name", "BotName", "bot_username", "UserName"))
		if name != "" {
			return strings.TrimPrefix(name, "@")
		}
	}

	botToken := r.resolveTelegramBotToken(ctx)
	if botToken == "" {
		return ""
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return ""
	}
	me, err := bot.GetMe()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(me.UserName, "@")
}

func (r *publicUserRepo) sendTelegramMessage(botToken string, chatID int64, text string) {
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

func (r *publicUserRepo) QueryUserInfo(ctx context.Context, userID int) (*userBiz.UserInfo, error) {
	userInfo, err := r.data.db.ProxyUser.Query().
		Where(proxyuser.IDEQ(int64(userID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	methods, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(int64(userID))).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	authMethods := make([]*userBiz.AuthMethod, 0, len(methods))
	for _, item := range methods {
		identifier := item.AuthIdentifier
		switch strings.ToLower(strings.TrimSpace(item.AuthType)) {
		case "mobile":
			identifier = phone.MaskPhoneNumber(item.AuthIdentifier)
		case "email":
		default:
			identifier = maskOpenID(item.AuthIdentifier)
		}
		authMethods = append(authMethods, &userBiz.AuthMethod{
			AuthType:       item.AuthType,
			AuthIdentifier: identifier,
			Verified:       item.Verified,
		})
	}
	sort.Slice(authMethods, func(i, j int) bool {
		return getAuthTypePriority(authMethods[i].AuthType) < getAuthTypePriority(authMethods[j].AuthType)
	})

	return &userBiz.UserInfo{
		ID:                    userInfo.ID,
		Avatar:                stringPointerValue(userInfo.Avatar),
		Balance:               int64Value(userInfo.Balance),
		Commission:            int64Value(userInfo.Commission),
		ReferralPercentage:    int32(userInfo.ReferralPercentage),
		OnlyFirstPurchase:     userInfo.OnlyFirstPurchase,
		GiftAmount:            int64Value(userInfo.GiftAmount),
		Telegram:              int64Value(userInfo.Telegram),
		ReferCode:             stringPointerValue(userInfo.ReferCode),
		RefererID:             int64Value(userInfo.RefererID),
		Enable:                userInfo.Enable,
		IsAdmin:               userInfo.IsAdmin,
		EnableBalanceNotify:   userInfo.EnableBalanceNotify,
		EnableLoginNotify:     userInfo.EnableLoginNotify,
		EnableSubscribeNotify: userInfo.EnableSubscribeNotify,
		EnableTradeNotify:     userInfo.EnableTradeNotify,
		AuthMethods:           authMethods,
		CreatedAt:             userInfo.CreatedAt.UnixMilli(),
		UpdatedAt:             userInfo.UpdatedAt.UnixMilli(),
	}, nil
}

func (r *publicUserRepo) GetLoginLog(ctx context.Context, userID int, page, size int) ([]*userBiz.LoginLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(systemlog.TypeLogin)),
			proxysystemlog.ObjectIDEQ(int64(userID)),
		)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	logs, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		Limit(size).
		Offset((page - 1) * size).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.LoginLog, 0, len(logs))
	for _, datum := range logs {
		var content systemlog.Login
		if err := content.Unmarshal([]byte(datum.Content)); err != nil {
			continue
		}
		list = append(list, &userBiz.LoginLog{
			ID:        datum.ID,
			UserID:    datum.ObjectID,
			LoginIP:   content.LoginIP,
			UserAgent: content.UserAgent,
			Success:   content.Success,
			Timestamp: datum.CreatedAt.UnixMilli(),
		})
	}

	return list, int64(total), nil
}

func (r *publicUserRepo) QueryUserBalanceLog(ctx context.Context, userID int) ([]*userBiz.BalanceLog, int64, error) {
	logs, err := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(systemlog.TypeBalance)),
			proxysystemlog.ObjectIDEQ(int64(userID)),
		).
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.BalanceLog, 0, len(logs))
	for _, datum := range logs {
		var content systemlog.Balance
		if err := content.Unmarshal([]byte(datum.Content)); err != nil {
			continue
		}
		list = append(list, &userBiz.BalanceLog{
			Type:      int32(content.Type),
			UserID:    datum.ObjectID,
			Amount:    content.Amount,
			OrderNo:   content.OrderNo,
			Balance:   content.Balance,
			Timestamp: content.Timestamp,
		})
	}

	return list, int64(len(logs)), nil
}

func (r *publicUserRepo) QueryUserCommissionLog(ctx context.Context, userID int, page, size int) ([]*userBiz.CommissionLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(systemlog.TypeCommission)),
			proxysystemlog.ObjectIDEQ(int64(userID)),
		)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	logs, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		Limit(size).
		Offset((page - 1) * size).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.CommissionLog, 0, len(logs))
	for _, datum := range logs {
		var content systemlog.Commission
		if err := content.Unmarshal([]byte(datum.Content)); err != nil {
			continue
		}
		list = append(list, &userBiz.CommissionLog{
			Type:      int32(content.Type),
			UserID:    datum.ObjectID,
			Amount:    content.Amount,
			OrderNo:   content.OrderNo,
			Timestamp: content.Timestamp,
		})
	}

	return list, int64(total), nil
}

func (r *publicUserRepo) QueryUserAffiliate(ctx context.Context, userID int) (int64, int64, error) {
	registers, err := r.data.db.ProxyUser.Query().
		Where(proxyuser.RefererIDEQ(int64(userID))).
		Count(ctx)
	if err != nil {
		return 0, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	logs, err := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(systemlog.TypeCommission)),
			proxysystemlog.ObjectIDEQ(int64(userID)),
		).
		All(ctx)
	if err != nil {
		return 0, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	var totalCommission int64
	for _, datum := range logs {
		var content systemlog.Commission
		if err := content.Unmarshal([]byte(datum.Content)); err != nil {
			continue
		}
		totalCommission += content.Amount
	}

	return int64(registers), totalCommission, nil
}

func (r *publicUserRepo) QueryUserAffiliateList(ctx context.Context, userID int, page, size int) ([]*userBiz.UserAffiliate, int64, error) {
	query := r.data.db.ProxyUser.Query().
		Where(proxyuser.RefererIDEQ(int64(userID)))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	users, err := query.
		Order(ent.Desc(proxyuser.FieldID)).
		Limit(size).
		Offset((page - 1) * size).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.UserAffiliate, 0, len(users))
	for _, item := range users {
		methods, methodErr := r.data.db.ProxyUserAuthMethod.Query().
			Where(proxyuserauthmethod.UserIDEQ(item.ID)).
			All(ctx)
		if methodErr != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}

		list = append(list, &userBiz.UserAffiliate{
			Avatar:       stringPointerValue(item.Avatar),
			Identifier:   pickAffiliateIdentifier(methods),
			RegisteredAt: item.CreatedAt.UnixMilli(),
			Enable:       item.Enable,
		})
	}

	return list, int64(total), nil
}

func (r *publicUserRepo) GetOAuthMethods(ctx context.Context, userID int) ([]*userBiz.AuthMethod, error) {
	methods, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserIDEQ(int64(userID))).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.AuthMethod, 0, len(methods))
	for _, item := range methods {
		list = append(list, &userBiz.AuthMethod{
			AuthType:       item.AuthType,
			AuthIdentifier: item.AuthIdentifier,
			Verified:       item.Verified,
		})
	}
	return list, nil
}

func (r *publicUserRepo) QueryUserSubscribe(ctx context.Context, userID int) ([]*userBiz.UserSubscribe, int64, error) {
	subscriptions, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.UserIDEQ(int64(userID)),
			proxyusersubscribe.StatusIn(0, 1, 2, 3),
		).
		Order(ent.Desc(proxyusersubscribe.FieldID)).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	now := time.Now()
	list := make([]*userBiz.UserSubscribe, 0, len(subscriptions))
	for _, item := range subscriptions {
		if !shouldKeepLegacyUserSubscribe(item, now) {
			continue
		}

		subscribePlan, err := r.data.db.ProxySubscribe.Get(ctx, item.SubscribeID)
		if err != nil {
			return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}

		subscribeInfo := &userBiz.Subscribe{
			ID:             subscribePlan.ID,
			Name:           subscribePlan.Name,
			Description:    stringPointerValue(subscribePlan.Description),
			Price:          subscribePlan.UnitPrice,
			Traffic:        subscribePlan.Traffic,
			DeviceLimit:    int32(subscribePlan.DeviceLimit),
			SpeedLimit:     int32(subscribePlan.SpeedLimit),
			UnitTime:       parseLegacyUnitTime(subscribePlan.UnitTime),
			UnitPrice:      subscribePlan.UnitPrice,
			ResetCycle:     int32(int64Value(subscribePlan.ResetCycle)),
			DeductionRatio: int32(int64Value(subscribePlan.DeductionRatio)),
			AllowDeduction: subscribePlan.AllowDeduction,
			Discount:       parseUserSubscribeDiscounts(subscribePlan.Discount),
			Enable:         subscribePlan.Sell,
			CreatedAt:      subscribePlan.CreatedAt.UnixMilli(),
			UpdatedAt:      subscribePlan.UpdatedAt.UnixMilli(),
		}

		expireAt := legacyUnixMillis(item.ExpireTime)
		userSubscribe := &userBiz.UserSubscribe{
			ID:          item.ID,
			UserID:      item.UserID,
			OrderID:     item.OrderID,
			SubscribeID: item.SubscribeID,
			Subscribe:   subscribeInfo,
			StartTime:   item.StartTime.UnixMilli(),
			ExpireTime:  expireAt,
			FinishedAt:  legacyUnixMillis(item.FinishedAt),
			ResetTime:   calculateNextResetTimeLegacy(expireAt, int64Value(subscribePlan.ResetCycle)),
			Traffic:     int64Value(item.Traffic),
			Download:    int64Value(item.Download),
			Upload:      int64Value(item.Upload),
			Token:       stringPointerValue(item.Token),
			Status: int32(int64Value(func() *int64 {
				if item.Status == nil {
					return nil
				}
				v := int64(*item.Status)
				return &v
			}())),
			CreatedAt: item.CreatedAt.UnixMilli(),
			UpdatedAt: item.UpdatedAt.UnixMilli(),
		}
		if item.Status != nil {
			userSubscribe.Status = int32(*item.Status)
		}

		list = append(list, userSubscribe)
	}

	return list, int64(len(list)), nil
}

func (r *publicUserRepo) GetSubscribeLog(ctx context.Context, userID int, page, size int) ([]*userBiz.UserSubscribeLog, int64, error) {
	query := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(systemlog.TypeSubscribe)),
			proxysystemlog.ObjectIDEQ(int64(userID)),
		)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	logs, err := query.
		Order(ent.Desc(proxysystemlog.FieldCreatedAt)).
		Limit(size).
		Offset((page - 1) * size).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.UserSubscribeLog, 0, len(logs))
	for _, datum := range logs {
		var content systemlog.Subscribe
		if err := content.Unmarshal([]byte(datum.Content)); err != nil {
			continue
		}
		list = append(list, &userBiz.UserSubscribeLog{
			ID:              datum.ID,
			UserID:          datum.ObjectID,
			UserSubscribeID: content.UserSubscribeId,
			Token:           content.Token,
			IP:              content.ClientIP,
			UserAgent:       content.UserAgent,
			Timestamp:       datum.CreatedAt.UnixMilli(),
		})
	}

	return list, int64(total), nil
}

func (r *publicUserRepo) ResetUserSubscribeToken(ctx context.Context, userID, userSubscribeID int) error {
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(int64(userSubscribeID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if userSub.UserID != int64(userID) {
		return responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}

	orderNo := ""
	if userSub.OrderID != 0 {
		orderInfo, err := r.data.db.ProxyOrder.Get(ctx, userSub.OrderID)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		orderNo = orderInfo.OrderNo
	}

	newToken := uuidx.SubscribeToken(orderNo + time.Now().Format("20060102150405.000"))
	newUUID := uuidx.NewUUID().String()

	updated, err := r.data.db.ProxyUserSubscribe.UpdateOneID(userSub.ID).
		SetToken(newToken).
		SetUUID(newUUID).
		Save(ctx)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	r.clearLegacyUserSubscribeCaches(ctx, updated)
	r.clearLegacySubscribeCaches(ctx, updated.SubscribeID)
	r.clearLegacyServerUserCaches(ctx)
	return nil
}

func (r *publicUserRepo) calculateRemainingAmount(ctx context.Context, userID, userSubscribeID int) (int64, error) {
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(int64(userSubscribeID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if userSub.UserID != int64(userID) {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}
	if userSub.OrderID == 0 {
		return 0, nil
	}
	if userSub.Status == nil || *userSub.Status != 1 {
		return 0, kratoserrors.BadRequest("SUBSCRIBE_NOT_IN_USE", "The subscription package is not in use")
	}

	subscribePlan, err := r.data.db.ProxySubscribe.Get(ctx, userSub.SubscribeID)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if !subscribePlan.AllowDeduction {
		return 0, kratoserrors.BadRequest("DEDUCTION_NOT_ALLOWED", "The subscription package does not support deductions")
	}

	orderInfo, err := r.data.db.ProxyOrder.Get(ctx, userSub.OrderID)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	orderAmount := orderInfo.Amount + orderInfo.GiftAmount
	orderQuantity := orderInfo.Quantity

	subOrders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.ParentIDEQ(orderInfo.ID),
			proxyorder.StatusIn(2, 5),
		).
		All(ctx)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	for _, subOrder := range subOrders {
		orderAmount += subOrder.Amount + subOrder.GiftAmount
		orderQuantity += subOrder.Quantity
	}

	expireTime := legacyTimeOrZero(userSub.ExpireTime)
	if !expireTime.After(userSub.StartTime) {
		return 0, nil
	}

	remainingAmount, err := deduction.CalculateRemainingAmount(
		deduction.Subscribe{
			StartTime:      userSub.StartTime,
			ExpireTime:     expireTime,
			Traffic:        int64Value(userSub.Traffic),
			Download:       int64Value(userSub.Download),
			Upload:         int64Value(userSub.Upload),
			UnitTime:       subscribePlan.UnitTime,
			UnitPrice:      subscribePlan.UnitPrice,
			ResetCycle:     int64Value(subscribePlan.ResetCycle),
			DeductionRatio: int64Value(subscribePlan.DeductionRatio),
		},
		deduction.Order{
			Amount:   orderAmount,
			Quantity: orderQuantity,
		},
	)
	if err != nil {
		return 0, kratoserrors.InternalServer("DEDUCTION_ERROR", fmt.Sprintf("CalculateRemainingAmount failed: %v", err))
	}
	return remainingAmount, nil
}

func (r *publicUserRepo) PreUnsubscribe(ctx context.Context, userID, id int) (int64, error) {
	return r.calculateRemainingAmount(ctx, userID, id)
}

func (r *publicUserRepo) Unsubscribe(ctx context.Context, userID, id int) error {
	userSub, err := r.data.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDEQ(int64(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if userSub.UserID != int64(userID) {
		return responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}
	if userSub.Status == nil || (*userSub.Status != 0 && *userSub.Status != 1 && *userSub.Status != 2) {
		return kratoserrors.BadRequest("INVALID_STATUS", "Subscription status invalid for cancellation")
	}

	remainingAmount, err := r.calculateRemainingAmount(ctx, userID, id)
	if err != nil {
		return err
	}

	var updatedSub *ent.ProxyUserSubscribe
	if err := r.data.db.TX(ctx, func(tx *ent.Tx) error {
		userEntity, err := tx.ProxyUser.Get(ctx, int64(userID))
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		orderInfo, err := tx.ProxyOrder.Get(ctx, userSub.OrderID)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}

		updatedSub, err = tx.ProxyUserSubscribe.UpdateOneID(userSub.ID).
			SetStatus(4).
			Save(ctx)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
		}

		currentBalance := int64Value(userEntity.Balance)
		currentGiftAmount := int64Value(userEntity.GiftAmount)

		var balance int64
		var gift int64
		if orderInfo.Method == "balance" {
			if orderInfo.GiftAmount >= remainingAmount {
				gift = remainingAmount
				balance = currentBalance
			} else {
				gift = orderInfo.GiftAmount
				balance = currentBalance + (remainingAmount - orderInfo.GiftAmount)
			}
		} else {
			balance = currentBalance + remainingAmount
			gift = 0
		}

		now := time.Now()
		balanceRefundAmount := balance - currentBalance
		if balanceRefundAmount > 0 {
			content, _ := (&systemlog.Balance{
				OrderNo:   orderInfo.OrderNo,
				Amount:    balanceRefundAmount,
				Type:      systemlog.BalanceTypeRefund,
				Balance:   balance,
				Timestamp: now.UnixMilli(),
			}).Marshal()
			if _, err := tx.ProxySystemLog.Create().
				SetType(int8(systemlog.TypeBalance)).
				SetDate(now.Format(time.DateOnly)).
				SetObjectID(int64(userID)).
				SetContent(string(content)).
				Save(ctx); err != nil {
				return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
			}
		}

		if gift > 0 {
			content, _ := (&systemlog.Gift{
				SubscribeId: updatedSub.ID,
				OrderNo:     orderInfo.OrderNo,
				Type:        systemlog.GiftTypeIncrease,
				Amount:      gift,
				Balance:     currentGiftAmount + gift,
				Remark:      "Unsubscribe refund",
				Timestamp:   now.UnixMilli(),
			}).Marshal()
			if _, err := tx.ProxySystemLog.Create().
				SetType(int8(systemlog.TypeGift)).
				SetDate(now.Format(time.DateOnly)).
				SetObjectID(int64(userID)).
				SetContent(string(content)).
				Save(ctx); err != nil {
				return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
			}
		}

		updateUser := tx.ProxyUser.UpdateOneID(int64(userID))
		if balanceRefundAmount > 0 {
			updateUser = updateUser.SetBalance(balance)
		}
		if gift > 0 {
			updateUser = updateUser.SetGiftAmount(currentGiftAmount + gift)
		}
		if err := updateUser.Exec(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
		}
		return nil
	}); err != nil {
		return err
	}

	r.clearLegacyUserCaches(ctx, int64(userID))
	r.clearLegacyUserSubscribeCaches(ctx, updatedSub)
	r.clearLegacySubscribeCaches(ctx, updatedSub.SubscribeID)
	return nil
}

func (r *publicUserRepo) UpdateUserNotify(ctx context.Context, userID int, enableLoginNotify, enableBalanceNotify, enableSubscribeNotify, enableTradeNotify bool) error {
	if err := r.data.db.ProxyUser.UpdateOneID(int64(userID)).
		SetEnableLoginNotify(enableLoginNotify).
		SetEnableBalanceNotify(enableBalanceNotify).
		SetEnableSubscribeNotify(enableSubscribeNotify).
		SetEnableTradeNotify(enableTradeNotify).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}
	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) UpdateUserPassword(ctx context.Context, userID int, password string) error {
	if err := r.data.db.ProxyUser.UpdateOneID(int64(userID)).
		SetPassword(tool.EncodePassWord(password)).
		SetAlgo("default").
		ClearSalt().
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}
	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) BindTelegram(ctx context.Context, session string, botName string) (string, int64, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return "", 0, responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}
	if strings.TrimSpace(botName) == "" {
		botName = r.resolveTelegramBotName(ctx)
	}
	if strings.TrimSpace(botName) == "" {
		return "", 0, kratoserrors.InternalServer("TELEGRAM_BOT_NOT_CONFIGURED", "telegram bot is not configured")
	}

	return fmt.Sprintf("https://t.me/%s?start=%s", strings.TrimPrefix(botName, "@"), session),
		time.Now().Add(300 * time.Second).UnixMilli(),
		nil
}

func (r *publicUserRepo) UnbindTelegram(ctx context.Context, userID int) error {
	method, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int64(userID)),
			proxyuserauthmethod.AuthTypeEQ("telegram"),
		).
		Only(ctx)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	chatID, err := strconv.ParseInt(strings.TrimSpace(method.AuthIdentifier), 10, 64)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}
	if chatID == 0 {
		return responsecode.NewKratosError(responsecode.ErrTelegramNotBound)
	}

	if err := r.data.db.ProxyUserAuthMethod.DeleteOneID(method.ID).Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
	}

	r.clearLegacyUserCaches(ctx, int64(userID))

	text, renderErr := tool.RenderTemplateToString(legacyTelegramUnbindMessageTemplate, map[string]string{
		"Id":   strconv.FormatInt(int64(userID), 10),
		"Time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if renderErr == nil {
		r.sendTelegramMessage(r.resolveTelegramBotToken(ctx), chatID, text)
	}
	return nil
}

func (r *publicUserRepo) BindOAuth(ctx context.Context, method, redirect string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "google":
		authMethod, err := r.data.db.ProxyAuthMethod.Query().
			Where(proxyauthmethod.MethodEQ("google")).
			Only(ctx)
		if err != nil {
			return "", responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		var cfg authmodel.GoogleAuthConfig
		if err := cfg.Unmarshal(authMethod.Config); err != nil {
			return "", kratoserrors.InternalServer("INVALID_OAUTH_CONFIG", err.Error())
		}
		client := googleoauth.New(&googleoauth.Config{
			ClientID:     cfg.ClientId,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  redirect,
		})
		stateCode := random.KeyNew(8, 1)
		if err := r.data.rdb.Set(ctx, fmt.Sprintf("google:%s", stateCode), redirect, 5*time.Minute).Err(); err != nil {
			return "", kratoserrors.InternalServer("REDIS_ERROR", err.Error())
		}
		return client.AuthCodeURL(stateCode, oauth2.AccessTypeOffline), nil

	case "apple":
		authMethod, err := r.data.db.ProxyAuthMethod.Query().
			Where(proxyauthmethod.MethodEQ("apple")).
			Only(ctx)
		if err != nil {
			return "", responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		var cfg authmodel.AppleAuthConfig
		if err := cfg.Unmarshal(authMethod.Config); err != nil {
			return "", kratoserrors.InternalServer("INVALID_OAUTH_CONFIG", err.Error())
		}
		stateCode := random.KeyNew(8, 1)
		if err := r.data.rdb.Set(ctx, fmt.Sprintf("apple:%s", stateCode), redirect, 5*time.Minute).Err(); err != nil {
			return "", kratoserrors.InternalServer("REDIS_ERROR", err.Error())
		}
		return fmt.Sprintf(
			"https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=name email&response_mode=form_post",
			cfg.ClientId,
			fmt.Sprintf("%s/v1/auth/oauth/callback/apple", cfg.RedirectURL),
			stateCode,
		), nil

	case "github", "facebook":
		return "", nil

	default:
		return "", kratoserrors.BadRequest("UNSUPPORTED_OAUTH_METHOD", fmt.Sprintf("oauth login method not support: %s", method))
	}
}

func (r *publicUserRepo) BindOAuthCallback(ctx context.Context, userID int, method string, callback string) error {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "google":
		var payload googleBindCallback
		if err := json.Unmarshal([]byte(callback), &payload); err != nil {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		redirect, err := r.data.rdb.Get(ctx, fmt.Sprintf("google:%s", payload.State)).Result()
		if err != nil {
			return kratoserrors.BadRequest("STATE_CODE_INVALID", "get google state code failed")
		}

		authMethod, err := r.data.db.ProxyAuthMethod.Query().
			Where(proxyauthmethod.MethodEQ("google")).
			Only(ctx)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}

		var cfg authmodel.GoogleAuthConfig
		if err := json.Unmarshal([]byte(authMethod.Config), &cfg); err != nil {
			return kratoserrors.InternalServer("INVALID_OAUTH_CONFIG", err.Error())
		}

		client := googleoauth.New(&googleoauth.Config{
			ClientID:     cfg.ClientId,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  redirect,
		})
		token, err := client.Exchange(ctx, payload.Code)
		if err != nil {
			return kratoserrors.InternalServer("OAUTH_EXCHANGE_FAILED", err.Error())
		}
		googleUserInfo, err := client.GetUserInfo(token.AccessToken)
		if err != nil {
			return kratoserrors.InternalServer("OAUTH_USERINFO_FAILED", err.Error())
		}

		existing, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthTypeEQ("google"),
				proxyuserauthmethod.AuthIdentifierEQ(googleUserInfo.OpenID),
			).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		if err == nil && existing != nil {
			return responsecode.NewKratosError(responsecode.ErrUserAlreadyExists)
		}

		if _, err := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(int64(userID)).
			SetAuthType("google").
			SetAuthIdentifier(googleUserInfo.OpenID).
			SetVerified(true).
			Save(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}

	case "apple":
		var payload appleBindCallback
		if err := json.Unmarshal([]byte(callback), &payload); err != nil {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		if _, err := r.data.rdb.Get(ctx, fmt.Sprintf("apple:%s", payload.State)).Result(); err != nil {
			return kratoserrors.BadRequest("STATE_CODE_INVALID", "get apple state code failed")
		}

		authMethod, err := r.data.db.ProxyAuthMethod.Query().
			Where(proxyauthmethod.MethodEQ("apple")).
			Only(ctx)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}

		var cfg authmodel.AppleAuthConfig
		if err := json.Unmarshal([]byte(authMethod.Config), &cfg); err != nil {
			return kratoserrors.InternalServer("INVALID_OAUTH_CONFIG", err.Error())
		}

		client, err := appleoauth.New(appleoauth.Config{
			ClientID:     cfg.ClientId,
			TeamID:       cfg.TeamID,
			KeyID:        cfg.KeyID,
			ClientSecret: cfg.ClientSecret,
			RedirectURI:  cfg.RedirectURL,
		})
		if err != nil {
			return kratoserrors.InternalServer("APPLE_CLIENT_INIT_FAILED", err.Error())
		}
		resp, err := client.VerifyWebToken(ctx, payload.Code)
		if err != nil {
			return kratoserrors.InternalServer("APPLE_VERIFY_FAILED", err.Error())
		}
		if resp.Error != "" {
			return kratoserrors.InternalServer("APPLE_VERIFY_FAILED", resp.Error)
		}

		appleUnique, err := appleoauth.GetUniqueID(resp.IDToken)
		if err != nil {
			return kratoserrors.InternalServer("APPLE_UNIQUE_ID_FAILED", err.Error())
		}

		existing, err := r.data.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.AuthTypeEQ("apple"),
				proxyuserauthmethod.AuthIdentifierEQ(appleUnique),
			).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		if err == nil && existing != nil {
			return responsecode.NewKratosError(responsecode.ErrUserAlreadyExists)
		}

		if _, err := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(int64(userID)).
			SetAuthType("apple").
			SetAuthIdentifier(appleUnique).
			SetVerified(true).
			Save(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}

	case "telegram":
		return nil

	default:
		return kratoserrors.BadRequest("UNSUPPORTED_OAUTH_METHOD", fmt.Sprintf("oauth login method not support: %s", method))
	}

	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) UnbindOAuth(ctx context.Context, userID int, method string) error {
	method = strings.ToLower(strings.TrimSpace(method))
	if method == "" || method == "email" || method == "mobile" {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	if _, err := r.data.db.ProxyUserAuthMethod.Delete().
		Where(
			proxyuserauthmethod.UserIDEQ(int64(userID)),
			proxyuserauthmethod.AuthTypeEQ(method),
		).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
	}
	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) VerifyEmail(ctx context.Context, userID int, email, code string) error {
	payload, cacheKey, err := r.loadVerifyCodePayload(
		ctx,
		verifyCodeEmailCacheKey(verifySceneSecurity, email),
		fmt.Sprintf("auth:verify:email:%s:%s", verifySceneSecurity, email),
	)
	if err != nil {
		return err
	}
	if payload.Code != code {
		return responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
	}
	r.deleteRedisKeys(ctx, cacheKey)

	method, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("email"),
			proxyuserauthmethod.AuthIdentifierEQ(email),
		).
		Only(ctx)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if method.UserID != int64(userID) {
		return responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}

	if err := r.data.db.ProxyUserAuthMethod.UpdateOneID(method.ID).
		SetVerified(true).
		Exec(ctx); err != nil {
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}
	r.clearLegacyUserCaches(ctx, int64(userID), email)
	return nil
}

func (r *publicUserRepo) UpdateBindMobile(ctx context.Context, userID int, areaCode, mobile, code string) error {
	phoneNumber, err := phone.FormatToE164(areaCode, mobile)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrTelephoneError)
	}

	payload, cacheKey, err := r.loadVerifyCodePayload(
		ctx,
		verifyCodeTelephoneCacheKey(verifySceneRegister, phoneNumber),
		fmt.Sprintf("auth:verify:telephone:%s:%s", verifySceneRegister, phoneNumber),
	)
	if err != nil {
		return err
	}
	if payload.Code != code {
		return responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
	}
	r.deleteRedisKeys(ctx, cacheKey)

	existing, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("mobile"),
			proxyuserauthmethod.AuthIdentifierEQ(mobile),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if err == nil && existing != nil {
		return responsecode.NewKratosError(responsecode.ErrTelephoneExist)
	}

	current, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int64(userID)),
			proxyuserauthmethod.AuthTypeEQ("mobile"),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	if ent.IsNotFound(err) {
		if _, err := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(int64(userID)).
			SetAuthType("mobile").
			SetAuthIdentifier(mobile).
			SetVerified(true).
			Save(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}
	} else {
		if err := r.data.db.ProxyUserAuthMethod.UpdateOneID(current.ID).
			SetAuthIdentifier(mobile).
			SetVerified(true).
			Exec(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
		}
	}

	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) UpdateBindEmail(ctx context.Context, userID int, email string) error {
	existing, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("email"),
			proxyuserauthmethod.AuthIdentifierEQ(email),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if err == nil && existing != nil {
		return responsecode.NewKratosError(responsecode.ErrDuplicateEmail)
	}

	current, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int64(userID)),
			proxyuserauthmethod.AuthTypeEQ("email"),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	if ent.IsNotFound(err) {
		if _, err := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(int64(userID)).
			SetAuthType("email").
			SetAuthIdentifier(email).
			SetVerified(false).
			Save(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}
	} else {
		if err := r.data.db.ProxyUserAuthMethod.UpdateOneID(current.ID).
			SetAuthIdentifier(email).
			SetVerified(false).
			Exec(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
		}
	}

	r.clearLegacyUserCaches(ctx, int64(userID), email)
	return nil
}

func (r *publicUserRepo) DeviceWSConnect(ctx context.Context) error {
	return nil
}

func (r *publicUserRepo) GetDeviceList(ctx context.Context, userID int) ([]*userBiz.UserDevice, int64, error) {
	devices, err := r.data.db.ProxyUserDevice.Query().
		Where(proxyuserdevice.UserIDEQ(int64(userID))).
		All(ctx)
	if err != nil {
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	list := make([]*userBiz.UserDevice, 0, len(devices))
	for _, item := range devices {
		list = append(list, &userBiz.UserDevice{
			ID:         item.ID,
			IP:         stringPointerValue(item.IP),
			Identifier: stringPointerValue(item.Identifier),
			UserAgent:  stringPointerValue(item.UserAgent),
			Online:     item.Online,
			Enabled:    item.Enabled,
			CreatedAt:  item.CreatedAt.UnixMilli(),
			UpdatedAt:  item.UpdatedAt.UnixMilli(),
		})
	}
	return list, int64(len(list)), nil
}

func (r *publicUserRepo) UnbindDevice(ctx context.Context, userID, deviceID int) error {
	deviceInfo, err := r.data.db.ProxyUserDevice.Query().
		Where(proxyuserdevice.IDEQ(int64(deviceID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrDeviceNotFound)
		}
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if deviceInfo.UserID != int64(userID) {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err := r.data.db.TX(ctx, func(tx *ent.Tx) error {
		if err := tx.ProxyUserDevice.DeleteOneID(deviceInfo.ID).Exec(ctx); err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
		}

		if identifier := stringPointerValue(deviceInfo.Identifier); identifier != "" {
			_, err := tx.ProxyUserAuthMethod.Delete().
				Where(
					proxyuserauthmethod.AuthTypeEQ("device"),
					proxyuserauthmethod.AuthIdentifierEQ(identifier),
				).
				Exec(ctx)
			if err != nil {
				return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
			}
		}

		count, err := tx.ProxyUserAuthMethod.Query().
			Where(proxyuserauthmethod.UserIDEQ(deviceInfo.UserID)).
			Count(ctx)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
		}
		if count < 1 {
			update := tx.ProxyUser.UpdateOneID(deviceInfo.UserID).
				SetDeletedAt(time.Now()).
				SetEnable(false).
				SetIsDel(0)
			if err := update.Exec(ctx); err != nil {
				return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	r.clearLegacyDeviceSessionCaches(ctx, stringPointerValue(deviceInfo.Identifier))
	r.clearLegacyUserCaches(ctx, int64(userID))
	return nil
}

func (r *publicUserRepo) GetDeviceOnlineStatistics(ctx context.Context, userID int) (*userBiz.DeviceOnlineStatistics, error) {
	var longestSingleConnection int64
	record, err := r.data.db.ProxyUserDeviceOnlineRecord.Query().
		Where(proxyuserdeviceonlinerecord.UserIDEQ(int64(userID))).
		Order(ent.Desc(proxyuserdeviceonlinerecord.FieldOnlineSeconds)).
		First(ctx)
	if err == nil && record != nil && record.OnlineSeconds != nil {
		longestSingleConnection = *record.OnlineSeconds / 60
	}

	var historyContinuousDays int64
	durationRecord, err := r.data.db.ProxyUserDeviceOnlineRecord.Query().
		Where(proxyuserdeviceonlinerecord.UserIDEQ(int64(userID))).
		Order(ent.Desc(proxyuserdeviceonlinerecord.FieldDurationDays)).
		First(ctx)
	if err == nil && durationRecord != nil && durationRecord.DurationDays != nil {
		historyContinuousDays = *durationRecord.DurationDays
	}

	records, err := r.data.db.ProxyUserDeviceOnlineRecord.Query().
		Where(
			proxyuserdeviceonlinerecord.UserIDEQ(int64(userID)),
			proxyuserdeviceonlinerecord.CreatedAtGTE(time.Now().AddDate(0, 0, -7)),
		).
		Order(ent.Desc(proxyuserdeviceonlinerecord.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	currentContinuousDays := int64(1)
	if len(records) > 0 && records[0].DurationDays != nil {
		currentContinuousDays = *records[0].DurationDays
	}

	onlineDays := make(map[string]*userBiz.WeeklyStat)
	for _, item := range records {
		onlineTime := item.CreatedAt
		if item.OnlineTime != nil {
			onlineTime = *item.OnlineTime
		}
		dateKey := onlineTime.Format(time.DateOnly)
		stat, ok := onlineDays[dateKey]
		if !ok {
			stat = &userBiz.WeeklyStat{
				DayName: onlineTime.Weekday().String(),
			}
			onlineDays[dateKey] = stat
		}
		if item.OnlineSeconds != nil {
			stat.Hours += float64(*item.OnlineSeconds)
		}
	}

	keys := make([]string, 0, 7)
	for i := 0; i < 7; i++ {
		dateKey := time.Now().AddDate(0, 0, -i).Format(time.DateOnly)
		if _, ok := onlineDays[dateKey]; !ok {
			parsed, _ := time.Parse(time.DateOnly, dateKey)
			onlineDays[dateKey] = &userBiz.WeeklyStat{DayName: parsed.Weekday().String()}
		}
		keys = append(keys, dateKey)
	}
	sort.Strings(keys)

	weeklyStats := make([]*userBiz.WeeklyStat, 0, len(keys))
	for index, key := range keys {
		stat := onlineDays[key]
		stat.Day = int32(index + 1)
		stat.Hours = stat.Hours / 3600
		weeklyStats = append(weeklyStats, stat)
	}

	return &userBiz.DeviceOnlineStatistics{
		WeeklyStats: weeklyStats,
		ConnectionRecords: &userBiz.ConnectionRecords{
			CurrentContinuousDays:   currentContinuousDays,
			HistoryContinuousDays:   historyContinuousDays,
			LongestSingleConnection: longestSingleConnection,
		},
	}, nil
}
