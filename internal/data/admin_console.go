package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyorder"
	"github.com/OmnTeam/npanel-pro/ent/proxyserver"
	"github.com/OmnTeam/npanel-pro/ent/proxysystemlog"
	"github.com/OmnTeam/npanel-pro/ent/proxyticket"
	"github.com/OmnTeam/npanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/npanel-pro/ent/proxyuser"
	v1 "github.com/OmnTeam/npanel-pro/internal/biz/admin/console"
	"github.com/OmnTeam/npanel-pro/internal/model"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type adminConsoleRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminConsoleRepo creates a new admin console repository
func NewAdminConsoleRepo(data *Data, logger log.Logger) v1.ConsoleRepo {
	return &adminConsoleRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/admin/console")),
	}
}

// ==================== Revenue Statistics Methods ====================

// QueryDateOrders queries orders by date
func (r *adminConsoleRepo) QueryDateOrders(ctx context.Context, date time.Time) (*v1.OrdersTotal, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(start),
			proxyorder.CreatedAtLTE(end),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryDateOrders failed", "error", err, "date", date)
		return nil, err
	}

	// Calculate totals in Go
	var amountTotal, newOrderAmount, renewalOrderAmount int64
	for _, order := range orders {
		amountTotal += int64(order.Amount)
		if order.IsNew {
			newOrderAmount += int64(order.Amount)
		} else {
			renewalOrderAmount += int64(order.Amount)
		}
	}

	return &v1.OrdersTotal{
		AmountTotal:        int(amountTotal),
		NewOrderAmount:     int(newOrderAmount),
		RenewalOrderAmount: int(renewalOrderAmount),
	}, nil
}

// QueryMonthlyOrders queries orders for a month
func (r *adminConsoleRepo) QueryMonthlyOrders(ctx context.Context, date time.Time) (*v1.OrdersTotal, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Query orders for the month
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLTE(lastDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryMonthlyOrders failed", "error", err, "date", date)
		return nil, err
	}

	// Calculate totals in Go
	var amountTotal, newOrderAmount, renewalOrderAmount int64
	for _, order := range orders {
		amountTotal += int64(order.Amount)
		if order.IsNew {
			newOrderAmount += int64(order.Amount)
		} else {
			renewalOrderAmount += int64(order.Amount)
		}
	}

	return &v1.OrdersTotal{
		AmountTotal:        int(amountTotal),
		NewOrderAmount:     int(newOrderAmount),
		RenewalOrderAmount: int(renewalOrderAmount),
	}, nil
}

// QueryTotalOrders queries total orders
func (r *adminConsoleRepo) QueryTotalOrders(ctx context.Context) (*v1.OrdersTotal, error) {
	// Query all orders
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryTotalOrders failed", "error", err)
		return nil, err
	}

	// Calculate totals in Go
	var amountTotal, newOrderAmount, renewalOrderAmount int64
	for _, order := range orders {
		amountTotal += int64(order.Amount)
		if order.IsNew {
			newOrderAmount += int64(order.Amount)
		} else {
			renewalOrderAmount += int64(order.Amount)
		}
	}

	return &v1.OrdersTotal{
		AmountTotal:        int(amountTotal),
		NewOrderAmount:     int(newOrderAmount),
		RenewalOrderAmount: int(renewalOrderAmount),
	}, nil
}

