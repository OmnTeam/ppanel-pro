package user

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// UserSubscribeService 用户订阅服务
type UserSubscribeService struct {
	v1.UnimplementedUserSubscribeServiceServer

	uc     *userbiz.SubscribeUsecase
	logger *log.Helper
}

// NewUserSubscribeService 创建用户订阅服务
func NewUserSubscribeService(uc *userbiz.SubscribeUsecase, logger log.Logger) *UserSubscribeService {
	return &UserSubscribeService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

func parseStringInt64Helper(s string) (int64, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return val, nil
}

// GetUserSubscribe 获取用户订阅列表
func (s *UserSubscribeService) GetUserSubscribe(ctx context.Context, req *v1.GetUserSubscribeRequest) (*v1.GetUserSubscribeReply, error) {
	list, total, err := s.uc.GetUserSubscribe(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.UserSubscribe, 0, len(list))
	for _, item := range list {
		protoItem := &v1.UserSubscribe{
			Id:          int64(item.ID),
			IdStr:       strconv.FormatInt(int64(item.ID), 10),
			UserId:      int64(item.UserID),
			OrderId:     int64(item.OrderID),
			SubscribeId: int64(item.SubscribeID),
			NodeGroupId: int64(item.NodeGroupID),
			GroupLocked: item.GroupLocked,
			StartTime:   item.StartTime.Unix(),
			CreatedAt:   item.CreatedAt.Unix(),
			UpdatedAt:   item.UpdatedAt.Unix(),
		}

		// 处理指针字段
		if item.ExpireTime != nil {
			protoItem.ExpireTime = item.ExpireTime.Unix()
		}
		if item.FinishedAt != nil {
			protoItem.FinishedAt = item.FinishedAt.Unix()
		}
		if item.Traffic != nil {
			protoItem.Traffic = int64(*item.Traffic)
		}
		if item.Download != nil {
			protoItem.Download = int64(*item.Download)
		}
		if item.Upload != nil {
			protoItem.Upload = int64(*item.Upload)
		}
		if item.Token != nil {
			protoItem.Token = *item.Token
		}
		if item.Status != nil {
			protoItem.Status = uint32(*item.Status)
		}
		if item.Token != nil {
			if short, shortErr := tool.FixedUniqueString(*item.Token, 8, ""); shortErr == nil {
				protoItem.Short = short
			}
		}
		if detail, detailErr := s.uc.GetUserSubscribeById(ctx, item.ID); detailErr == nil && detail != nil {
			protoItem.Subscribe = detail.Subscribe
			protoItem.ResetTime = calculateNextResetTime(protoItem)
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeReply{
		Code:    int32(responsecode.UserSubscribeQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserSubscribeQuerySuccess],
		Data: &v1.GetUserSubscribeData{
			Total: total,
			List:  protoList,
		},
	}, nil
}

func calculateNextResetTime(sub *v1.UserSubscribe) int64 {
	if sub == nil || sub.Subscribe == nil {
		return 0
	}

	resetTime := time.Unix(sub.ExpireTime, 0)
	now := time.Now()

	switch sub.Subscribe.ResetCycle {
	case 0:
		return 0
	case 1:
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
	case 2:
		if resetTime.Day() > now.Day() {
			return time.Date(now.Year(), now.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		}
		return time.Date(now.Year(), now.Month()+1, resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	case 3:
		targetTime := time.Date(now.Year(), resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		if targetTime.Before(now) {
			targetTime = time.Date(now.Year()+1, resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		}
		return targetTime.UnixMilli()
	default:
		return 0
	}
}

// CreateUserSubscribe 创建用户订阅
func (s *UserSubscribeService) CreateUserSubscribe(ctx context.Context, req *v1.CreateUserSubscribeRequest) (*v1.CreateUserSubscribeReply, error) {
	if _, err := s.uc.CreateUserSubscribe(ctx, req); err != nil {
		return nil, err
	}

	return &v1.CreateUserSubscribeReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}

// UpdateUserSubscribe 更新用户订阅
func (s *UserSubscribeService) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) (*v1.UpdateUserSubscribeReply, error) {
	err := s.uc.UpdateUserSubscribe(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserSubscribeReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}

// DeleteUserSubscribe 删除用户订阅
func (s *UserSubscribeService) DeleteUserSubscribe(ctx context.Context, req *v1.DeleteUserSubscribeRequest) (*v1.DeleteUserSubscribeReply, error) {
	if req.UserSubscribeId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err := s.uc.DeleteUserSubscribe(ctx, req.UserSubscribeId)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserSubscribeReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}

// GetUserSubscribeById 根据ID获取用户订阅详情
func (s *UserSubscribeService) GetUserSubscribeById(ctx context.Context, req *v1.GetUserSubscribeByIdRequest) (*v1.GetUserSubscribeByIdReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	subscribe, err := s.uc.GetUserSubscribeById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserSubscribeByIdReply{
		Code:    int32(responsecode.UserSubscribeQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserSubscribeQuerySuccess],
		Data:    subscribe,
	}, nil
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (s *UserSubscribeService) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) (*v1.GetUserSubscribeDevicesReply, error) {
	list, total, err := s.uc.GetUserSubscribeDevices(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.UserDevice, 0, len(list))
	for _, item := range list {
		protoItem := &v1.UserDevice{
			Id:        int64(item.ID),
			Online:    item.Online,
			Enabled:   item.Enabled,
			CreatedAt: item.CreatedAt.Unix(),
			UpdatedAt: item.UpdatedAt.Unix(),
		}

		// 处理指针字段
		if item.IP != nil {
			protoItem.Ip = *item.IP
		}
		if item.Identifier != nil {
			protoItem.Identifier = *item.Identifier
		}
		if item.UserAgent != nil {
			protoItem.UserAgent = *item.UserAgent
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeDevicesReply{
		Code:    int32(responsecode.UserDeviceListQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.UserDeviceListQuerySuccess],
		Data: &v1.GetUserSubscribeDevicesData{
			Total: total,
			List:  protoList,
		},
	}, nil
}

// GetUserSubscribeLogs 获取用户订阅日志
func (s *UserSubscribeService) GetUserSubscribeLogs(ctx context.Context, req *v1.GetUserSubscribeLogsRequest) (*v1.GetUserSubscribeLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.UserSubscribeLog, 0, len(list))
	for _, item := range list {
		content := &logmodel.Subscribe{}
		_ = content.Unmarshal([]byte(item.Content))

		protoItem := &v1.UserSubscribeLog{
			Id:              int64(item.ID),
			UserId:          0,
			UserSubscribeId: int64(item.ObjectID),
			Token:           content.Token,
			Ip:              content.ClientIP,
			UserAgent:       content.UserAgent,
			Timestamp:       item.CreatedAt.Unix(),
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeLogsReply{
		Code:    int32(responsecode.FilterSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterSubscribeLogSuccess],
		Data: &v1.GetUserSubscribeLogsData{
			Total: total,
			List:  protoList,
		},
	}, nil
}

// GetUserSubscribeResetTrafficLogs 获取用户订阅重置流量日志
func (s *UserSubscribeService) GetUserSubscribeResetTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeResetTrafficLogsRequest) (*v1.GetUserSubscribeResetTrafficLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeResetTrafficLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.ResetSubscribeTrafficLog, 0, len(list))
	for _, item := range list {
		content := &logmodel.ResetSubscribe{}
		_ = content.Unmarshal([]byte(item.Content))

		protoItem := &v1.ResetSubscribeTrafficLog{
			Id:              int64(item.ID),
			Type:            int32(content.Type),
			UserSubscribeId: int64(item.ObjectID),
			OrderNo:         content.OrderNo,
			Timestamp:       content.Timestamp,
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeResetTrafficLogsReply{
		Code:    int32(responsecode.FilterResetSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterResetSubscribeLogSuccess],
		Data: &v1.GetUserSubscribeResetTrafficLogsData{
			Total: total,
			List:  protoList,
		},
	}, nil
}

// GetUserSubscribeTrafficLogs 获取用户订阅流量日志
func (s *UserSubscribeService) GetUserSubscribeTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeTrafficLogsRequest) (*v1.GetUserSubscribeTrafficLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeTrafficLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	protoList := make([]*v1.TrafficLog, 0, len(list))
	for _, item := range list {
		protoItem := &v1.TrafficLog{
			Id:          int64(item.ID),
			ServerId:    int64(item.ServerID),
			UserId:      int64(item.UserID),
			SubscribeId: int64(item.SubscribeID),
			Download:    item.Download,
			Upload:      item.Upload,
			Timestamp:   item.Timestamp.Unix(),
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeTrafficLogsReply{
		Code:    int32(responsecode.FilterUserSubscribeTrafficLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterUserSubscribeTrafficLogSuccess],
		Data: &v1.GetUserSubscribeTrafficLogsData{
			Total: total,
			List:  protoList,
		},
	}, nil
}

// ResetUserSubscribeToken 重置用户订阅令牌
func (s *UserSubscribeService) ResetUserSubscribeToken(ctx context.Context, req *v1.ResetUserSubscribeTokenRequest) (*v1.ResetUserSubscribeTokenReply, error) {
	if req.UserSubscribeId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err := s.uc.ResetUserSubscribeToken(ctx, req.UserSubscribeId); err != nil {
		return nil, err
	}

	return &v1.ResetUserSubscribeTokenReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}

// ToggleUserSubscribeStatus 切换用户订阅状态
func (s *UserSubscribeService) ToggleUserSubscribeStatus(ctx context.Context, req *v1.ToggleUserSubscribeStatusRequest) (*v1.ToggleUserSubscribeStatusReply, error) {
	if req.UserSubscribeId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err := s.uc.ToggleUserSubscribeStatus(ctx, req.UserSubscribeId); err != nil {
		return nil, err
	}

	return &v1.ToggleUserSubscribeStatusReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}

// ResetUserSubscribeTraffic 重置用户订阅流量
func (s *UserSubscribeService) ResetUserSubscribeTraffic(ctx context.Context, req *v1.ResetUserSubscribeTrafficRequest) (*v1.ResetUserSubscribeTrafficReply, error) {
	if req.UserSubscribeId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if err := s.uc.ResetUserSubscribeTraffic(ctx, req.UserSubscribeId); err != nil {
		return nil, err
	}

	return &v1.ResetUserSubscribeTrafficReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
	}, nil
}
