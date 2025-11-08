package console

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// OrdersTotal represents order statistics totals
type OrdersTotal struct {
	AmountTotal        int64                  `json:"amount_total"`
	NewOrderAmount     int64                  `json:"new_order_amount"`
	RenewalOrderAmount int64                  `json:"renewal_order_amount"`
	List               []*OrdersTotalWithDate `json:"list,omitempty"`
}

// OrdersTotalWithDate represents order statistics with date
type OrdersTotalWithDate struct {
	Date               string `json:"date"`
	AmountTotal        int64  `json:"amount_total"`
	NewOrderAmount     int64  `json:"new_order_amount"`
	RenewalOrderAmount int64  `json:"renewal_order_amount"`
}

// UserStatistics represents user statistics
type UserStatistics struct {
	Date              string            `json:"date,omitempty"`
	Register          int64             `json:"register"`
	NewOrderUsers     int64             `json:"new_order_users"`
	RenewalOrderUsers int64             `json:"renewal_order_users"`
	List              []*UserStatistics `json:"list,omitempty"`
}

// UserTrafficData represents user traffic ranking data
type UserTrafficData struct {
	SID      int64 `json:"sid"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

// ServerTrafficData represents server traffic ranking data
type ServerTrafficData struct {
	ServerID int64  `json:"server_id"`
	Name     string `json:"name"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// ConsoleRepo defines the repository interface for console operations
type ConsoleRepo interface {
	// Revenue Statistics
	QueryDateOrders(ctx context.Context, date time.Time) (*OrdersTotal, error)
	QueryMonthlyOrders(ctx context.Context, date time.Time) (*OrdersTotal, error)
	QueryTotalOrders(ctx context.Context) (*OrdersTotal, error)
	QueryDailyOrdersList(ctx context.Context, date time.Time) ([]*OrdersTotalWithDate, error)
	QueryMonthlyOrdersList(ctx context.Context, date time.Time) ([]*OrdersTotalWithDate, error)

	// User Statistics
	QueryRegisterUserTotalByDate(ctx context.Context, date time.Time) (int64, error)
	QueryRegisterUserTotalByMonthly(ctx context.Context, date time.Time) (int64, error)
	QueryRegisterUserTotal(ctx context.Context) (int64, error)
	QueryDateUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error)
	QueryMonthlyUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error)
	QueryTotalUserCounts(ctx context.Context) (newUsers int64, renewalUsers int64, err error)
	QueryDailyUserStatisticsList(ctx context.Context, date time.Time) ([]*UserStatistics, error)
	QueryMonthlyUserStatisticsList(ctx context.Context, date time.Time) ([]*UserStatistics, error)

	// Ticket Statistics
	QueryWaitReplyTotal(ctx context.Context) (int64, error)

	// Server Statistics
	QueryOnlineServers(ctx context.Context) (int64, error)
	QueryOfflineServers(ctx context.Context) (int64, error)
	QueryOnlineUsers(ctx context.Context) (int64, error)
	QueryTodayTraffic(ctx context.Context, date time.Time) (upload int64, download int64, err error)
	QueryMonthlyTraffic(ctx context.Context, date time.Time) (upload int64, download int64, err error)
	QueryTodayUserTrafficRanking(ctx context.Context, date time.Time) ([]*UserTrafficData, error)
	QueryYesterdayUserTrafficRanking(ctx context.Context, date time.Time) ([]*UserTrafficData, error)
	QueryTodayServerTrafficRanking(ctx context.Context, date time.Time) ([]*ServerTrafficData, error)
	QueryYesterdayServerTrafficRanking(ctx context.Context, date time.Time) ([]*ServerTrafficData, error)
}

// ConsoleUsecase handles console business logic
type ConsoleUsecase struct {
	repo ConsoleRepo
	log  *log.Helper
}

// NewConsoleUsecase creates a new console usecase
func NewConsoleUsecase(repo ConsoleRepo, logger log.Logger) *ConsoleUsecase {
	return &ConsoleUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/console")),
	}
}

