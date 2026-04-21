package console

import (
	"context"
	"strconv"

	"github.com/OmnTeam/ppanel-pro/api/admin/console/v1"
	consolebiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/console"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// ConsoleService 控制台服务
type ConsoleService struct {
	v1.UnimplementedAdminConsoleServer

	uc  *consolebiz.ConsoleUsecase
	log *log.Helper
}

// NewConsoleService 创建控制台服务
func NewConsoleService(uc *consolebiz.ConsoleUsecase, logger log.Logger) *ConsoleService {
	return &ConsoleService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/admin/console")),
	}
}

// QueryRevenueStatistics 查询营收统计
func (s *ConsoleService) QueryRevenueStatistics(ctx context.Context, req *v1.QueryRevenueStatisticsRequest) (*v1.QueryRevenueStatisticsReply, error) {
	resp, err := s.uc.QueryRevenueStatistics(ctx)
	if err != nil {
		s.log.Errorw("msg", "query revenue statistics failed", "error", err)
		return nil, err
	}

	// 转换为 proto message
	return &v1.QueryRevenueStatisticsReply{
		Code:    int32(responsecode.QueryRevenueStatisticsSuccess),
		Message: responsecode.CodeMessages[responsecode.QueryRevenueStatisticsSuccess],
		Data: &v1.QueryRevenueStatisticsData{
			Today:   convertOrdersTotal(resp.Today),
			Monthly: convertOrdersTotal(resp.Monthly),
			All:     convertOrdersTotal(resp.All),
		},
	}, nil
}

// QueryUserStatistics 查询用户统计
func (s *ConsoleService) QueryUserStatistics(ctx context.Context, req *v1.QueryUserStatisticsRequest) (*v1.QueryUserStatisticsReply, error) {
	resp, err := s.uc.QueryUserStatistics(ctx)
	if err != nil {
		s.log.Errorw("msg", "query user statistics failed", "error", err)
		return nil, err
	}

	// 转换为 proto message
	return &v1.QueryUserStatisticsReply{
		Code:    int32(responsecode.QueryUserStatisticsSuccess),
		Message: responsecode.CodeMessages[responsecode.QueryUserStatisticsSuccess],
		Data: &v1.QueryUserStatisticsData{
			Today:   convertUserStatistics(resp.Today),
			Monthly: convertUserStatistics(resp.Monthly),
			All:     convertUserStatistics(resp.All),
		},
	}, nil
}

// QueryTicketWaitReply 查询待回复工单数量
func (s *ConsoleService) QueryTicketWaitReply(ctx context.Context, req *v1.QueryTicketWaitReplyRequest) (*v1.QueryTicketWaitReplyReply, error) {
	resp, err := s.uc.QueryTicketWaitReply(ctx)
	if err != nil {
		s.log.Errorw("msg", "query ticket wait reply failed", "error", err)
		return nil, err
	}

	return &v1.QueryTicketWaitReplyReply{
		Code:    int32(responsecode.QueryTicketWaitReplySuccess),
		Message: responsecode.CodeMessages[responsecode.QueryTicketWaitReplySuccess],
		Data: &v1.QueryTicketWaitReplyData{
			Count: int64(resp.Count),
		},
	}, nil
}

