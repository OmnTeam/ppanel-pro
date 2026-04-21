package user

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	userBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/user"
	withdrawalBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/withdrawal"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// UserService Public User服务实现
type UserService struct {
	v1.UnimplementedUserServer
	uc           *userBiz.UserUseCase
	withdrawalUc *withdrawalBiz.WithdrawalUsecase
}

// NewUserService 创建Public User服务
func NewUserService(uc *userBiz.UserUseCase, withdrawalUc *withdrawalBiz.WithdrawalUsecase) *UserService {
	return &UserService{
		uc:           uc,
		withdrawalUc: withdrawalUc,
	}
}

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func parseStringID(id string) (int64, error) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return v, nil
}

// QueryUserInfo 查询用户信息
func (s *UserService) QueryUserInfo(ctx context.Context, req *emptypb.Empty) (*v1.UserInfoReply, error) {
	userID := middleware.GetUserID(ctx)

	userInfo, err := s.uc.QueryUserInfo(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	authMethods := make([]*v1.UserAuthMethod, 0, len(userInfo.AuthMethods))
	for _, method := range userInfo.AuthMethods {
		authMethods = append(authMethods, &v1.UserAuthMethod{
			AuthType:       method.AuthType,
			AuthIdentifier: method.AuthIdentifier,
			Verified:       method.Verified,
		})
	}

	return &v1.UserInfoReply{
		Code:    int32(responsecode.UserInfoQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserInfoQuerySuccess],
		Data: &v1.UserInfoData{
			Id:                    formatInt64(userInfo.ID),
			Avatar:                userInfo.Avatar,
			Balance:               userInfo.Balance,
			Commission:            userInfo.Commission,
			ReferralPercentage:    userInfo.ReferralPercentage,
			OnlyFirstPurchase:     userInfo.OnlyFirstPurchase,
			GiftAmount:            userInfo.GiftAmount,
			Telegram:              userInfo.Telegram,
			ReferCode:             userInfo.ReferCode,
			RefererId:             userInfo.RefererID,
			Enable:                userInfo.Enable,
			IsAdmin:               userInfo.IsAdmin,
			EnableBalanceNotify:   userInfo.EnableBalanceNotify,
			EnableLoginNotify:     userInfo.EnableLoginNotify,
			EnableSubscribeNotify: userInfo.EnableSubscribeNotify,
			EnableTradeNotify:     userInfo.EnableTradeNotify,
			AuthMethods:           authMethods,
			CreatedAt:             userInfo.CreatedAt,
			UpdatedAt:             userInfo.UpdatedAt,
		},
	}, nil
}

// GetLoginLog 获取登录日志
func (s *UserService) GetLoginLog(ctx context.Context, req *v1.GetLoginLogRequest) (*v1.LoginLogReply, error) {
	userID := middleware.GetUserID(ctx)

	logs, total, err := s.uc.GetLoginLog(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.UserLoginLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.UserLoginLog{
			Id:        formatInt64(log.ID),
			UserId:    formatInt64(log.UserID),
			LoginIp:   log.LoginIP,
			UserAgent: log.UserAgent,
			Success:   log.Success,
			Timestamp: log.Timestamp,
		})
	}

	return &v1.LoginLogReply{
		Code:    int32(responsecode.LoginLogQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.LoginLogQuerySuccess],
		Data: &v1.LoginLogData{
			List:  list,
			Total: total,
		},
	}, nil
}

// QueryUserBalanceLog 查询用户余额日志
func (s *UserService) QueryUserBalanceLog(ctx context.Context, req *emptypb.Empty) (*v1.BalanceLogReply, error) {
	userID := middleware.GetUserID(ctx)

	logs, total, err := s.uc.QueryUserBalanceLog(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.BalanceLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.BalanceLog{
			Type:      log.Type,
			UserId:    formatInt64(log.UserID),
			Amount:    log.Amount,
			OrderNo:   log.OrderNo,
			Balance:   log.Balance,
			Timestamp: log.Timestamp,
		})
	}

	return &v1.BalanceLogReply{
		Code:    int32(responsecode.BalanceLogQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.BalanceLogQuerySuccess],
		Data: &v1.BalanceLogData{
			List:  list,
			Total: total,
		},
	}, nil
}