// QueryDailyOrdersList queries daily orders list for current month
func (r *adminConsoleRepo) QueryDailyOrdersList(ctx context.Context, date time.Time) ([]*v1.OrdersTotalWithDate, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextDay := date.AddDate(0, 0, 1).Truncate(24 * time.Hour)

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryDailyOrdersList failed", "error", err, "date", date)
		return nil, err
	}

	// Group by date and calculate totals in Go
	dailyMap := make(map[string]*v1.OrdersTotalWithDate)
	for _, order := range orders {
		dateStr := order.CreatedAt.Format("2006-01-02")
		if _, exists := dailyMap[dateStr]; !exists {
			dailyMap[dateStr] = &v1.OrdersTotalWithDate{Date: dateStr}
		}

		dailyMap[dateStr].AmountTotal += int(order.Amount)
		if order.IsNew {
			dailyMap[dateStr].NewOrderAmount += int(order.Amount)
		} else {
			dailyMap[dateStr].RenewalOrderAmount += int(order.Amount)
		}
	}

	// Convert map to slice
	result := make([]*v1.OrdersTotalWithDate, 0, len(dailyMap))
	for _, total := range dailyMap {
		result = append(result, total)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// QueryMonthlyOrdersList queries monthly orders list for past 6 months
func (r *adminConsoleRepo) QueryMonthlyOrdersList(ctx context.Context, date time.Time) ([]*v1.OrdersTotalWithDate, error) {
	sixMonthsAgo := date.AddDate(0, -5, 0)
	firstDay := time.Date(sixMonthsAgo.Year(), sixMonthsAgo.Month(), 1, 0, 0, 0, 0, sixMonthsAgo.Location())
	nextDay := date.AddDate(0, 0, 1).Truncate(24 * time.Hour)

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryMonthlyOrdersList failed", "error", err, "date", date)
		return nil, err
	}

	// Group by month and calculate totals in Go
	monthlyMap := make(map[string]*v1.OrdersTotalWithDate)
	for _, order := range orders {
		monthStr := order.CreatedAt.Format("2006-01")
		if _, exists := monthlyMap[monthStr]; !exists {
			monthlyMap[monthStr] = &v1.OrdersTotalWithDate{Date: monthStr}
		}

		monthlyMap[monthStr].AmountTotal += int(order.Amount)
		if order.IsNew {
			monthlyMap[monthStr].NewOrderAmount += int(order.Amount)
		} else {
			monthlyMap[monthStr].RenewalOrderAmount += int(order.Amount)
		}
	}

	// Convert map to slice
	result := make([]*v1.OrdersTotalWithDate, 0, len(monthlyMap))
	for _, total := range monthlyMap {
		result = append(result, total)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// ==================== User Statistics Methods ====================

// QueryRegisterUserTotalByDate queries user registration count by date
func (r *adminConsoleRepo) QueryRegisterUserTotalByDate(ctx context.Context, date time.Time) (int, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)

	count, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.CreatedAtGTE(start),
			proxyuser.CreatedAtLTE(end),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryRegisterUserTotalByDate failed", "error", err, "date", date)
		return 0, err
	}

	return count, nil
}

// QueryRegisterUserTotalByMonthly queries user registration count by month
func (r *adminConsoleRepo) QueryRegisterUserTotalByMonthly(ctx context.Context, date time.Time) (int, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)

	count, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.CreatedAtGTE(firstDay),
			proxyuser.CreatedAtLTE(lastDay),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryRegisterUserTotalByMonthly failed", "error", err, "date", date)
		return 0, err
	}

	return count, nil
}

// QueryRegisterUserTotal queries total user registration count
func (r *adminConsoleRepo) QueryRegisterUserTotal(ctx context.Context) (int, error) {
	count, err := r.data.db.ProxyUser.Query().
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryRegisterUserTotal failed", "error", err)
		return 0, err
	}

	return count, nil
}

// QueryDateUserCounts queries new and renewal user counts by date
func (r *adminConsoleRepo) QueryDateUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(start),
			proxyorder.CreatedAtLTE(end),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryDateUserCounts failed", "error", err, "date", date)
		return 0, 0, err
	}

	// Track unique users
	newUserSet := make(map[int]bool)
	renewalUserSet := make(map[int]bool)

	for _, order := range orders {
		if order.IsNew {
			newUserSet[int(order.UserID)] = true
		} else {
			renewalUserSet[int(order.UserID)] = true
		}
	}

	return int64(len(newUserSet)), int64(len(renewalUserSet)), nil
}

// QueryMonthlyUserCounts queries new and renewal user counts by month
func (r *adminConsoleRepo) QueryMonthlyUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Query orders for the month
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLTE(lastDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryMonthlyUserCounts failed", "error", err, "date", date)
		return 0, 0, err
	}

	// Track unique users
	newUserSet := make(map[int]bool)
	renewalUserSet := make(map[int]bool)

	for _, order := range orders {
		if order.IsNew {
			newUserSet[int(order.UserID)] = true
		} else {
			renewalUserSet[int(order.UserID)] = true
		}
	}

	return int64(len(newUserSet)), int64(len(renewalUserSet)), nil
}

