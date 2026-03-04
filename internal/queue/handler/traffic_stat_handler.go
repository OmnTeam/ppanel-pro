package handler

import (
	"context"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// TrafficStatHandler 流量统计处理器
// 用于每日流量统计，生成用户和服务器流量排行榜
type TrafficStatHandler struct {
	db     *ent.Client
	rdb    *redis.Client
	logger *log.Helper
}

// NewTrafficStatHandler creates a new traffic statistics handler
func NewTrafficStatHandler(db *ent.Client, rdb *redis.Client, logger log.Logger) *TrafficStatHandler {
	return &TrafficStatHandler{
		db:     db,
		rdb:    rdb,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask processes the traffic statistics task
func (h *TrafficStatHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	now := time.Now()

	// 获取昨天的日期范围（统计昨天的流量）
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.Local)
	end := start.Add(24*time.Hour - time.Nanosecond)
	date := start.Format(time.DateOnly)

	h.logger.WithContext(ctx).Infof("[TrafficStatHandler] Processing traffic statistics for date: %s (from %s to %s)",
		date, start.Format(time.RFC3339), end.Format(time.RFC3339))

	// 1. 查询用户流量统计并生成排行榜
	userTop10, err := h.processUserTraffic(ctx, start, end, date)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatHandler] Failed to process user traffic: %v", err)
		// Continue processing
	}

	// 2. 查询服务器流量统计并生成排行榜
	serverTop10, err := h.processServerTraffic(ctx, start, end, date)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatHandler] Failed to process server traffic: %v", err)
		// Continue processing
	}

	// 3. 查询总体流量统计
	totalTraffic, err := h.processTotalTraffic(ctx, start, end, date)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatHandler] Failed to process total traffic: %v", err)
		// Continue processing
	}

	// 4. 缓存统计数据到Redis
	h.cacheStatistics(ctx, date, userTop10, serverTop10, totalTraffic)

	// 5. 清理旧的流量日志（可选，根据配置）
	// h.cleanOldTrafficLogs(ctx, end, 7) // 保留7天

	h.logger.WithContext(ctx).Infof("[TrafficStatHandler] Traffic statistics task completed: date=%s, users=%d, servers=%d, total_upload=%d, total_download=%d",
		date, len(userTop10), len(serverTop10), totalTraffic.Upload, totalTraffic.Download)

	return nil
}

// UserTrafficStat 用户流量统计
type UserTrafficStat struct {
	SubscribeID int64
	UserID      int64
	Upload      int64
	Download    int64
	Total       int64
}

// ServerTrafficStat 服务器流量统计
type ServerTrafficStat struct {
	ServerID int64
	Upload   int64
	Download int64
	Total    int64
}

// TotalTrafficStat 总流量统计
type TotalTrafficStat struct {
	Upload   int64
	Download int64
	Total    int64
}