// QueryUserCommissionLog 查询用户佣金日志
func (s *UserService) QueryUserCommissionLog(ctx context.Context, req *v1.QueryUserCommissionLogRequest) (*v1.CommissionLogReply, error) {
	userID := middleware.GetUserID(ctx)

	logs, total, err := s.uc.QueryUserCommissionLog(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.CommissionLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.CommissionLog{
			Type:      log.Type,
			UserId:    formatInt64(log.UserID),
			Amount:    log.Amount,
			OrderNo:   log.OrderNo,
			Timestamp: log.Timestamp,
		})
	}

	return &v1.CommissionLogReply{
		Code:    int32(responsecode.CommissionLogQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.CommissionLogQuerySuccess],
		Data: &v1.CommissionLogData{
			List:  list,
			Total: total,
		},
	}, nil
}

// QueryUserAffiliate 查询用户推荐数量
func (s *UserService) QueryUserAffiliate(ctx context.Context, req *emptypb.Empty) (*v1.UserAffiliateReply, error) {
	userID := middleware.GetUserID(ctx)

	registers, totalCommission, err := s.uc.QueryUserAffiliate(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	return &v1.UserAffiliateReply{
		Code:    int32(responsecode.AffiliateQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.AffiliateQuerySuccess],
		Data: &v1.UserAffiliateData{
			Registers:       registers,
			TotalCommission: totalCommission,
		},
	}, nil
}

// QueryUserAffiliateList 查询用户推荐列表
func (s *UserService) QueryUserAffiliateList(ctx context.Context, req *v1.QueryUserAffiliateListRequest) (*v1.UserAffiliateListReply, error) {
	userID := middleware.GetUserID(ctx)

	affiliates, total, err := s.uc.QueryUserAffiliateList(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.UserAffiliate, 0, len(affiliates))
	for _, affiliate := range affiliates {
		list = append(list, &v1.UserAffiliate{
			Avatar:       affiliate.Avatar,
			Identifier:   affiliate.Identifier,
			RegisteredAt: affiliate.RegisteredAt,
			Enable:       affiliate.Enable,
		})
	}

	return &v1.UserAffiliateListReply{
		Code:    int32(responsecode.AffiliateListQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.AffiliateListQuerySuccess],
		Data: &v1.UserAffiliateListData{
			List:  list,
			Total: total,
		},
	}, nil
}

// GetOAuthMethods 获取OAuth方法
func (s *UserService) GetOAuthMethods(ctx context.Context, req *emptypb.Empty) (*v1.OAuthMethodsReply, error) {
	userID := middleware.GetUserID(ctx)

	methods, err := s.uc.GetOAuthMethods(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.UserAuthMethod, 0, len(methods))
	for _, method := range methods {
		list = append(list, &v1.UserAuthMethod{
			AuthType:       method.AuthType,
			AuthIdentifier: method.AuthIdentifier,
			Verified:       method.Verified,
		})
	}

	return &v1.OAuthMethodsReply{
		Code:    int32(responsecode.OAuthMethodsQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthMethodsQuerySuccess],
		Data: &v1.OAuthMethodsData{
			Methods: list,
		},
	}, nil
}

// QueryUserSubscribe 查询用户订阅
func (s *UserService) QueryUserSubscribe(ctx context.Context, req *emptypb.Empty) (*v1.UserSubscribeReply, error) {
	userID := middleware.GetUserID(ctx)

	list, total, err := s.uc.QueryUserSubscribe(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	subscribeList := make([]*v1.UserSubscribe, 0, len(list))
	for _, item := range list {
		sub := &v1.UserSubscribe{
			Id:          formatInt64(item.ID),
			UserId:      formatInt64(item.UserID),
			OrderId:     item.OrderID,
			SubscribeId: formatInt64(item.SubscribeID),
			StartTime:   item.StartTime,
			ExpireTime:  item.ExpireTime,
			FinishedAt:  item.FinishedAt,
			ResetTime:   item.ResetTime,
			Traffic:     item.Traffic,
			Download:    item.Download,
			Upload:      item.Upload,
			Token:       item.Token,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		if item.Subscribe != nil {
			sub.Subscribe = &v1.Subscribe{
				Id:             formatInt64(item.Subscribe.ID),
				Name:           item.Subscribe.Name,
				Description:    item.Subscribe.Description,
				Price:          item.Subscribe.Price,
				Traffic:        item.Subscribe.Traffic,
				DeviceLimit:    item.Subscribe.DeviceLimit,
				SpeedLimit:     item.Subscribe.SpeedLimit,
				UnitTime:       item.Subscribe.UnitTime,
				UnitPrice:      item.Subscribe.UnitPrice,
				ResetCycle:     item.Subscribe.ResetCycle,
				DeductionRatio: item.Subscribe.DeductionRatio,
				AllowDeduction: item.Subscribe.AllowDeduction,
				Enable:         item.Subscribe.Enable,
				CreatedAt:      item.Subscribe.CreatedAt,
				UpdatedAt:      item.Subscribe.UpdatedAt,
			}

			for _, discount := range item.Subscribe.Discount {
				sub.Subscribe.Discount = append(sub.Subscribe.Discount, &v1.SubscribeDiscount{
					Quantity:   discount.Quantity,
					Percentage: discount.Percentage,
				})
			}
		}

		subscribeList = append(subscribeList, sub)
	}

	return &v1.UserSubscribeReply{
		Code:    int32(responsecode.UserSubscribeQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserSubscribeQuerySuccess],
		Data: &v1.UserSubscribeData{
			List:  subscribeList,
			Total: total,
		},
	}, nil
}

// GetSubscribeLog 获取订阅日志
func (s *UserService) GetSubscribeLog(ctx context.Context, req *v1.GetSubscribeLogRequest) (*v1.SubscribeLogReply, error) {
	userID := middleware.GetUserID(ctx)

	logs, total, err := s.uc.GetSubscribeLog(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.UserSubscribeLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.UserSubscribeLog{
			Id:              formatInt64(log.ID),
			UserId:          formatInt64(log.UserID),
			UserSubscribeId: formatInt64(log.UserSubscribeID),
			Token:           log.Token,
			Ip:              log.IP,
			UserAgent:       log.UserAgent,
			Timestamp:       log.Timestamp,
		})
	}

	return &v1.SubscribeLogReply{
		Code:    int32(responsecode.SubscribeLogQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.SubscribeLogQuerySuccess],
		Data: &v1.SubscribeLogData{
			List:  list,
			Total: total,
		},
	}, nil
}

// ResetUserSubscribeToken 重置订阅令牌
func (s *UserService) ResetUserSubscribeToken(ctx context.Context, req *v1.ResetUserSubscribeTokenRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	userSubscribeID, err := parseStringID(req.UserSubscribeId)
	if err != nil {
		return nil, err
	}

	err = s.uc.ResetUserSubscribeToken(ctx, int(userID), int(userSubscribeID))
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.SubscribeTokenResetSuccess),
		Message: responsecode.CodeMessages[responsecode.SubscribeTokenResetSuccess],
	}, nil
}

