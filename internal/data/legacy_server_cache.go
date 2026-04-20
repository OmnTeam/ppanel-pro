package data

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	legacyServerUserListCachePrefix = "server:user:"
	legacyServerConfigCachePrefix   = "server:config:"
)

func LegacyServerUserListCacheKey(serverID int64) string {
	return fmt.Sprintf("%s%d", legacyServerUserListCachePrefix, serverID)
}

func LegacyServerConfigCacheKey(serverID int64, protocol string) string {
	return fmt.Sprintf("%s%d:%s", legacyServerConfigCachePrefix, serverID, protocol)
}

func ClearLegacyServerCachesByServerIDs(ctx context.Context, rdb *redis.Client, serverIDs []int64) error {
	if rdb == nil || len(serverIDs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(serverIDs)*2)
	seen := make(map[string]struct{})
	for _, serverID := range serverIDs {
		if serverID <= 0 {
			continue
		}

		userKey := LegacyServerUserListCacheKey(serverID)
		if _, ok := seen[userKey]; !ok {
			seen[userKey] = struct{}{}
			keys = append(keys, userKey)
		}

		var cursor uint64
		pattern := fmt.Sprintf("%s%d*", legacyServerConfigCachePrefix, serverID)
		for {
			scanKeys, newCursor, err := rdb.Scan(ctx, cursor, pattern, 999).Result()
			if err != nil {
				return err
			}
			for _, key := range scanKeys {
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				keys = append(keys, key)
			}
			cursor = newCursor
			if cursor == 0 {
				break
			}
		}
	}

	if len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}

func ClearLegacyServerAllCaches(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}

	keys := make([]string, 0, 64)
	seen := make(map[string]struct{})
	for _, pattern := range []string{legacyServerUserListCachePrefix + "*", legacyServerConfigCachePrefix + "*"} {
		var cursor uint64
		for {
			scanKeys, newCursor, err := rdb.Scan(ctx, cursor, pattern, 999).Result()
			if err != nil {
				return err
			}
			for _, key := range scanKeys {
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				keys = append(keys, key)
			}
			cursor = newCursor
			if cursor == 0 {
				break
			}
		}
	}

	if len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}
