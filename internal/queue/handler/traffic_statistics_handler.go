package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// TrafficStatisticsHandler 实时流量统计处理器
// 用于处理节点服务器上报的实时流量数据
type TrafficStatisticsHandler struct {
	db     *ent.Client
	logger *log.Helper
}

// NewTrafficStatisticsHandler creates a new traffic statistics handler
func NewTrafficStatisticsHandler(db *ent.Client, logger log.Logger) *TrafficStatisticsHandler {
	return &TrafficStatisticsHandler{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask processes the real-time traffic statistics task
func (h *TrafficStatisticsHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload types.TrafficStatistics
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to unmarshal payload: %v", err)
		return nil // Don't retry on malformed payload
	}

	if len(payload.Logs) == 0 {
		h.logger.WithContext(ctx).Warnf("[TrafficStatisticsHandler] Empty payload received for server_id: %d", payload.ServerID)
		return nil
	}

	h.logger.WithContext(ctx).Infof("[TrafficStatisticsHandler] Processing traffic statistics: server_id=%d, protocol=%s, logs=%d",
		payload.ServerID, payload.Protocol, len(payload.Logs))

	// 1. 查询服务器信息获取协议倍率
	serverInfo, err := h.db.ProxyNode.Query().
		Where(proxynode.IDEQ(payload.ServerID)).
		First(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Server not found: %d, error: %v", payload.ServerID, err)
		return nil
	}

	// 2. 获取协议倍率（默认1.0）
	ratio := h.getProtocolRatio(serverInfo, payload.Protocol)

	// 3. 获取实时流量倍率（默认1.0）
	realTimeMultiplier := h.getRealTimeMultiplier()

	now := time.Now()
	processedCount := 0
	skippedCount := 0

	// 4. 处理每个用户的流量数据
	for _, logEntry := range payload.Logs {
		// 查询用户订阅信息
		sub, err := h.db.ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.IDEQ(logEntry.SID)).
			First(ctx)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Subscribe not found: %d", logEntry.SID)
			continue
		}

		// 跳过如果流量低于阈值
		if logEntry.Download+logEntry.Upload <= 100 { // Threshold: 100 bytes
			skippedCount++
			continue
		}

		// 应用倍率
		d := int64(float64(logEntry.Download) * ratio * realTimeMultiplier)
		u := int64(float64(logEntry.Upload) * ratio * realTimeMultiplier)

		// 更新用户订阅流量
		err = h.db.ProxyUserSubscribe.UpdateOne(sub).
			AddUpload(u).
			AddDownload(d).
			Exec(ctx)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to update subscribe: %d, error: %v", logEntry.SID, err)
			continue
		}

		// 创建流量日志记录
		_, err = h.db.ProxyTrafficLog.Create().
			SetServerID(payload.ServerID).
			SetSubscribeID(logEntry.SID).
			SetUserID(sub.UserID).
			SetUpload(int(u)).
			SetDownload(int(d)).
			SetTimestamp(now).
			Save(ctx)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to create traffic log: %v", err)
		}

		processedCount++
	}

	h.logger.WithContext(ctx).Infof("[TrafficStatisticsHandler] Traffic statistics task completed: processed=%d, skipped=%d",
		processedCount, skippedCount)

	return nil
}

// getProtocolRatio 获取协议倍率
func (h *TrafficStatisticsHandler) getProtocolRatio(server *ent.ProxyNode, protocol string) float64 {
	// TODO: 解析服务器的协议配置并获取指定协议的倍率
	// 当前返回默认倍率1.0

	// 示例逻辑（需要根据实际的协议配置结构实现）:
	// protocols, err := server.UnmarshalProtocols()
	// if err != nil {
	//     return 1.0
	// }
	// for _, p := range protocols {
	//     if normalizeProtocol(p.Type) == normalizeProtocol(protocol) && p.Ratio > 0 {
	//         return p.Ratio
	//     }
	// }

	return 1.0
}

// getRealTimeMultiplier 获取实时流量倍率
func (h *TrafficStatisticsHandler) getRealTimeMultiplier() float64 {
	// TODO: 从缓存或配置管理器获取实时倍率
	// 当前返回默认倍率1.0
	return 1.0
}