// QueryRevenueStatistics queries revenue statistics
func (uc *ConsoleUsecase) QueryRevenueStatistics(ctx context.Context) (*RevenueStatisticsResponse, error) {
	now := time.Now()

	// Get today's revenue statistics
	today, err := uc.repo.QueryDateOrders(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryDateOrders error", "error", err)
		return nil, err
	}

	// Get monthly revenue statistics
	monthly, err := uc.repo.QueryMonthlyOrders(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyOrders error", "error", err)
		return nil, err
	}

	// Get monthly daily list
	monthlyList, err := uc.repo.QueryDailyOrdersList(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryDailyOrdersList error", "error", err)
		// Don't return error, just continue with empty list
		monthlyList = []*OrdersTotalWithDate{}
	}
	monthly.List = monthlyList

	// Get all revenue statistics
	all, err := uc.repo.QueryTotalOrders(ctx)
	if err != nil {
		uc.log.Errorw("QueryTotalOrders error", "error", err)
		return nil, err
	}

	// Get all monthly list (past 6 months)
	allList, err := uc.repo.QueryMonthlyOrdersList(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyOrdersList error", "error", err)
		// Don't return error, just continue with empty list
		allList = []*OrdersTotalWithDate{}
	}
	all.List = allList

	return &RevenueStatisticsResponse{
		Today:   today,
		Monthly: monthly,
		All:     all,
	}, nil
}

// QueryUserStatistics queries user statistics
func (uc *ConsoleUsecase) QueryUserStatistics(ctx context.Context) (*UserStatisticsResponse, error) {
	now := time.Now()
	resp := &UserStatisticsResponse{
		Today:   &UserStatistics{},
		Monthly: &UserStatistics{},
		All:     &UserStatistics{},
	}

	// Query today user register count
	todayRegister, err := uc.repo.QueryRegisterUserTotalByDate(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryRegisterUserTotalByDate error", "error", err)
	} else {
		resp.Today.Register = todayRegister
	}

	// Query today user purchase count
	newToday, renewalToday, err := uc.repo.QueryDateUserCounts(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryDateUserCounts error", "error", err)
	} else {
		resp.Today.NewOrderUsers = newToday
		resp.Today.RenewalOrderUsers = renewalToday
	}

	// Query month user register count
	monthRegister, err := uc.repo.QueryRegisterUserTotalByMonthly(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryRegisterUserTotalByMonthly error", "error", err)
	} else {
		resp.Monthly.Register = monthRegister
	}

	// Query month user purchase count
	newMonth, renewalMonth, err := uc.repo.QueryMonthlyUserCounts(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyUserCounts error", "error", err)
	} else {
		resp.Monthly.NewOrderUsers = newMonth
		resp.Monthly.RenewalOrderUsers = renewalMonth
	}

	// Get monthly daily user statistics list
	monthlyList, err := uc.repo.QueryDailyUserStatisticsList(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryDailyUserStatisticsList error", "error", err)
		monthlyList = []*UserStatistics{}
	}
	resp.Monthly.List = monthlyList

	// Query all user count
	allRegister, err := uc.repo.QueryRegisterUserTotal(ctx)
	if err != nil {
		uc.log.Errorw("QueryRegisterUserTotal error", "error", err)
	} else {
		resp.All.Register = allRegister
	}

	// Query all user order counts
	allNew, allRenewal, err := uc.repo.QueryTotalUserCounts(ctx)
	if err != nil {
		uc.log.Errorw("QueryTotalUserCounts error", "error", err)
	} else {
		resp.All.NewOrderUsers = allNew
		resp.All.RenewalOrderUsers = allRenewal
	}

	// Get all monthly user statistics list (past 6 months)
	allList, err := uc.repo.QueryMonthlyUserStatisticsList(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyUserStatisticsList error", "error", err)
		allList = []*UserStatistics{}
	}
	resp.All.List = allList

	return resp, nil
}

// QueryTicketWaitReply queries waiting reply ticket count
func (uc *ConsoleUsecase) QueryTicketWaitReply(ctx context.Context) (*TicketWaitReplyResponse, error) {
	count, err := uc.repo.QueryWaitReplyTotal(ctx)
	if err != nil {
		uc.log.Errorw("QueryWaitReplyTotal error", "error", err)
		return nil, err
	}

	return &TicketWaitReplyResponse{
		Count: count,
	}, nil
}

