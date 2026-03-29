package data

import (
	"github.com/OmnTeam/ppanel-pro/ent"
	redemptionBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/redemption"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type publicRedemptionRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicRedemptionRepo 创建Public Redemption仓库
func NewPublicRedemptionRepo(data *Data, logger log.Logger) redemptionBiz.RedemptionRepo {
	return &publicRedemptionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetDB 获取数据库客户端
func (r *publicRedemptionRepo) GetDB() *ent.Client {
	return r.data.db
}

// GetRedis 获取Redis客户端
func (r *publicRedemptionRepo) GetRedis() *redis.Client {
	return r.data.rdb
}

// GetQueue 获取队列客户端
func (r *publicRedemptionRepo) GetQueue() *asynq.Client {
	return r.data.queue
}
