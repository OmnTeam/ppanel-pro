package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyserver"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	"github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type TrafficStatisticsHandler struct {
	db     *ent.Client
	logger *log.Helper
}

type trafficMultiplierPeriod struct {
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Multiplier float64 `json:"multiplier"`
}

func NewTrafficStatisticsHandler(db *ent.Client, logger log.Logger) *TrafficStatisticsHandler {
	return &TrafficStatisticsHandler{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

func (h *TrafficStatisticsHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload types.TrafficStatistics
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to unmarshal payload: %v", err)
		return nil
	}
	if len(payload.Logs) == 0 {
		h.logger.WithContext(ctx).Warnf("[TrafficStatisticsHandler] Empty payload received for server_id: %d", payload.ServerID)
		return nil
	}

	serverInfo, err := h.db.ProxyServer.Query().
		Where(proxyserver.IDEQ(payload.ServerID)).
		First(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Server not found: %d, error: %v", payload.ServerID, err)
		return nil
	}

	ratio, ok := h.getProtocolRatio(serverInfo, payload.Protocol)
	if !ok {
		h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Protocol not found: %s", payload.Protocol)
		return nil
	}

	now := time.Now()
	threshold := h.getTrafficReportThreshold(ctx)
	realTimeMultiplier := h.getRealTimeMultiplier(ctx, now)
	processedCount := 0
	skippedCount := 0

	for _, logEntry := range payload.Logs {
		sub, err := h.db.ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.IDEQ(logEntry.SID)).
			First(ctx)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Subscribe not found: %d", logEntry.SID)
			continue
		}
		if logEntry.Download+logEntry.Upload <= threshold {
			skippedCount++
			continue
		}

		d := int64(float64(logEntry.Download) * ratio * realTimeMultiplier)
		u := int64(float64(logEntry.Upload) * ratio * realTimeMultiplier)
		isExpired := sub.ExpireTime != nil && now.After(*sub.ExpireTime)

		update := h.db.ProxyUserSubscribe.UpdateOneID(sub.ID)
		if isExpired {
			update = update.AddExpiredDownload(d).AddExpiredUpload(u)
		} else {
			update = update.AddDownload(d).AddUpload(u)
		}
		if err := update.Exec(ctx); err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to update subscribe: %d, error: %v", logEntry.SID, err)
			continue
		}

		if _, err := h.db.ProxyTrafficLog.Create().
			SetServerID(payload.ServerID).
			SetSubscribeID(logEntry.SID).
			SetUserID(sub.UserID).
			SetUpload(u).
			SetDownload(d).
			SetTimestamp(now).
			Save(ctx); err != nil {
			h.logger.WithContext(ctx).Errorf("[TrafficStatisticsHandler] Failed to create traffic log: %v", err)
		}

		processedCount++
	}

	h.logger.WithContext(ctx).Infof("[TrafficStatisticsHandler] Traffic statistics task completed: processed=%d, skipped=%d",
		processedCount, skippedCount)
	return nil
}

func (h *TrafficStatisticsHandler) getProtocolRatio(server *ent.ProxyServer, protocol string) (float64, bool) {
	protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
	if err != nil {
		h.logger.Errorf("[TrafficStatisticsHandler] Failed to unmarshal protocols: %v", err)
		return 1.0, false
	}

	for _, item := range protocols {
		if item == nil || strings.ToLower(item.Type) != strings.ToLower(protocol) {
			continue
		}
		if item.Ratio > 0 {
			return item.Ratio, true
		}
		return 1.0, true
	}
	return 1.0, false
}

func (h *TrafficStatisticsHandler) getTrafficReportThreshold(ctx context.Context) int64 {
	value, err := h.systemValue(ctx, "TrafficReportThreshold", "traffic_report_threshold")
	if err != nil || strings.TrimSpace(value) == "" {
		return 0
	}
	threshold, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return threshold
}

func (h *TrafficStatisticsHandler) getRealTimeMultiplier(ctx context.Context, now time.Time) float64 {
	value, err := h.systemValue(ctx, "NodeMultiplierConfig", "node_multiplier_config", "NodeMultiplier")
	if err != nil || strings.TrimSpace(value) == "" {
		return 1.0
	}

	var periods []trafficMultiplierPeriod
	if err := json.Unmarshal([]byte(value), &periods); err != nil {
		return 1.0
	}
	for _, period := range periods {
		if compatTimeWithinPeriod(now, period.StartTime, period.EndTime) {
			if period.Multiplier > 0 {
				return period.Multiplier
			}
			return 1.0
		}
	}
	return 1.0
}

func (h *TrafficStatisticsHandler) systemValue(ctx context.Context, keys ...string) (string, error) {
	for _, key := range keys {
		item, err := h.db.ProxySystem.Query().
			Where(
				proxysystem.CategoryEQ("server"),
				proxysystem.KeyEQ(key),
			).
			First(ctx)
		if err == nil {
			return item.Value, nil
		}
	}
	return "", nil
}

func compatTimeWithinPeriod(current time.Time, start, end string) bool {
	startTime, err := time.Parse("15:04.000", start)
	if err != nil {
		return false
	}
	endTime, err := time.Parse("15:04.000", end)
	if err != nil {
		return false
	}

	currentTime := time.Date(0, 1, 1, current.Hour(), current.Minute(), 0, 0, time.UTC)
	startFormatted := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endFormatted := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	if startFormatted.Before(endFormatted) {
		return currentTime.After(startFormatted) && currentTime.Before(endFormatted)
	}
	return currentTime.After(startFormatted) || currentTime.Before(endFormatted)
}
