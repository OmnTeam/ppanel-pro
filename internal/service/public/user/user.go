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

// QueryUserInfo 查询用户信息
func (s *UserService) QueryUserInfo(ctx context.Context, req *emptypb.Empty) (*v1.UserInfoReply, error) {
	// 从context获取用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	userInfo, err := s.uc.QueryUserInfo(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	// 转换AuthMethods
	authMethods := make([]*v1.UserAuthMethod, 0, len(userInfo.AuthMethods))
	for _, method := range userInfo.AuthMethods {
		authMethods = append(authMethods, &v1.UserAuthMethod{
			AuthType:       method.AuthType,
			AuthIdentifier: method.AuthIdentifier,
			Verified:       method.Verified,
		})
	}

	// 构建响应
	return &v1.UserInfoReply{
		Code:    int32(responsecode.UserInfoQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserInfoQuerySuccess],
		Data: &v1.UserInfoData{
			Id:                    userInfo.ID,
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
	// 从context获取用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	logs, total, err := s.uc.GetLoginLog(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.UserLoginLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.UserLoginLog{
			Id:        log.ID,
			UserId:    log.UserID,
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
	// 从context获取租户ID和用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	logs, total, err := s.uc.QueryUserBalanceLog(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.BalanceLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.BalanceLog{
			Type:      log.Type,
			UserId:    log.UserID,
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
	// 从context获取租户ID和用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	logs, total, err := s.uc.QueryUserCommissionLog(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.CommissionLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.CommissionLog{
			Type:      log.Type,
			UserId:    log.UserID,
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
	// 从context获取租户ID和用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
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
	// 从context获取租户ID和用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	affiliates, total, err := s.uc.QueryUserAffiliateList(ctx, int(userID), int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	// 转换结果
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
	// 从context获取租户ID和用户ID
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	methods, err := s.uc.GetOAuthMethods(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	// 转换结果
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

	// 转换结果
	subscribeList := make([]*v1.UserSubscribe, 0, len(list))
	for _, item := range list {
		sub := &v1.UserSubscribe{
			Id:          item.ID,
			UserId:      item.UserID,
			OrderId:     item.OrderID,
			SubscribeId: item.SubscribeID,
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
				Id:             item.Subscribe.ID,
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

			// 转换Discount
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

	// 转换结果
	list := make([]*v1.UserSubscribeLog, 0, len(logs))
	for _, log := range logs {
		list = append(list, &v1.UserSubscribeLog{
			Id:              log.ID,
			UserId:          log.UserID,
			UserSubscribeId: log.UserSubscribeID,
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

	err := s.uc.ResetUserSubscribeToken(ctx, int(userID), int(req.UserSubscribeId))
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

	deductionAmount, err := s.uc.PreUnsubscribe(ctx, int(userID), int(req.Id))
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

	err := s.uc.Unsubscribe(ctx, int(userID), int(req.Id))
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
	// 从context获取session（JWT中的session ID）
	// 原项目从 l.ctx.Value("session").(string) 获取
	session := middleware.GetSessionID(ctx)

	// 从配置获取Telegram Bot名称
	botName := "" // 将从系统配置获取

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

	// 转换结果
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

	err := s.uc.UnbindDevice(ctx, int(userID), int(req.Id))
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

	// 转换每周统计
	weeklyStats := make([]*v1.WeeklyStat, 0, len(stats.WeeklyStats))
	for _, stat := range stats.WeeklyStats {
		weeklyStats = append(weeklyStats, &v1.WeeklyStat{
			Day:     stat.Day,
			DayName: stat.DayName,
			Hours:   stat.Hours,
		})
	}

	// 转换连接记录
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

	// 转换amount string to int64
	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 调用业务层
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
			Id:        strconv.FormatInt(withdrawal.ID, 10),
			UserId:    strconv.FormatInt(withdrawal.UserID, 10),
			Amount:    strconv.FormatInt(withdrawal.Amount, 10),
			Content:   withdrawal.Content,
			Status:    int32(withdrawal.Status),
			Reason:    withdrawal.Reason,
			CreatedAt: strconv.FormatInt(withdrawal.CreatedAt.UnixMilli(), 10),
			UpdatedAt: strconv.FormatInt(withdrawal.UpdatedAt.UnixMilli(), 10),
		},
	}, nil
}

// QueryWithdrawalLog 查询提现日志
func (s *UserService) QueryWithdrawalLog(ctx context.Context, req *v1.QueryWithdrawalLogRequest) (*v1.WithdrawalLogListReply, error) {
	userID := middleware.GetUserID(ctx)

	// 调用业务层
	withdrawals, total, err := s.withdrawalUc.QueryWithdrawalLog(ctx, int64(userID), int32(req.Page), int32(req.Size))
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.WithdrawalLogData, 0, len(withdrawals))
	for _, w := range withdrawals {
		list = append(list, &v1.WithdrawalLogData{
			Id:        strconv.FormatInt(w.ID, 10),
			UserId:    strconv.FormatInt(w.UserID, 10),
			Amount:    strconv.FormatInt(w.Amount, 10),
			Content:   w.Content,
			Status:    int32(w.Status),
			Reason:    w.Reason,
			CreatedAt: strconv.FormatInt(w.CreatedAt.UnixMilli(), 10),
			UpdatedAt: strconv.FormatInt(w.UpdatedAt.UnixMilli(), 10),
		})
	}

	return &v1.WithdrawalLogListReply{
		Code:    int32(responsecode.UserInfoQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserInfoQuerySuccess],
		Data: &v1.WithdrawalLogListData{
			List:  list,
			Total: strconv.FormatInt(int64(total), 10),
		},
	}, nil
}