// processUserTraffic 处理用户流量统计
func (h *TrafficStatHandler) processUserTraffic(ctx context.Context, start, end time.Time, date string) ([]*UserTrafficStat, error) {
	// 查询所有符合条件的流量日志
	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 按订阅ID聚合
	subscribeMap := make(map[int64]*UserTrafficStat)
	for _, log := range logs {
		if log.SubscribeID == 0 {
			continue
		}
		if _, exists := subscribeMap[log.SubscribeID]; !exists {
			subscribeMap[log.SubscribeID] = &UserTrafficStat{
				SubscribeID: log.SubscribeID,
				UserID:      log.UserID,
				Upload:      0,
				Download:    0,
			}
		}
		subscribeMap[log.SubscribeID].Upload += int64(log.Upload)
		subscribeMap[log.SubscribeID].Download += int64(log.Download)
		subscribeMap[log.SubscribeID].Total = subscribeMap[log.SubscribeID].Upload + subscribeMap[log.SubscribeID].Download
	}

	// 构建结果并排序
	var results []*UserTrafficStat
	for _, stat := range subscribeMap {
		results = append(results, stat)
	}

	// 按总流量排序（降序）
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Total < results[j].Total {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 只返回Top 10
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}

// processServerTraffic 处理服务器流量统计
func (h *TrafficStatHandler) processServerTraffic(ctx context.Context, start, end time.Time, date string) ([]*ServerTrafficStat, error) {
	// 查询所有符合条件的流量日志
	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 按服务器ID聚合
	serverMap := make(map[int64]*ServerTrafficStat)
	for _, log := range logs {
		if log.ServerID == 0 {
			continue
		}
		if _, exists := serverMap[log.ServerID]; !exists {
			serverMap[log.ServerID] = &ServerTrafficStat{
				ServerID: log.ServerID,
				Upload:   0,
				Download: 0,
			}
		}
		serverMap[log.ServerID].Upload += int64(log.Upload)
		serverMap[log.ServerID].Download += int64(log.Download)
		serverMap[log.ServerID].Total = serverMap[log.ServerID].Upload + serverMap[log.ServerID].Download
	}

	// 构建结果并排序
	var results []*ServerTrafficStat
	for _, stat := range serverMap {
		results = append(results, stat)
	}

	// 按总流量排序（降序）
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Total < results[j].Total {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 只返回Top 10
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}

// processTotalTraffic 处理总体流量统计
func (h *TrafficStatHandler) processTotalTraffic(ctx context.Context, start, end time.Time, date string) (*TotalTrafficStat, error) {
	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	total := &TotalTrafficStat{}
	for _, log := range logs {
		total.Upload += int64(log.Upload)
		total.Download += int64(log.Download)
	}
	total.Total = total.Upload + total.Download

	return total, nil
}

// cacheStatistics 缓存统计数据到Redis
func (h *TrafficStatHandler) cacheStatistics(ctx context.Context, date string, userTop10 []*UserTrafficStat, serverTop10 []*ServerTrafficStat, totalTraffic *TotalTrafficStat) {
	// 缓存用户流量排行榜
	userRankKey := "traffic:stat:user:" + date
	for i, stat := range userTop10 {
		field := string(rune(i + 1))
		h.rdb.HSet(ctx, userRankKey, field, stat)
	}
	h.rdb.Expire(ctx, userRankKey, 30*24*time.Hour) // 保留30天

	// 缓存服务器流量排行榜
	serverRankKey := "traffic:stat:server:" + date
	for i, stat := range serverTop10 {
		field := string(rune(i + 1))
		h.rdb.HSet(ctx, serverRankKey, field, stat)
	}
	h.rdb.Expire(ctx, serverRankKey, 30*24*time.Hour) // 保留30天

	// 缓存总流量
	totalKey := "traffic:stat:total:" + date
	h.rdb.HSet(ctx, totalKey, "upload", totalTraffic.Upload)
	h.rdb.HSet(ctx, totalKey, "download", totalTraffic.Download)
	h.rdb.HSet(ctx, totalKey, "total", totalTraffic.Total)
	h.rdb.Expire(ctx, totalKey, 30*24*time.Hour) // 保留30天
}

// cleanOldTrafficLogs 清理旧的流量日志
func (h *TrafficStatHandler) cleanOldTrafficLogs(ctx context.Context, beforeDate time.Time, retainDays int) {
	cutoffDate := beforeDate.AddDate(0, 0, -retainDays)

	deletedCount, err := h.db.ProxyTrafficLog.Delete().
		Where(proxytrafficlog.TimestampLTE(cutoffDate)).
		Exec(ctx)

	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatHandler] Failed to delete old traffic logs: %v", err)
		return
	}

	h.logger.WithContext(ctx).Infof("[TrafficStatHandler] Cleaned old traffic logs: deleted=%d, before=%s",
		deletedCount, cutoffDate.Format(time.RFC3339))
}