// PreUnsubscribe 预退订
func (s *UserService) PreUnsubscribe(ctx context.Context, req *v1.PreUnsubscribeRequest) (*v1.UnsubscribeInfoReply, error) {
	userID := middleware.GetUserID(ctx)

	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	deductionAmount, err := s.uc.PreUnsubscribe(ctx, int(userID), int(id))
	if err != nil {
		return nil, err
	}

	return &v1.UnsubscribeInfoReply{
		Code:    int32(responsecode.PreUnsubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.PreUnsubscribeSuccess],
		Data: &v1.UnsubscribeInfoData{
			DeductionAmount: deductionAmount,
		},
	}, nil
}

// Unsubscribe 退订
func (s *UserService) Unsubscribe(ctx context.Context, req *v1.UnsubscribeRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.uc.Unsubscribe(ctx, int(userID), int(id))
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.UnsubscribeSuccess),
		Message: responsecode.CodeMessages[responsecode.UnsubscribeSuccess],
	}, nil
}

// UpdateUserNotify 更新通知设置
func (s *UserService) UpdateUserNotify(ctx context.Context, req *v1.UpdateUserNotifyRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UpdateUserNotify(ctx, int(userID), req.EnableLoginNotify, req.EnableBalanceNotify, req.EnableSubscribeNotify, req.EnableTradeNotify)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.NotifyUpdateSuccess),
		Message: responsecode.CodeMessages[responsecode.NotifyUpdateSuccess],
	}, nil
}

// UpdateUserPassword 更新密码
func (s *UserService) UpdateUserPassword(ctx context.Context, req *v1.UpdateUserPasswordRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UpdateUserPassword(ctx, int(userID), req.Password)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.PasswordUpdateSuccess),
		Message: responsecode.CodeMessages[responsecode.PasswordUpdateSuccess],
	}, nil
}

