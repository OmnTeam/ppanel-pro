package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/redis/go-redis/v9"
)

const (
	compatLegacyCacheUserIDPrefix             = "cache:user:id:"
	compatLegacyCacheUserEmailPrefix          = "cache:user:email:"
	compatLegacyCacheUserSubscribeTokenPrefix = "cache:user:subscribe:token:"
	compatLegacyCacheUserSubscribeUserPrefix  = "cache:user:subscribe:user:"
	compatLegacyCacheUserSubscribeIDPrefix    = "cache:user:subscribe:id:"
	compatLegacyCacheSubscribeIDPrefix        = "cache:subscribe:id:"
	compatLegacyCacheSubscribeServersPrefix   = "cache:subscribe:servers:"
)

type compatPathTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

type compatDailyTrafficStats struct {
	Date     string `json:"date"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    int64  `json:"total"`
}

type compatTrafficStatsData struct {
	TotalUpload   int64                     `json:"total_upload"`
	TotalDownload int64                     `json:"total_download"`
	TotalTraffic  int64                     `json:"total_traffic"`
	List          []compatDailyTrafficStats `json:"list"`
}

type compatUpdateUserRulesRequest struct {
	Rules []string `json:"rules"`
}

type compatUpdateUserSubscribeNoteRequest struct {
	UserSubscribeID int64  `json:"user_subscribe_id"`
	Note            string `json:"note"`
}

func compatCurrentUser(ctx context.Context) (*ent.ProxyUser, error) {
	user, ok := ctx.Value(constant.CtxKeyUser).(*ent.ProxyUser)
	if !ok || user == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}
	return user, nil
}

func compatCurrentSessionID(ctx context.Context) string {
	sessionID, _ := ctx.Value(constant.CtxKeySessionID).(string)
	return strings.TrimSpace(sessionID)
}

func compatUnix(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.Unix()
}

func compatInt64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func compatInt8Value(value *int8) int8 {
	if value == nil {
		return 0
	}
	return *value
}

func compatStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func compatSystemValue(ctx context.Context, dataLayer *data.Data, category string, keys ...string) (string, error) {
	if dataLayer == nil || dataLayer.DB() == nil {
		return "", fmt.Errorf("data layer unavailable")
	}

	entries, err := dataLayer.DB().ProxySystem.Query().
		Where(proxysystem.CategoryEQ(category)).
		All(ctx)
	if err != nil {
		return "", err
	}

	for _, lookupKey := range keys {
		normalizedLookup := compatNormalizeConfigKey(lookupKey)
		for _, entry := range entries {
			if compatNormalizeConfigKey(entry.Key) == normalizedLookup {
				return entry.Value, nil
			}
		}
	}

	return "", fmt.Errorf("system config not found: %s", strings.Join(keys, ","))
}

func compatNormalizeConfigKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "_", "")
	return key
}

func compatUserEmails(ctx context.Context, dataLayer *data.Data, userID int64) ([]string, error) {
	if dataLayer == nil || dataLayer.DB() == nil {
		return nil, fmt.Errorf("data layer unavailable")
	}

	methods, err := dataLayer.DB().ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ("email"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(methods))
	for _, item := range methods {
		if email := strings.TrimSpace(item.AuthIdentifier); email != "" {
			result = append(result, email)
		}
	}
	return result, nil
}

func compatDeleteKeys(ctx context.Context, rdb *redis.Client, keys ...string) {
	if rdb == nil || len(keys) == 0 {
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
	_ = rdb.Del(ctx, filtered...).Err()
}

func compatClearUserCache(ctx context.Context, rdb *redis.Client, userID int64, extraEmails ...string) {
	keys := []string{fmt.Sprintf("%s%d", compatLegacyCacheUserIDPrefix, userID)}
	for _, email := range extraEmails {
		if email = strings.TrimSpace(email); email != "" {
			keys = append(keys, fmt.Sprintf("%s%s", compatLegacyCacheUserEmailPrefix, email))
		}
	}
	compatDeleteKeys(ctx, rdb, keys...)
}

func compatClearUserSubscribeCaches(ctx context.Context, rdb *redis.Client, userSub *ent.ProxyUserSubscribe) {
	if userSub == nil {
		return
	}

	keys := []string{
		fmt.Sprintf("%s%d", compatLegacyCacheUserSubscribeUserPrefix, userSub.UserID),
		fmt.Sprintf("%s%d", compatLegacyCacheUserSubscribeIDPrefix, userSub.ID),
	}
	if userSub.Token != nil && strings.TrimSpace(*userSub.Token) != "" {
		keys = append(keys, fmt.Sprintf("%s%s", compatLegacyCacheUserSubscribeTokenPrefix, *userSub.Token))
	}
	compatDeleteKeys(ctx, rdb, keys...)
}

func compatClearSubscribeCaches(ctx context.Context, rdb *redis.Client, subscribeID int64) {
	compatDeleteKeys(
		ctx,
		rdb,
		fmt.Sprintf("%s%d", compatLegacyCacheSubscribeIDPrefix, subscribeID),
		fmt.Sprintf("%s%d", compatLegacyCacheSubscribeServersPrefix, subscribeID),
	)
}

func compatValidateRequiredString(value, typeName, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return compatRequiredFieldError(typeName, fieldName)
	}
	return nil
}

func compatValidateRequiredInt(value int, typeName, fieldName string) error {
	if value == 0 {
		return compatRequiredFieldError(typeName, fieldName)
	}
	return nil
}

func compatRequiredFieldError(typeName, fieldName string) error {
	return compatParamError(fmt.Sprintf(
		"Key: '%s.%s' Error:Field validation for '%s' failed on the 'required' tag",
		typeName,
		fieldName,
		fieldName,
	))
}