// QueryServerTotalData 查询服务器总数据
func (s *ConsoleService) QueryServerTotalData(ctx context.Context, req *v1.QueryServerTotalDataRequest) (*v1.QueryServerTotalDataReply, error) {
	resp, err := s.uc.QueryServerTotalData(ctx)
	if err != nil {
		s.log.Errorw("msg", "query server total data failed", "error", err)
		return nil, err
	}

	// 转换为 proto message
	data := &v1.QueryServerTotalDataData{
		OnlineUsers:     int64(resp.OnlineUsers),
		OnlineServers:   int64(resp.OnlineServers),
		OfflineServers:  int64(resp.OfflineServers),
		TodayUpload:     int64(resp.TodayUpload),
		TodayDownload:   int64(resp.TodayDownload),
		MonthlyUpload:   int64(resp.MonthlyUpload),
		MonthlyDownload: int64(resp.MonthlyDownload),
		UpdatedAt:       resp.UpdatedAt,
	}

	// 转换服务器流量排行
	if len(resp.ServerTrafficRankingToday) > 0 {
		data.ServerTrafficRankingToday = make([]*v1.ServerTrafficData, 0, len(resp.ServerTrafficRankingToday))
		for _, item := range resp.ServerTrafficRankingToday {
			data.ServerTrafficRankingToday = append(data.ServerTrafficRankingToday, &v1.ServerTrafficData{
				ServerId: strconv.FormatInt(int64(item.ServerID), 10),
				Name:     item.Name,
				Upload:   int64(item.Upload),
				Download: int64(item.Download),
			})
		}
	}

	if len(resp.ServerTrafficRankingYesterday) > 0 {
		data.ServerTrafficRankingYesterday = make([]*v1.ServerTrafficData, 0, len(resp.ServerTrafficRankingYesterday))
		for _, item := range resp.ServerTrafficRankingYesterday {
			data.ServerTrafficRankingYesterday = append(data.ServerTrafficRankingYesterday, &v1.ServerTrafficData{
				ServerId: strconv.FormatInt(int64(item.ServerID), 10),
				Name:     item.Name,
				Upload:   int64(item.Upload),
				Download: int64(item.Download),
			})
		}
	}

	// 转换用户流量排行
	if len(resp.UserTrafficRankingToday) > 0 {
		data.UserTrafficRankingToday = make([]*v1.UserTrafficData, 0, len(resp.UserTrafficRankingToday))
		for _, item := range resp.UserTrafficRankingToday {
			data.UserTrafficRankingToday = append(data.UserTrafficRankingToday, &v1.UserTrafficData{
				Sid:      strconv.FormatInt(int64(item.SID), 10),
				Upload:   int64(item.Upload),
				Download: int64(item.Download),
			})
		}
	}

	if len(resp.UserTrafficRankingYesterday) > 0 {
		data.UserTrafficRankingYesterday = make([]*v1.UserTrafficData, 0, len(resp.UserTrafficRankingYesterday))
		for _, item := range resp.UserTrafficRankingYesterday {
			data.UserTrafficRankingYesterday = append(data.UserTrafficRankingYesterday, &v1.UserTrafficData{
				Sid:      strconv.FormatInt(int64(item.SID), 10),
				Upload:   int64(item.Upload),
				Download: int64(item.Download),
			})
		}
	}

	return &v1.QueryServerTotalDataReply{
		Code:    int32(responsecode.QueryServerTotalDataSuccess),
		Message: responsecode.CodeMessages[responsecode.QueryServerTotalDataSuccess],
		Data:    data,
	}, nil
}

// Helper functions for conversion

func convertOrdersTotal(orders *consolebiz.OrdersTotal) *v1.OrdersStatistics {
	if orders == nil {
		return nil
	}

	result := &v1.OrdersStatistics{
		AmountTotal:        int64(orders.AmountTotal),
		NewOrderAmount:     int64(orders.NewOrderAmount),
		RenewalOrderAmount: int64(orders.RenewalOrderAmount),
	}

	if len(orders.List) > 0 {
		result.List = make([]*v1.OrdersStatisticsWithDate, 0, len(orders.List))
		for _, item := range orders.List {
			result.List = append(result.List, &v1.OrdersStatisticsWithDate{
				Date:               item.Date,
				AmountTotal:        int64(item.AmountTotal),
				NewOrderAmount:     int64(item.NewOrderAmount),
				RenewalOrderAmount: int64(item.RenewalOrderAmount),
			})
		}
	}

	return result
}

func convertUserStatistics(stats *consolebiz.UserStatistics) *v1.UserStatistics {
	if stats == nil {
		return nil
	}

	result := &v1.UserStatistics{
		Date:              stats.Date,
		Register:          int64(stats.Register),
		NewOrderUsers:     int64(stats.NewOrderUsers),
		RenewalOrderUsers: int64(stats.RenewalOrderUsers),
	}

	if len(stats.List) > 0 {
		result.List = make([]*v1.UserStatistics, 0, len(stats.List))
		for _, item := range stats.List {
			result.List = append(result.List, &v1.UserStatistics{
				Date:              item.Date,
				Register:          int64(item.Register),
				NewOrderUsers:     int64(item.NewOrderUsers),
				RenewalOrderUsers: int64(item.RenewalOrderUsers),
			})
		}
	}

	return result
}