// BindTelegram 绑定Telegram
func (s *UserService) BindTelegram(ctx context.Context, req *emptypb.Empty) (*v1.TelegramBindReply, error) {
	session := middleware.GetSessionID(ctx)
	botName := ""

	url, expiredAt, err := s.uc.BindTelegram(ctx, session, botName)
	if err != nil {
		return nil, err
	}

	return &v1.TelegramBindReply{
		Code:    int32(responsecode.TelegramBindSuccess),
		Message: responsecode.CodeMessages[responsecode.TelegramBindSuccess],
		Data: &v1.TelegramBindData{
			Url:       url,
			ExpiredAt: expiredAt,
		},
	}, nil
}

// UnbindTelegram 解绑Telegram
func (s *UserService) UnbindTelegram(ctx context.Context, req *emptypb.Empty) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UnbindTelegram(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.TelegramUnbindSuccess),
		Message: responsecode.CodeMessages[responsecode.TelegramUnbindSuccess],
	}, nil
}

// BindOAuth 绑定OAuth
func (s *UserService) BindOAuth(ctx context.Context, req *v1.BindOAuthRequest) (*v1.OAuthBindReply, error) {
	redirect, err := s.uc.BindOAuth(ctx, req.Method, req.Redirect)
	if err != nil {
		return nil, err
	}

	return &v1.OAuthBindReply{
		Code:    int32(responsecode.OAuthBindSuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthBindSuccess],
		Data: &v1.OAuthBindData{
			Redirect: redirect,
		},
	}, nil
}

// BindOAuthCallback OAuth回调
func (s *UserService) BindOAuthCallback(ctx context.Context, req *v1.BindOAuthCallbackRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.BindOAuthCallback(ctx, int(userID), req.Method, req.Callback)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.OAuthCallbackSuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthCallbackSuccess],
	}, nil
}

// UnbindOAuth 解绑OAuth
func (s *UserService) UnbindOAuth(ctx context.Context, req *v1.UnbindOAuthRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UnbindOAuth(ctx, int(userID), req.Method)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.OAuthUnbindSuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthUnbindSuccess],
	}, nil
}

// VerifyEmail 验证邮箱
func (s *UserService) VerifyEmail(ctx context.Context, req *v1.VerifyEmailRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.VerifyEmail(ctx, int(userID), req.Email, req.Code)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.EmailVerifySuccess),
		Message: responsecode.CodeMessages[responsecode.EmailVerifySuccess],
	}, nil
}

// UpdateBindMobile 更新绑定手机
func (s *UserService) UpdateBindMobile(ctx context.Context, req *v1.UpdateBindMobileRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UpdateBindMobile(ctx, int(userID), req.AreaCode, req.Mobile, req.Code)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.MobileBindSuccess),
		Message: responsecode.CodeMessages[responsecode.MobileBindSuccess],
	}, nil
}

// UpdateBindEmail 更新绑定邮箱
func (s *UserService) UpdateBindEmail(ctx context.Context, req *v1.UpdateBindEmailRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	err := s.uc.UpdateBindEmail(ctx, int(userID), req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.EmailBindSuccess),
		Message: responsecode.CodeMessages[responsecode.EmailBindSuccess],
	}, nil
}

// DeviceWSConnect 设备WebSocket连接
func (s *UserService) DeviceWSConnect(ctx context.Context, req *emptypb.Empty) (*v1.CommonReply, error) {
	err := s.uc.DeviceWSConnect(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.UserCreated),
		Message: "设备连接成功",
	}, nil
}

