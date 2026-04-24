package console

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// OrdersTotal represents order statistics totals
type OrdersTotal struct {
	AmountTotal        int                    `json:"amount_total"`
	NewOrderAmount     int                    `json:"new_order_amount"`
	RenewalOrderAmount int                    `json:"renewal_order_amount"`
	List               []*OrdersTotalWithDate `json:"list,omitempty"`
}

// OrdersTotalWithDate represents order statistics with date
type OrdersTotalWithDate struct {
	Date               string `json:"date"`
	AmountTotal        int    `json:"amount_total"`
	NewOrderAmount     int    `json:"new_order_amount"`
	RenewalOrderAmount int    `json:"renewal_order_amount"`
}

// UserStatistics represents user statistics
type UserStatistics struct {
	Date              string            `json:"date,omitempty"`
	Register          int               `json:"register"`
	NewOrderUsers     int               `json:"new_order_users"`
	RenewalOrderUsers int               `json:"renewal_order_users"`
	List              []*UserStatistics `json:"list,omitempty"`
}

// UserTrafficData represents user traffic ranking data
type UserTrafficData struct {
	SID      int `json:"sid"`
	Upload   int `json:"upload"`
	Download int `json:"download"`
}

// ServerTrafficData represents server traffic ranking data
type ServerTrafficData struct {
	ServerID int    `json:"server_id"`
	Name     string `json:"name"`
	Upload   int    `json:"upload"`
	Download int    `json:"download"`
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
	QueryRegisterUserTotalByDate(ctx context.Context, date time.Time) (int, error)
	QueryRegisterUserTotalByMonthly(ctx context.Context, date time.Time) (int, error)
	QueryRegisterUserTotal(ctx context.Context) (int, error)
	QueryDateUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error)
	QueryMonthlyUserCounts(ctx context.Context, date time.Time) (newUsers int64, renewalUsers int64, err error)
	QueryTotalUserCounts(ctx context.Context) (newUsers int64, renewalUsers int64, err error)
	QueryDailyUserStatisticsList(ctx context.Context, date time.Time) ([]*UserStatistics, error)
	QueryMonthlyUserStatisticsList(ctx context.Context, date time.Time) ([]*UserStatistics, error)

	// Ticket Statistics
	QueryWaitReplyTotal(ctx context.Context) (int, error)

	// Server Statistics
	QueryOnlineServers(ctx context.Context) (int, error)
	QueryOfflineServers(ctx context.Context) (int, error)
	QueryOnlineUsers(ctx context.Context) (int, error)
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
	if strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo" {
		return uc.mockRevenueStatistics(), nil
	}
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
	if strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo" {
		return uc.mockUserStatistics(), nil
	}
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
		resp.Today.Register = int(todayRegister)
	}

	// Query today user purchase count
	newToday, renewalToday, err := uc.repo.QueryDateUserCounts(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryDateUserCounts error", "error", err)
	} else {
		resp.Today.NewOrderUsers = int(newToday)
		resp.Today.RenewalOrderUsers = int(renewalToday)
	}

	// Query month user register count
	monthRegister, err := uc.repo.QueryRegisterUserTotalByMonthly(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryRegisterUserTotalByMonthly error", "error", err)
	} else {
		resp.Monthly.Register = int(monthRegister)
	}

	// Query month user purchase count
	newMonth, renewalMonth, err := uc.repo.QueryMonthlyUserCounts(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyUserCounts error", "error", err)
	} else {
		resp.Monthly.NewOrderUsers = int(newMonth)
		resp.Monthly.RenewalOrderUsers = int(renewalMonth)
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
		resp.All.Register = int(allRegister)
	}

	// Query all user order counts
	allNew, allRenewal, err := uc.repo.QueryTotalUserCounts(ctx)
	if err != nil {
		uc.log.Errorw("QueryTotalUserCounts error", "error", err)
	} else {
		resp.All.NewOrderUsers = int(allNew)
		resp.All.RenewalOrderUsers = int(allRenewal)
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
	if strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo" {
		return uc.mockServerTotalData(now), nil
	}
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
		resp.TodayUpload = int(todayUpload)
		resp.TodayDownload = int(todayDownload)
	}

	// Query monthly traffic
	monthlyUpload, monthlyDownload, err := uc.repo.QueryMonthlyTraffic(ctx, now)
	if err != nil {
		uc.log.Errorw("QueryMonthlyTraffic error", "error", err)
	} else {
		resp.MonthlyUpload = int(monthlyUpload)
		resp.MonthlyDownload = int(monthlyDownload)
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
	Count int `json:"count"`
}

type ServerTotalDataResponse struct {
	OnlineUsers                   int                  `json:"online_users"`
	OnlineServers                 int                  `json:"online_servers"`
	OfflineServers                int                  `json:"offline_servers"`
	TodayUpload                   int                  `json:"today_upload"`
	TodayDownload                 int                  `json:"today_download"`
	MonthlyUpload                 int                  `json:"monthly_upload"`
	MonthlyDownload               int                  `json:"monthly_download"`
	UpdatedAt                     int64                `json:"updated_at"`
	ServerTrafficRankingToday     []*ServerTrafficData `json:"server_traffic_ranking_today,omitempty"`
	ServerTrafficRankingYesterday []*ServerTrafficData `json:"server_traffic_ranking_yesterday,omitempty"`
	UserTrafficRankingToday       []*UserTrafficData   `json:"user_traffic_ranking_today,omitempty"`
	UserTrafficRankingYesterday   []*UserTrafficData   `json:"user_traffic_ranking_yesterday,omitempty"`
}