// QueryTotalUserCounts queries total new and renewal user counts
func (r *adminConsoleRepo) QueryTotalUserCounts(ctx context.Context) (newUsers int64, renewalUsers int64, err error) {
	// Query all orders
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryTotalUserCounts failed", "error", err)
		return 0, 0, err
	}

	// Track unique users
	newUserSet := make(map[int]bool)
	renewalUserSet := make(map[int]bool)

	for _, order := range orders {
		if order.IsNew {
			newUserSet[int(order.UserID)] = true
		} else {
			renewalUserSet[int(order.UserID)] = true
		}
	}

	return int64(len(newUserSet)), int64(len(renewalUserSet)), nil
}

// QueryDailyUserStatisticsList queries daily user statistics list for current month
func (r *adminConsoleRepo) QueryDailyUserStatisticsList(ctx context.Context, date time.Time) ([]*v1.UserStatistics, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextDay := date.AddDate(0, 0, 1).Truncate(24 * time.Hour)

	// Query user registrations for the date range
	users, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.CreatedAtGTE(firstDay),
			proxyuser.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryDailyUserStatisticsList registration failed", "error", err, "date", date)
		return nil, err
	}

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryDailyUserStatisticsList orders failed", "error", err, "date", date)
		return nil, err
	}

	// Calculate statistics in Go
	statsMap := make(map[string]*v1.UserStatistics)

	// Process user registrations
	for _, user := range users {
		dateStr := user.CreatedAt.Format("2006-01-02")
		if _, exists := statsMap[dateStr]; !exists {
			statsMap[dateStr] = &v1.UserStatistics{Date: dateStr}
		}
		statsMap[dateStr].Register++
	}

	// Process orders
	newOrderSeen := make(map[string]map[int64]struct{})
	renewalOrderSeen := make(map[string]map[int64]struct{})
	for _, order := range orders {
		dateStr := order.CreatedAt.Format("2006-01-02")
		if _, exists := statsMap[dateStr]; !exists {
			continue
		}
		key := int64(order.UserID)
		if order.IsNew {
			if _, ok := newOrderSeen[dateStr]; !ok {
				newOrderSeen[dateStr] = make(map[int64]struct{})
			}
			if _, ok := newOrderSeen[dateStr][key]; ok {
				continue
			}
			newOrderSeen[dateStr][key] = struct{}{}
			statsMap[dateStr].NewOrderUsers++
		} else {
			if _, ok := renewalOrderSeen[dateStr]; !ok {
				renewalOrderSeen[dateStr] = make(map[int64]struct{})
			}
			if _, ok := renewalOrderSeen[dateStr][key]; ok {
				continue
			}
			renewalOrderSeen[dateStr][key] = struct{}{}
			statsMap[dateStr].RenewalOrderUsers++
		}
	}

	// Convert map to slice
	result := make([]*v1.UserStatistics, 0, len(statsMap))
	for _, stat := range statsMap {
		result = append(result, stat)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// QueryMonthlyUserStatisticsList queries monthly user statistics list for past 6 months
func (r *adminConsoleRepo) QueryMonthlyUserStatisticsList(ctx context.Context, date time.Time) ([]*v1.UserStatistics, error) {
	sixMonthsAgo := date.AddDate(0, -5, 0)
	firstDay := time.Date(sixMonthsAgo.Year(), sixMonthsAgo.Month(), 1, 0, 0, 0, 0, sixMonthsAgo.Location())
	nextDay := date.AddDate(0, 0, 1).Truncate(24 * time.Hour)

	// Query user registrations for the date range
	users, err := r.data.db.ProxyUser.Query().
		Where(
			proxyuser.CreatedAtGTE(firstDay),
			proxyuser.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryMonthlyUserStatisticsList registration failed", "error", err, "date", date)
		return nil, err
	}

	// Query orders for the date range
	orders, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.StatusIn(2, 5),
			proxyorder.MethodNEQ("balance"),
			proxyorder.CreatedAtGTE(firstDay),
			proxyorder.CreatedAtLT(nextDay),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryMonthlyUserStatisticsList orders failed", "error", err, "date", date)
		return nil, err
	}

	// Calculate statistics in Go
	statsMap := make(map[string]*v1.UserStatistics)

	// Process user registrations
	for _, user := range users {
		monthStr := user.CreatedAt.Format("2006-01")
		if _, exists := statsMap[monthStr]; !exists {
			statsMap[monthStr] = &v1.UserStatistics{Date: monthStr}
		}
		statsMap[monthStr].Register++
	}

	// Process orders
	newOrderSeen := make(map[string]map[int64]struct{})
	renewalOrderSeen := make(map[string]map[int64]struct{})
	for _, order := range orders {
		monthStr := order.CreatedAt.Format("2006-01")
		if _, exists := statsMap[monthStr]; !exists {
			continue
		}
		key := int64(order.UserID)
		if order.IsNew {
			if _, ok := newOrderSeen[monthStr]; !ok {
				newOrderSeen[monthStr] = make(map[int64]struct{})
			}
			if _, ok := newOrderSeen[monthStr][key]; ok {
				continue
			}
			newOrderSeen[monthStr][key] = struct{}{}
			statsMap[monthStr].NewOrderUsers++
		} else {
			if _, ok := renewalOrderSeen[monthStr]; !ok {
				renewalOrderSeen[monthStr] = make(map[int64]struct{})
			}
			if _, ok := renewalOrderSeen[monthStr][key]; ok {
				continue
			}
			renewalOrderSeen[monthStr][key] = struct{}{}
			statsMap[monthStr].RenewalOrderUsers++
		}
	}

	// Convert map to slice
	result := make([]*v1.UserStatistics, 0, len(statsMap))
	for _, stat := range statsMap {
		result = append(result, stat)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// ==================== Ticket Statistics Methods ====================

// QueryWaitReplyTotal queries waiting reply ticket count
func (r *adminConsoleRepo) QueryWaitReplyTotal(ctx context.Context) (int, error) {
	count, err := r.data.db.ProxyTicket.Query().
		Where(
			proxyticket.StatusEQ(1), // 1 = Pending (waiting for reply)
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryWaitReplyTotal failed", "error", err)
		return 0, err
	}

	return count, nil
}

// ==================== Server Statistics Methods ====================

// QueryOnlineServers queries online server count
func (r *adminConsoleRepo) QueryOnlineServers(ctx context.Context) (int, error) {
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	count, err := r.data.db.ProxyServer.Query().
		Where(
			proxyserver.LastReportedAtGT(fiveMinutesAgo),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryOnlineServers failed", "error", err)
		return 0, err
	}

	return count, nil
}

// QueryOfflineServers queries offline server count
func (r *adminConsoleRepo) QueryOfflineServers(ctx context.Context) (int, error) {
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	count, err := r.data.db.ProxyServer.Query().
		Where(
			proxyserver.Or(
				proxyserver.LastReportedAtLTE(fiveMinutesAgo),
				proxyserver.LastReportedAtIsNil(),
			),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("QueryOfflineServers failed", "error", err)
		return 0, err
	}

	return count, nil
}

// QueryOnlineUsers queries online user count
func (r *adminConsoleRepo) QueryOnlineUsers(ctx context.Context) (int, error) {
	now := time.Now().Unix()
	if err := r.data.rdb.ZRemRangeByScore(ctx, OnlineUserSubscribeCacheKeyWithGlobal, "-inf", fmt.Sprintf("%d", now)).Err(); err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	count, err := r.data.rdb.ZCard(ctx, OnlineUserSubscribeCacheKeyWithGlobal).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	if count == 0 {
		fallbackCount, fallbackErr := r.queryOnlineUsersFromServerCaches(ctx)
		if fallbackErr != nil {
			return 0, fallbackErr
		}
		return fallbackCount, nil
	}
	return int(count), nil
}

func (r *adminConsoleRepo) queryOnlineUsersFromServerCaches(ctx context.Context) (int, error) {
	var (
		cursor      uint64
		subscribeID = make(map[int64]struct{})
	)

	for {
		keys, nextCursor, err := r.data.rdb.Scan(ctx, cursor, "node:online:subscribe:*", 200).Result()
		if err != nil {
			if err == redis.Nil {
				return 0, nil
			}
			return 0, err
		}
		for _, key := range keys {
			if key == OnlineUserSubscribeCacheKeyWithGlobal {
				continue
			}
			raw, err := r.data.rdb.Get(ctx, key).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				return 0, err
			}
			if raw == "" {
				continue
			}
			var onlineUsers map[int64][]string
			if err := json.Unmarshal([]byte(raw), &onlineUsers); err != nil {
				continue
			}
			for sid := range onlineUsers {
				subscribeID[sid] = struct{}{}
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return len(subscribeID), nil
}

// QueryTodayTraffic queries today's traffic
func (r *adminConsoleRepo) QueryTodayTraffic(ctx context.Context, date time.Time) (upload int64, download int64, err error) {
	todayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)

	// Query traffic logs for today
	trafficLogs, err := r.data.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(todayStart),
			proxytrafficlog.TimestampLTE(todayEnd),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryTodayTraffic failed", "error", err, "date", date)
		return 0, 0, err
	}

	// Calculate totals in Go
	var totalUpload, totalDownload int64
	for _, log := range trafficLogs {
		totalUpload += int64(log.Upload)
		totalDownload += int64(log.Download)
	}

	return totalUpload, totalDownload, nil
}

// QueryMonthlyTraffic queries monthly traffic
func (r *adminConsoleRepo) QueryMonthlyTraffic(ctx context.Context, date time.Time) (upload int64, download int64, err error) {
	// Get today's traffic
	todayUpload, todayDownload, err := r.QueryTodayTraffic(ctx, date)
	if err != nil {
		return 0, 0, err
	}

	// Get traffic statistics from system log for the current month (excluding today)
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	yesterdayEnd := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Add(-time.Second)

	logs, err := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(model.TypeTrafficStat)),
			proxysystemlog.DateGTE(firstDay.Format(time.DateOnly)),
			proxysystemlog.DateLTE(yesterdayEnd.Format(time.DateOnly)),
		).
		All(ctx)

	if err != nil {
		r.log.Errorw("QueryMonthlyTraffic failed to query system log", "error", err, "date", date)
		// Don't return error, just use today's traffic
		return todayUpload, todayDownload, nil
	}

	var totalUpload, totalDownload int64 = todayUpload, todayDownload

	for _, logEntry := range logs {
		var trafficStat model.TrafficStat
		if err := json.Unmarshal([]byte(logEntry.Content), &trafficStat); err != nil {
			r.log.Warnw("Failed to unmarshal traffic stat", "error", err, "log_id", logEntry.ID)
			continue
		}
		totalUpload += trafficStat.Upload
		totalDownload += trafficStat.Download
	}

	return totalUpload, totalDownload, nil
}