// GetDeviceList 获取设备列表
func (s *UserService) GetDeviceList(ctx context.Context, req *emptypb.Empty) (*v1.GetDeviceListReply, error) {
	userID := middleware.GetUserID(ctx)

	list, total, err := s.uc.GetDeviceList(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	deviceList := make([]*v1.UserDevice, 0, len(list))
	for _, device := range list {
		deviceList = append(deviceList, &v1.UserDevice{
			Id:         device.ID,
			Ip:         device.IP,
			Identifier: device.Identifier,
			UserAgent:  device.UserAgent,
			Online:     device.Online,
			Enabled:    device.Enabled,
			CreatedAt:  device.CreatedAt,
			UpdatedAt:  device.UpdatedAt,
		})
	}

	return &v1.GetDeviceListReply{
		Code:    int32(responsecode.UserDeviceListQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserDeviceListQuerySuccess],
		Data: &v1.GetDeviceListData{
			List:  deviceList,
			Total: total,
		},
	}, nil
}

// UnbindDevice 解绑设备
func (s *UserService) UnbindDevice(ctx context.Context, req *v1.UnbindDeviceRequest) (*v1.CommonReply, error) {
	userID := middleware.GetUserID(ctx)

	deviceID, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.uc.UnbindDevice(ctx, int(userID), int(deviceID))
	if err != nil {
		return nil, err
	}

	return &v1.CommonReply{
		Code:    int32(responsecode.UserDeviceUnbindSuccess),
		Message: responsecode.CodeMessages[responsecode.UserDeviceUnbindSuccess],
	}, nil
}

// GetDeviceOnlineStatistics 获取设备在线统计
func (s *UserService) GetDeviceOnlineStatistics(ctx context.Context, req *emptypb.Empty) (*v1.GetDeviceOnlineStatisticsReply, error) {
	userID := middleware.GetUserID(ctx)

	stats, err := s.uc.GetDeviceOnlineStatistics(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	weeklyStats := make([]*v1.WeeklyStat, 0, len(stats.WeeklyStats))
	for _, stat := range stats.WeeklyStats {
		weeklyStats = append(weeklyStats, &v1.WeeklyStat{
			Day:     stat.Day,
			DayName: stat.DayName,
			Hours:   stat.Hours,
		})
	}

	connectionRecords := &v1.ConnectionRecords{
		CurrentContinuousDays:   stats.ConnectionRecords.CurrentContinuousDays,
		HistoryContinuousDays:   stats.ConnectionRecords.HistoryContinuousDays,
		LongestSingleConnection: stats.ConnectionRecords.LongestSingleConnection,
	}

	return &v1.GetDeviceOnlineStatisticsReply{
		Code:    int32(responsecode.UserDeviceStatisticsQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserDeviceStatisticsQuerySuccess],
		Data: &v1.GetDeviceOnlineStatisticsData{
			WeeklyStats:       weeklyStats,
			ConnectionRecords: connectionRecords,
		},
	}, nil
}

// CommissionWithdraw 佣金提现
func (s *UserService) CommissionWithdraw(ctx context.Context, req *v1.CommissionWithdrawRequest) (*v1.WithdrawalLogReply, error) {
	userID := middleware.GetUserID(ctx)

	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	withdrawal, err := s.withdrawalUc.CommissionWithdraw(ctx, int64(userID), &withdrawalBiz.CommissionWithdrawRequest{
		Amount:  amount,
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}

	return &v1.WithdrawalLogReply{
		Code:    int32(responsecode.UserInfoQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserInfoQuerySuccess],
		Data: &v1.WithdrawalLogData{
			Id:        formatInt64(withdrawal.ID),
			UserId:    formatInt64(withdrawal.UserID),
			Amount:    formatInt64(withdrawal.Amount),
			Content:   withdrawal.Content,
			Status:    int32(withdrawal.Status),
			Reason:    withdrawal.Reason,
			CreatedAt: formatInt64(withdrawal.CreatedAt.UnixMilli()),
			UpdatedAt: formatInt64(withdrawal.UpdatedAt.UnixMilli()),
		},
	}, nil
}

// QueryWithdrawalLog 查询提现日志
func (s *UserService) QueryWithdrawalLog(ctx context.Context, req *v1.QueryWithdrawalLogRequest) (*v1.WithdrawalLogListReply, error) {
	userID := middleware.GetUserID(ctx)

	withdrawals, total, err := s.withdrawalUc.QueryWithdrawalLog(ctx, int64(userID), int32(req.Page), int32(req.Size))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.WithdrawalLogData, 0, len(withdrawals))
	for _, w := range withdrawals {
		list = append(list, &v1.WithdrawalLogData{
			Id:        formatInt64(w.ID),
			UserId:    formatInt64(w.UserID),
			Amount:    formatInt64(w.Amount),
			Content:   w.Content,
			Status:    int32(w.Status),
			Reason:    w.Reason,
			CreatedAt: formatInt64(w.CreatedAt.UnixMilli()),
			UpdatedAt: formatInt64(w.UpdatedAt.UnixMilli()),
		})
	}

	return &v1.WithdrawalLogListReply{
		Code:    int32(responsecode.UserInfoQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserInfoQuerySuccess],
		Data: &v1.WithdrawalLogListData{
			List:  list,
			Total: formatInt64(int64(total)),
		},
	}, nil
}
