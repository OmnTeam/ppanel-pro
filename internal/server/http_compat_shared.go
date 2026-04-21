package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
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
	UserSubscribeID compatFlexibleInt64 `json:"user_subscribe_id"`
	Note            string              `json:"note"`
}

type compatFlexibleInt64 int64

func (v *compatFlexibleInt64) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*v = 0
		return nil
	}
	if strings.HasPrefix(raw, "\"") {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*v = compatFlexibleInt64(compatParseInt64String(text))
		return nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*v = compatFlexibleInt64(parsed)
	return nil
}

func (v *compatFlexibleInt64) UnmarshalText(text []byte) error {
	raw := strings.TrimSpace(string(text))
	if raw == "" {
		*v = 0
		return nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*v = compatFlexibleInt64(parsed)
	return nil
}

func compatInt64SliceToStringSlice(values []int64) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strconv.FormatInt(value, 10))
	}
	return result
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
	helper := log.NewHelper(log.With(log.DefaultLogger, "module", "server/compat/config"))

	entries, err := dataLayer.DB().ProxySystem.Query().
		Where(proxysystem.CategoryEQ(category)).
		Order(
			ent.Desc(proxysystem.FieldUpdatedAt),
			ent.Desc(proxysystem.FieldID),
		).
		All(ctx)
	if err != nil {
		return "", err
	}

	for _, lookupKey := range keys {
		for _, entry := range entries {
			if strings.TrimSpace(entry.Key) == strings.TrimSpace(lookupKey) {
				helper.Infof(
					"[compatSystemValue] category=%s lookup_keys=%v matched_key=%q matched_value=%q mode=exact",
					category,
					keys,
					entry.Key,
					entry.Value,
				)
				return entry.Value, nil
			}
		}
	}

	for _, lookupKey := range keys {
		normalizedLookup := compatNormalizeConfigKey(lookupKey)
		candidates := make([]string, 0)
		for _, entry := range entries {
			if compatNormalizeConfigKey(entry.Key) != normalizedLookup {
				continue
			}
			candidates = append(candidates, fmt.Sprintf("%s=%s", entry.Key, entry.Value))
		}
		if len(candidates) == 0 {
			continue
		}
		entry := candidates[0]
		parts := strings.SplitN(entry, "=", 2)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		helper.Infof(
			"[compatSystemValue] category=%s lookup_keys=%v matched_key=%q matched_value=%q mode=normalized candidates=%v",
			category,
			keys,
			parts[0],
			value,
			candidates,
		)
		return value, nil
	}

	helper.Errorf("[compatSystemValue] category=%s lookup_keys=%v not found", category, keys)
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

func compatBindQuery(ctx khttp.Context, req interface{}) error {
	if req == nil {
		return nil
	}
	return ctx.BindQuery(req)
}

func compatBindBodyAndQuery(ctx khttp.Context, req interface{}) error {
	if req == nil {
		return nil
	}
	if err := ctx.Bind(req); err != nil {
		return err
	}
	return ctx.BindQuery(req)
}
