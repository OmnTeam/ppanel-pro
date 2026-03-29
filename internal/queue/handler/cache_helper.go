package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	queueLegacyCacheUserIDPrefix             = "cache:user:id:"
	queueLegacyCacheUserEmailPrefix          = "cache:user:email:"
	queueLegacyCacheUserSubscribeTokenPrefix = "cache:user:subscribe:token:"
	queueLegacyCacheUserSubscribeUserPrefix  = "cache:user:subscribe:user:"
	queueLegacyCacheUserSubscribeIDPrefix    = "cache:user:subscribe:id:"
	queueLegacyCacheSubscribeIDPrefix        = "cache:subscribe:id:"
	queueLegacyCacheSubscribeServersPrefix   = "cache:subscribe:servers:"
)

func deleteLegacyRedisKeys(ctx context.Context, rdb *redis.Client, logger *kratoslog.Helper, keys ...string) {
	if rdb == nil || len(keys) == 0 {
		return
	}

	seen := make(map[string]struct{}, len(keys))
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		filtered = append(filtered, key)
	}
	if len(filtered) == 0 {
		return
	}

	if err := rdb.Del(ctx, filtered...).Err(); err != nil && logger != nil {
		logger.Warnw("delete legacy redis keys failed", "error", err, "keys", filtered)
	}
}

func clearLegacyUserSubscribeCaches(ctx context.Context, rdb *redis.Client, logger *kratoslog.Helper, userSubs ...*ent.ProxyUserSubscribe) {
	if rdb == nil || len(userSubs) == 0 {
		return
	}

	keys := make([]string, 0, len(userSubs)*3)
	for _, userSub := range userSubs {
		if userSub == nil {
			continue
		}
		keys = append(keys,
			fmt.Sprintf("%s%d", queueLegacyCacheUserSubscribeUserPrefix, userSub.UserID),
			fmt.Sprintf("%s%d", queueLegacyCacheUserSubscribeIDPrefix, userSub.ID),
		)
		if userSub.Token != nil && strings.TrimSpace(*userSub.Token) != "" {
			keys = append(keys, fmt.Sprintf("%s%s", queueLegacyCacheUserSubscribeTokenPrefix, *userSub.Token))
		}
	}

	deleteLegacyRedisKeys(ctx, rdb, logger, keys...)
}

func clearLegacySubscribeCaches(ctx context.Context, rdb *redis.Client, logger *kratoslog.Helper, subscribeIDs ...int64) {
	if rdb == nil || len(subscribeIDs) == 0 {
		return
	}

	keys := make([]string, 0, len(subscribeIDs)*2)
	seen := make(map[int64]struct{}, len(subscribeIDs))
	for _, subscribeID := range subscribeIDs {
		if subscribeID <= 0 {
			continue
		}
		if _, ok := seen[subscribeID]; ok {
			continue
		}
		seen[subscribeID] = struct{}{}
		keys = append(keys,
			fmt.Sprintf("%s%d", queueLegacyCacheSubscribeIDPrefix, subscribeID),
			fmt.Sprintf("%s%d", queueLegacyCacheSubscribeServersPrefix, subscribeID),
		)
	}

	deleteLegacyRedisKeys(ctx, rdb, logger, keys...)
}

func clearLegacyUserCaches(ctx context.Context, db *ent.Client, rdb *redis.Client, logger *kratoslog.Helper, userIDs ...int64) {
	if db == nil || rdb == nil || len(userIDs) == 0 {
		return
	}

	uniqueUserIDs := make([]int64, 0, len(userIDs))
	seenUserIDs := make(map[int64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seenUserIDs[userID]; ok {
			continue
		}
		seenUserIDs[userID] = struct{}{}
		uniqueUserIDs = append(uniqueUserIDs, userID)
	}
	if len(uniqueUserIDs) == 0 {
		return
	}

	keys := make([]string, 0, len(uniqueUserIDs)*2)
	for _, userID := range uniqueUserIDs {
		keys = append(keys, fmt.Sprintf("%s%d", queueLegacyCacheUserIDPrefix, userID))
	}

	authMethods, err := db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDIn(uniqueUserIDs...),
			proxyuserauthmethod.AuthTypeEQ("email"),
		).
		All(ctx)
	if err != nil {
		if logger != nil {
			logger.Warnw("query legacy user cache emails failed", "error", err, "user_ids", uniqueUserIDs)
		}
		deleteLegacyRedisKeys(ctx, rdb, logger, keys...)
		return
	}

	for _, authMethod := range authMethods {
		email := strings.TrimSpace(authMethod.AuthIdentifier)
		if email == "" {
			continue
		}
		keys = append(keys, fmt.Sprintf("%s%s", queueLegacyCacheUserEmailPrefix, email))
	}

	deleteLegacyRedisKeys(ctx, rdb, logger, keys...)
}
