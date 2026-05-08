package handler

import (
	"context"
	"encoding/json"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/internal/logic/order"
	"github.com/OmnTeam/npanel-pro/internal/queue/types"
	"github.com/OmnTeam/npanel-pro/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// CloseOrderHandler 关闭超时订单处理器
// 完整复刻原项目逻辑：server-master/queue/logic/order/deferCloseOrderLogic.go
type CloseOrderHandler struct {
	db           *ent.Client
	logger       log.Logger
	cacheService *service.CacheService
}

// NewCloseOrderHandler 创建关闭订单处理器
func NewCloseOrderHandler(db *ent.Client, logger log.Logger, cacheService *service.CacheService) *CloseOrderHandler {
	return &CloseOrderHandler{
		db:           db,
		logger:       logger,
		cacheService: cacheService,
	}
}

// ProcessTask 处理任务
// 关闭超时未支付的订单（15分钟后）
// 完整复刻原项目逻辑：server-master/queue/logic/order/deferCloseOrderLogic.go:26-47
func (h *CloseOrderHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// 1. 解析payload (复刻原项目 line 27-34)
	payload := types.DeferCloseOrderPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.NewHelper(h.logger).Errorf("[CloseOrder] unmarshal payload failed: %v", err)
		return nil // 返回nil避免重试
	}

	// 2. 调用独立的 CloseOrder 业务逻辑 (复刻原项目 line 36-38)
	err := logic.NewCloseOrderLogic(ctx, h.db, h.logger, h.cacheService).CloseOrder(&logic.CloseOrderRequest{
		OrderNo: payload.OrderNo,
	})

	// 3. 处理重试逻辑 (复刻原项目 line 39-46)
	count, ok := asynq.GetRetryCount(ctx)
	if !ok {
		return nil
	}
	if err != nil && count < 3 {
		return err
	}
	return nil
}
