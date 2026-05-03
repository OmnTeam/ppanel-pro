package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// StatusCacheKey 服务器状态缓存键格式
	StatusCacheKey = "node:status:%d"
	// OnlineUserCacheKeyWithSubscribe 在线用户订阅缓存键格式
	OnlineUserCacheKeyWithSubscribe = "node:online:subscribe:%d:%s"
	// OnlineUserSubscribeCacheKeyWithGlobal 全局在线订阅缓存键
	OnlineUserSubscribeCacheKeyWithGlobal = "node:online:subscribe:global"
	// CacheExpiry 缓存过期时间（秒）
	CacheExpiry = 300

	// Legacy auth/common cache keys must stay byte-for-byte compatible with the
	// old project because verification flows and rate limits are shared.
	AuthCodeCacheKey          = "auth:verify:email"
	AuthCodeTelephoneCacheKey = "auth:verify:telephone"
	SendIntervalKeyPrefix     = "send:interval:"
	SendCountLimitKeyPrefix   = "send:limit:"
	RegisterIpKeyPrefix       = "register:ip:"

	// System config cache keys
	CurrencyConfigKey   = "system:currency_config"
	InviteConfigKey     = "system:invite_config"
	NodeConfigKey       = "system:node_config"
	TosConfigKey        = "system:tos_config"
	RegisterConfigKey   = "system:register_config"
	SiteConfigKey       = "system:site_config"
	SubscribeConfigKey  = "system:subscribe_config"
	VerifyCodeConfigKey = "system:verify_code_config"
	VerifyConfigKey     = "system:verify_config"
	GlobalConfigKey     = "system:global_config"
	CommonStatCacheKey  = "common:stat"
)

// Status 服务器状态
type Status struct {
	CPU       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`
	UpdatedAt int     `json:"updated_at"`
}

// OnlineUserSubscribe 在线用户订阅映射 map[subscribeID][]IP
type OnlineUserSubscribe map[int64][]string

type onlineUserScore struct {
	expireAt int64
}

func RebuildOnlineUserSubscribeGlobalCache(ctx context.Context, rdb redis.UniversalClient) error {
	if rdb == nil {
		return nil
	}

	now := time.Now()
	scores := make(map[int64]int64)
	var cursor uint64

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "node:online:subscribe:*", 200).Result()
		if err != nil {
			if err == redis.Nil {
				break
			}
			return err
		}
		for _, key := range keys {
			if key == OnlineUserSubscribeCacheKeyWithGlobal {
				continue
			}
			raw, err := rdb.Get(ctx, key).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}
			if raw == "" {
				continue
			}
			ttl, err := rdb.TTL(ctx, key).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}
			if ttl <= 0 {
				continue
			}

			var onlineUsers OnlineUserSubscribe
			if err := json.Unmarshal([]byte(raw), &onlineUsers); err != nil {
				continue
			}

			expireAt := now.Add(ttl).Unix()
			for sid := range onlineUsers {
				if expireAt > scores[sid] {
					scores[sid] = expireAt
				}
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	pipe := rdb.Pipeline()
	pipe.Del(ctx, OnlineUserSubscribeCacheKeyWithGlobal)
	for sid, expireAt := range scores {
		pipe.ZAdd(ctx, OnlineUserSubscribeCacheKeyWithGlobal, redis.Z{
			Score:  float64(expireAt),
			Member: sid,
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

// StatusCache 获取服务器状态缓存 - 严格按照原始逻辑
func (d *Data) StatusCache(ctx context.Context, serverID int) (*Status, error) {
	key := fmt.Sprintf(StatusCacheKey, serverID)

	result, err := d.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 缓存不存在，返回空状态（不报错）
			return &Status{}, nil
		}
		return nil, err
	}

	if result == "" {
		return &Status{}, nil
	}

	var status Status
	if err := json.Unmarshal([]byte(result), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// OnlineUserSubscribe 获取在线用户订阅信息 - 严格按照原始逻辑
func (d *Data) OnlineUserSubscribe(ctx context.Context, serverID int64, protocol string) (OnlineUserSubscribe, error) {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverID, protocol)

	result, err := d.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 缓存不存在，返回空map（不报错）
			return OnlineUserSubscribe{}, nil
		}
		return nil, err
	}

	if result == "" {
		return OnlineUserSubscribe{}, nil
	}

	var subscribe OnlineUserSubscribe
	if err := json.Unmarshal([]byte(result), &subscribe); err != nil {
		return nil, err
	}

	return subscribe, nil
}
