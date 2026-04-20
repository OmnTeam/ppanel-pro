package handler

import (
	"context"
	"github.com/redis/go-redis/v9"
)

const (
	legacyServerUserListCachePrefix = "server:user:"
	legacyServerConfigCachePrefix   = "server:config:"
)

func clearLegacyServerAllCaches(ctx context.Context, rdb *redis.Client) error {
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