func (uc *ConsoleUsecase) mockRevenueStatistics() *RevenueStatisticsResponse {
	now := time.Now()
	monthlyList := make([]*OrdersTotalWithDate, 7)
	for i := 0; i < 7; i++ {
		dayDate := now.AddDate(0, 0, -(6 - i))
		baseAmount := 25000 + ((6 - i) * 3000) + ((6-i)%3)*8000
		monthlyList[i] = &OrdersTotalWithDate{
			Date:               dayDate.Format("2006-01-02"),
			AmountTotal:        baseAmount,
			NewOrderAmount:     int(float64(baseAmount) * 0.68),
			RenewalOrderAmount: int(float64(baseAmount) * 0.32),
		}
	}
	allList := make([]*OrdersTotalWithDate, 6)
	for i := 0; i < 6; i++ {
		monthDate := now.AddDate(0, -(5 - i), 0)
		baseAmount := 1800000 + ((5 - i) * 200000) + ((5-i)%2)*500000
		allList[i] = &OrdersTotalWithDate{
			Date:               monthDate.Format("2006-01"),
			AmountTotal:        baseAmount,
			NewOrderAmount:     int(float64(baseAmount) * 0.68),
			RenewalOrderAmount: int(float64(baseAmount) * 0.32),
		}
	}
	return &RevenueStatisticsResponse{
		Today: &OrdersTotal{AmountTotal: 35888, NewOrderAmount: 22888, RenewalOrderAmount: 13000},
		Monthly: &OrdersTotal{
			AmountTotal:        888888,
			NewOrderAmount:     588888,
			RenewalOrderAmount: 300000,
			List:               monthlyList,
		},
		All: &OrdersTotal{
			AmountTotal:        12888888,
			NewOrderAmount:     8588888,
			RenewalOrderAmount: 4300000,
			List:               allList,
		},
	}
}

func (uc *ConsoleUsecase) mockUserStatistics() *UserStatisticsResponse {
	now := time.Now()
	monthlyList := make([]*UserStatistics, 7)
	for i := 0; i < 7; i++ {
		dayDate := now.AddDate(0, 0, -(6 - i))
		baseRegister := 18 + ((6 - i) * 3) + ((6-i)%3)*8
		monthlyList[i] = &UserStatistics{
			Date:              dayDate.Format("2006-01-02"),
			Register:          baseRegister,
			NewOrderUsers:     int(float64(baseRegister) * 0.65),
			RenewalOrderUsers: int(float64(baseRegister) * 0.35),
		}
	}
	allList := make([]*UserStatistics, 6)
	for i := 0; i < 6; i++ {
		monthDate := now.AddDate(0, -(5 - i), 0)
		baseRegister := 1800 + ((5 - i) * 200) + ((5-i)%2)*500
		allList[i] = &UserStatistics{
			Date:              monthDate.Format("2006-01"),
			Register:          baseRegister,
			NewOrderUsers:     int(float64(baseRegister) * 0.65),
			RenewalOrderUsers: int(float64(baseRegister) * 0.35),
		}
	}
	return &UserStatisticsResponse{
		Today: &UserStatistics{Register: 28, NewOrderUsers: 18, RenewalOrderUsers: 10},
		Monthly: &UserStatistics{
			Register:          888,
			NewOrderUsers:     588,
			RenewalOrderUsers: 300,
			List:              monthlyList,
		},
		All: &UserStatistics{
			Register: 18888,
			List:     allList,
		},
	}
}

func (uc *ConsoleUsecase) mockServerTotalData(now time.Time) *ServerTotalDataResponse {
	serverNames := []string{"香港-01", "美国-洛杉矶", "日本-东京", "新加坡-01", "韩国-首尔", "台湾-01", "德国-法兰克福", "英国-伦敦", "加拿大-多伦多", "澳洲-悉尼"}
	serverTrafficToday := make([]*ServerTrafficData, 10)
	serverTrafficYesterday := make([]*ServerTrafficData, 10)
	for i := 0; i < 10; i++ {
		serverTrafficToday[i] = &ServerTrafficData{
			ServerID: i + 1,
			Name:     serverNames[i],
			Upload:   500000000 + (i * 100000000) + (i%3)*200000000,
			Download: 2000000000 + (i * 300000000) + (i%4)*500000000,
		}
		serverTrafficYesterday[i] = &ServerTrafficData{
			ServerID: i + 1,
			Name:     serverNames[i],
			Upload:   480000000 + (i * 95000000) + (i%3)*180000000,
			Download: 1900000000 + (i * 280000000) + (i%4)*450000000,
		}
	}
	return &ServerTotalDataResponse{
		OnlineUsers:                   1688,
		OnlineServers:                 8,
		OfflineServers:                2,
		TodayUpload:                   8888888888,
		TodayDownload:                 28888888888,
		MonthlyUpload:                 288888888888,
		MonthlyDownload:               888888888888,
		UpdatedAt:                     now.Unix(),
		ServerTrafficRankingToday:     serverTrafficToday,
		ServerTrafficRankingYesterday: serverTrafficYesterday,
	}
}