// QueryTodayUserTrafficRanking queries today's user traffic ranking
func (r *adminConsoleRepo) QueryTodayUserTrafficRanking(ctx context.Context, date time.Time) ([]*v1.UserTrafficData, error) {
	todayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)

	// Query traffic logs for today
	trafficLogs, err := r.data.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(todayStart),
			proxytrafficlog.TimestampLTE(todayEnd),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryTodayUserTrafficRanking failed", "error", err, "date", date)
		return nil, err
	}

	// Group by user and calculate totals in Go
	userTrafficMap := make(map[int64]*v1.UserTrafficData)
	for _, log := range trafficLogs {
		subscribeID := log.SubscribeID
		if _, exists := userTrafficMap[subscribeID]; !exists {
			userTrafficMap[subscribeID] = &v1.UserTrafficData{
				SID:      int(subscribeID),
				Upload:   0,
				Download: 0,
			}
		}
		userTrafficMap[subscribeID].Upload += int(log.Upload)
		userTrafficMap[subscribeID].Download += int(log.Download)
	}

	// Convert to slice and sort by total traffic
	result := make([]*v1.UserTrafficData, 0, len(userTrafficMap))
	for _, data := range userTrafficMap {
		result = append(result, data)
	}

	sort.Slice(result, func(i, j int) bool {
		totalI := result[i].Upload + result[i].Download
		totalJ := result[j].Upload + result[j].Download
		return totalI > totalJ
	})

	// Take top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result, nil
}

