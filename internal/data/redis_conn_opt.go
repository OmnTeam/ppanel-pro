package data

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// compatibleRedisConnOpt keeps Asynq on the same Redis compatibility policy as
// the main application client, so older Redis deployments don't emit startup
// handshake warnings.
type compatibleRedisConnOpt struct {
	asynq.RedisClientOpt
}

func (opt compatibleRedisConnOpt) MakeRedisClient() interface{} {
	return redis.NewClient(&redis.Options{
		Network:      opt.Network,
		Addr:         opt.Addr,
		Username:     opt.Username,
		Password:     opt.Password,
		DB:           opt.DB,
		DialTimeout:  opt.DialTimeout,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		PoolSize:     opt.PoolSize,
		TLSConfig:    opt.TLSConfig,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})
}
