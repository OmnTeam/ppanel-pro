package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ServerTrafficData 服务器流量数据
type ServerTrafficData struct {
	ServerID int64
	Name     string
	Upload   int64
	Download int64
}

// UserTrafficData 用户流量数据
type UserTrafficData struct {
	UserID   int64
	Upload   int64
	Download int64
}

// ServerDataHandler 服务器数据统计处理器
// 用于统计服务器和用户的流量排行榜数据，并缓存到Redis
type ServerDataHandler struct {
	db     *ent.Client
	rdb    *redis.Client
	logger *log.Helper
}

// NewServerDataHandler creates a new server data handler
func NewServerDataHandler(db *ent.Client, rdb *redis.Client, logger log.Logger) *ServerDataHandler {
	return &ServerDataHandler{
		db:     db,
		rdb:    rdb,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask processes the server data statistics task
func (h *ServerDataHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	now := time.Now()

	// 1. 获取服务器流量排行榜（Top 10）
	top10ServerToday, top10ServerYesterday, top10UserToday, top10UserYesterday := h.getRanking(ctx, now)

	// 2. 获取总流量统计
	totalUploadToday, totalDownloadToday, totalDownloadMonthly, totalUploadMonthly := h.trafficCount(ctx, now)

	// 3. 构建服务器数据
	serverData := map[string]interface{}{
		"server_traffic_ranking_today":     top10ServerToday,
		"server_traffic_ranking_yesterday": top10ServerYesterday,
		"user_traffic_ranking_today":       top10UserToday,
		"user_traffic_ranking_yesterday":   top10UserYesterday,
		"today_upload":                     totalUploadToday,
		"today_download":                   totalDownloadToday,
		"monthly_upload":                   totalUploadMonthly,
		"monthly_download":                 totalDownloadMonthly,
		"updated_at":                       now.UnixMilli(),
	}

	// 4. 序列化数据
	data, err := json.Marshal(serverData)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[ServerDataHandler] Failed to marshal server data: %v", err)
		return err
	}

	// 5. 缓存到Redis
	cacheKey := "server:count:data"
	if err := h.rdb.Set(ctx, cacheKey, data, 0).Err(); err != nil {
		h.logger.WithContext(ctx).Errorf("[ServerDataHandler] Failed to cache server data: %v", err)
		return err
	}

	h.logger.WithContext(ctx).Infof("[ServerDataHandler] Server data statistics task completed successfully")
	return nil
}

// getRanking 获取流量排行榜
func (h *ServerDataHandler) getRanking(ctx context.Context, now time.Time) (
	top10ServerToday, top10ServerYesterday []*ServerTrafficData,
	top10UserToday, top10UserYesterday []*UserTrafficData,
) {
	// 获取今天的服务器流量排行榜
	top10ServerToday = h.topServersTrafficByDay(ctx, now, 10)

	// 获取昨天的服务器流量排行榜
	top10ServerYesterday = h.topServersTrafficByDay(ctx, now.AddDate(0, 0, -1), 10)

	// 获取今天的用户流量排行榜
	top10UserToday = h.topUsersTrafficByDay(ctx, now, 10)

	// 获取昨天的用户流量排行榜
	top10UserYesterday = h.topUsersTrafficByDay(ctx, now.AddDate(0, 0, -1), 10)

	return
}

// trafficCount 统计总流量
func (h *ServerDataHandler) trafficCount(ctx context.Context, now time.Time) (
	totalUploadToday, totalDownloadToday, totalUploadMonthly, totalDownloadMonthly int64,
) {
	// 今天的流量统计
	todayTotal := h.queryTrafficByDay(ctx, now)
	totalUploadToday = todayTotal.Upload
	totalDownloadToday = todayTotal.Download

	// 本月的流量统计
	monthlyTotal := h.queryTrafficByMonthly(ctx, now)
	totalUploadMonthly = monthlyTotal.Upload
	totalDownloadMonthly = monthlyTotal.Download

	return
}

// topServersTrafficByDay 按天查询Top N服务器流量
func (h *ServerDataHandler) topServersTrafficByDay(ctx context.Context, day time.Time, limit int) []*ServerTrafficData {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24*time.Hour - time.Nanosecond)

	// 查询所有符合条件的流量日志
	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[topServersTrafficByDay] Failed to query traffic logs: %v", err)
		return []*ServerTrafficData{}
	}

	// 按服务器ID聚合
	serverMap := make(map[int64]*ServerTrafficData)
	for _, log := range logs {
		if log.ServerID == 0 {
			continue
		}
		if _, exists := serverMap[log.ServerID]; !exists {
			serverMap[log.ServerID] = &ServerTrafficData{
				ServerID: log.ServerID,
				Upload:   0,
				Download: 0,
			}
		}
		serverMap[log.ServerID].Upload += int64(log.Upload)
		serverMap[log.ServerID].Download += int64(log.Download)
	}

	// 查询服务器名称并构建结果
	var results []*ServerTrafficData
	for _, data := range serverMap {
		// 查询服务器信息获取名称
		server, err := h.db.ProxyNode.Query().
			Where(proxynode.IDEQ(data.ServerID)).
			First(ctx)
		if err != nil {
			h.logger.WithContext(ctx).Warnf("[topServersTrafficByDay] Server not found: %d", data.ServerID)
			data.Name = "Unknown"
		} else {
			data.Name = server.Name
		}
		results = append(results, data)
	}

	// 按总流量排序（降序）
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Upload+results[i].Download < results[j].Upload+results[j].Download {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制返回数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// topUsersTrafficByDay 按天查询Top N用户流量
func (h *ServerDataHandler) topUsersTrafficByDay(ctx context.Context, day time.Time, limit int) []*UserTrafficData {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24*time.Hour - time.Nanosecond)

	// 查询所有符合条件的流量日志
	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[topUsersTrafficByDay] Failed to query traffic logs: %v", err)
		return []*UserTrafficData{}
	}

	// 按用户ID聚合
	userMap := make(map[int64]*UserTrafficData)
	for _, log := range logs {
		if log.UserID == 0 {
			continue
		}
		if _, exists := userMap[log.UserID]; !exists {
			userMap[log.UserID] = &UserTrafficData{
				UserID:   log.UserID,
				Upload:   0,
				Download: 0,
			}
		}
		userMap[log.UserID].Upload += int64(log.Upload)
		userMap[log.UserID].Download += int64(log.Download)
	}

	// 构建结果
	var results []*UserTrafficData
	for _, data := range userMap {
		results = append(results, data)
	}

	// 按总流量排序（降序）
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Upload+results[i].Download < results[j].Upload+results[j].Download {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制返回数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// TrafficTotal 流量总计
type TrafficTotal struct {
	Upload   int64
	Download int64
}

// queryTrafficByDay 查询指定日期的总流量
func (h *ServerDataHandler) queryTrafficByDay(ctx context.Context, day time.Time) *TrafficTotal {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24*time.Hour - time.Nanosecond)

	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[queryTrafficByDay] Failed to query traffic logs: %v", err)
		return &TrafficTotal{}
	}

	total := &TrafficTotal{}
	for _, log := range logs {
		total.Upload += int64(log.Upload)
		total.Download += int64(log.Download)
	}

	return total
}

// queryTrafficByMonthly 查询指定月份的总流量
func (h *ServerDataHandler) queryTrafficByMonthly(ctx context.Context, date time.Time) *TrafficTotal {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	logs, err := h.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(start),
			proxytrafficlog.TimestampLTE(end),
		).
		All(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[queryTrafficByMonthly] Failed to query traffic logs: %v", err)
		return &TrafficTotal{}
	}

	total := &TrafficTotal{}
	for _, log := range logs {
		total.Upload += int64(log.Upload)
		total.Download += int64(log.Download)
	}

	return total
}