// QueryYesterdayUserTrafficRanking queries yesterday's user traffic ranking from system log
func (r *adminConsoleRepo) QueryYesterdayUserTrafficRanking(ctx context.Context, date time.Time) ([]*v1.UserTrafficData, error) {
	yesterday := date.AddDate(0, 0, -1)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	log, err := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(model.TypeUserTrafficRank)),
			proxysystemlog.DateEQ(yesterdayStart.Format(time.DateOnly)),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// No ranking data for yesterday
			return []*v1.UserTrafficData{}, nil
		}
		r.log.Errorw("QueryYesterdayUserTrafficRanking failed", "error", err, "date", date)
		return nil, err
	}

	var ranking model.UserTrafficRank
	if err := json.Unmarshal([]byte(log.Content), &ranking); err != nil {
		r.log.Errorw("Failed to unmarshal user traffic ranking", "error", err, "log_id", log.ID)
		return nil, err
	}

	result := make([]*v1.UserTrafficData, 0, len(ranking.Rank))
	for _, item := range ranking.Rank {
		result = append(result, &v1.UserTrafficData{
			SID:      int(item.SubscribeID),
			Upload:   int(item.Upload),
			Download: int(item.Download),
		})
	}

	return result, nil
}

// QueryTodayServerTrafficRanking queries today's server traffic ranking
func (r *adminConsoleRepo) QueryTodayServerTrafficRanking(ctx context.Context, date time.Time) ([]*v1.ServerTrafficData, error) {
	todayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)

	// Query traffic logs for today
	trafficLogs, err := r.data.db.ProxyTrafficLog.Query().
		Where(
			proxytrafficlog.TimestampGTE(todayStart),
			proxytrafficlog.TimestampLTE(todayEnd),
		).
		All(ctx)
	if err != nil {
		r.log.Errorw("QueryTodayServerTrafficRanking failed", "error", err, "date", date)
		return nil, err
	}

	// Group by server and calculate totals in Go
	serverTrafficMap := make(map[int64]*v1.ServerTrafficData)
	for _, log := range trafficLogs {
		serverID := log.ServerID
		if _, exists := serverTrafficMap[serverID]; !exists {
			serverTrafficMap[serverID] = &v1.ServerTrafficData{
				ServerID: int(serverID),
				Name:     fmt.Sprintf("Server %d", serverID),
				Upload:   0,
				Download: 0,
			}
		}
		serverTrafficMap[serverID].Upload += int(log.Upload)
		serverTrafficMap[serverID].Download += int(log.Download)
	}

	// Get server names
	for serverID, data := range serverTrafficMap {
		server, err := r.data.db.ProxyServer.Query().
			Where(
				proxyserver.IDEQ(serverID),
			).
			Only(ctx)

		if err != nil {
			r.log.Warnw("Failed to get server name", "error", err, "server_id", serverID)
			// Keep the default name
		} else {
			data.Name = server.Name
		}
	}

	// Convert to slice and sort by total traffic
	result := make([]*v1.ServerTrafficData, 0, len(serverTrafficMap))
	for _, data := range serverTrafficMap {
		result = append(result, data)
	}

	sort.Slice(result, func(i, j int) bool {
		totalI := result[i].Upload + result[i].Download
		totalJ := result[j].Upload + result[j].Download
		return totalI > totalJ
	})

	// Take top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result, nil
}