// QueryServerTotalData queries server total data
func (uc *ConsoleUsecase) QueryServerTotalData(ctx context.Context) (*ServerTotalDataResponse, error) {
	now := time.Now()
	resp := &ServerTotalDataResponse{
		UpdatedAt: now.Unix(),
	}

	// Query online servers
	onlineServers, err := uc.repo.QueryOnlineServers(ctx)
	if err != nil {
		uc.log.Errorw("QueryOnlineServers error", "error", err)
	} else {
		resp.OnlineServers = onlineServers
	}

	// Query offline servers
	offlineServers, err := uc.repo.QueryOfflineServers(ctx)
	if err != nil {
		uc.log.Errorw("QueryOfflineServers error", "error", err)
	} else {
		resp.OfflineServers = offlineServers
	}

	// Query today traffic
	todayUpload, todayDownload, err := uc.repo.QueryTodayTraffic(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryTodayTraffic error", "error", err)
	} else {
		resp.TodayUpload = todayUpload
		resp.TodayDownload = todayDownload
	}

	// Query monthly traffic
	monthlyUpload, monthlyDownload, err := uc.repo.QueryMonthlyTraffic(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyTraffic error", "error", err)
	} else {
		resp.MonthlyUpload = monthlyUpload
		resp.MonthlyDownload = monthlyDownload
	}

	// Query online users
	onlineUsers, err := uc.repo.QueryOnlineUsers(ctx)
	if err != nil {
		uc.log.Errorw("QueryOnlineUsers error", "error", err)
	} else {
		resp.OnlineUsers = onlineUsers
	}

	// Query today user traffic ranking
	todayUserRanking, err := uc.repo.QueryTodayUserTrafficRanking(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryTodayUserTrafficRanking error", "error", err)
	} else {
		resp.UserTrafficRankingToday = todayUserRanking
	}

	// Query yesterday user traffic ranking
	yesterdayUserRanking, err := uc.repo.QueryYesterdayUserTrafficRanking(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryYesterdayUserTrafficRanking error", "error", err)
	} else {
		resp.UserTrafficRankingYesterday = yesterdayUserRanking
	}

	// Query today server traffic ranking
	todayServerRanking, err := uc.repo.QueryTodayServerTrafficRanking(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryTodayServerTrafficRanking error", "error", err)
	} else {
		resp.ServerTrafficRankingToday = todayServerRanking
	}

	// Query yesterday server traffic ranking
	yesterdayServerRanking, err := uc.repo.QueryYesterdayServerTrafficRanking(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryYesterdayServerTrafficRanking error", "error", err)
	} else {
		resp.ServerTrafficRankingYesterday = yesterdayServerRanking
	}

	return resp, nil
}

// Response structures

type RevenueStatisticsResponse struct {
	Today   *OrdersTotal `json:"today"`
	Monthly *OrdersTotal `json:"monthly"`
	All     *OrdersTotal `json:"all"`
}

type UserStatisticsResponse struct {
	Today   *UserStatistics `json:"today"`
	Monthly *UserStatistics `json:"monthly"`
	All     *UserStatistics `json:"all"`
}

type TicketWaitReplyResponse struct {
	Count int64 `json:"count"`
}

type ServerTotalDataResponse struct {
	OnlineUsers                   int64                `json:"online_users"`
	OnlineServers                 int64                `json:"online_servers"`
	OfflineServers                int64                `json:"offline_servers"`
	TodayUpload                   int64                `json:"today_upload"`
	TodayDownload                 int64                `json:"today_download"`
	MonthlyUpload                 int64                `json:"monthly_upload"`
	MonthlyDownload               int64                `json:"monthly_download"`
	UpdatedAt                     int64                `json:"updated_at"`
	ServerTrafficRankingToday     []*ServerTrafficData `json:"server_traffic_ranking_today,omitempty"`
	ServerTrafficRankingYesterday []*ServerTrafficData `json:"server_traffic_ranking_yesterday,omitempty"`
	UserTrafficRankingToday       []*UserTrafficData   `json:"user_traffic_ranking_today,omitempty"`
	UserTrafficRankingYesterday   []*UserTrafficData   `json:"user_traffic_ranking_yesterday,omitempty"`
}