// QueryYesterdayServerTrafficRanking queries yesterday's server traffic ranking from system log
func (r *adminConsoleRepo) QueryYesterdayServerTrafficRanking(ctx context.Context, date time.Time) ([]*v1.ServerTrafficData, error) {
	yesterday := date.AddDate(0, 0, -1)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	log, err := r.data.db.ProxySystemLog.Query().
		Where(
			proxysystemlog.TypeEQ(int8(model.TypeServerTrafficRank)),
			proxysystemlog.DateEQ(yesterdayStart.Format(time.DateOnly)),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// No ranking data for yesterday
			return []*v1.ServerTrafficData{}, nil
		}
		r.log.Errorw("QueryYesterdayServerTrafficRanking failed", "error", err, "date", date)
		return nil, err
	}

	var ranking model.ServerTrafficRank
	if err := json.Unmarshal([]byte(log.Content), &ranking); err != nil {
		r.log.Errorw("Failed to unmarshal server traffic ranking", "error", err, "log_id", log.ID)
		return nil, err
	}

	result := make([]*v1.ServerTrafficData, 0, len(ranking.Rank))
	for _, item := range ranking.Rank {
		// Get server name
		server, err := r.data.db.ProxyServer.Query().
			Where(
				proxyserver.IDEQ(item.ServerID),
			).
			Only(ctx)

		var name string
		if err != nil {
			r.log.Warnw("Failed to get server name", "error", err, "server_id", item.ServerID)
			name = fmt.Sprintf("Server %d", item.ServerID)
		} else {
			name = server.Name
		}

		result = append(result, &v1.ServerTrafficData{
			ServerID: int(item.ServerID),
			Name:     name,
			Upload:   int(item.Upload),
			Download: int(item.Download),
		})
	}

	return result, nil
}
