package user

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
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
			UserId:      int64(item.UserID),
			OrderId:     int64(item.OrderID),
			SubscribeId: int64(item.SubscribeID),
			StartTime:   item.StartTime.UnixMilli(),
			CreatedAt:   item.CreatedAt.UnixMilli(),
			UpdatedAt:   item.UpdatedAt.UnixMilli(),
		}

		// 处理指针字段
		if item.ExpireTime != nil {
			protoItem.ExpireTime = item.ExpireTime.UnixMilli()
		}
		if item.FinishedAt != nil {
			protoItem.FinishedAt = item.FinishedAt.UnixMilli()
		}
		if item.Traffic != nil {
			protoItem.Traffic = *item.Traffic
		}
		if item.Download != nil {
			protoItem.Download = *item.Download
		}
		if item.Upload != nil {
			protoItem.Upload = *item.Upload
		}
		if item.Token != nil {
			protoItem.Token = *item.Token
		}
		if item.UUID != nil {
			protoItem.Uuid = *item.UUID
		}
		if item.Status != nil {
			protoItem.Status = int32(*item.Status)
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeReply{
		Total: total,
		List:  protoList,
	}, nil
}

// CreateUserSubscribe 创建用户订阅
func (s *UserSubscribeService) CreateUserSubscribe(ctx context.Context, req *v1.CreateUserSubscribeRequest) (*v1.CreateUserSubscribeReply, error) {
	id, err := s.uc.CreateUserSubscribe(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserSubscribeReply{
		Id: id,
	}, nil
}

// UpdateUserSubscribe 更新用户订阅
func (s *UserSubscribeService) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) (*v1.UpdateUserSubscribeReply, error) {
	err := s.uc.UpdateUserSubscribe(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserSubscribeReply{}, nil
}

// DeleteUserSubscribe 删除用户订阅
func (s *UserSubscribeService) DeleteUserSubscribe(ctx context.Context, req *v1.DeleteUserSubscribeRequest) (*v1.DeleteUserSubscribeReply, error) {
	err := s.uc.DeleteUserSubscribe(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserSubscribeReply{}, nil
}

// GetUserSubscribeById 根据ID获取用户订阅详情
func (s *UserSubscribeService) GetUserSubscribeById(ctx context.Context, req *v1.GetUserSubscribeByIdRequest) (*v1.GetUserSubscribeByIdReply, error) {
	subscribe, subscribeName, err := s.uc.GetUserSubscribeById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// 注意：proto定义中只返回subscribe，subscribeName暂时无法返回
	// 可以考虑将subscribeName添加到UserSubscribe消息中作为扩展字段
	_ = subscribeName // 避免未使用变量警告

	return &v1.GetUserSubscribeByIdReply{
		Subscribe: subscribe,
	}, nil
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (s *UserSubscribeService) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) (*v1.GetUserSubscribeDevicesReply, error) {
	list, _, err := s.uc.GetUserSubscribeDevices(ctx, req) // 忽略total，proto中没有该字段
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.UserDevice, 0, len(list))
	for _, item := range list {
		protoItem := &v1.UserDevice{
			Id:        int64(item.ID),
			UserId:    int64(item.UserID),
			Online:    item.Online,
			Enabled:   item.Enabled,
			CreatedAt: item.CreatedAt.UnixMilli(),
			UpdatedAt: item.UpdatedAt.UnixMilli(),
		}

		// 处理指针字段
		if item.SubscribeID != nil {
			protoItem.SubscribeId = int64(*item.SubscribeID)
		}
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
		Devices: protoList,
	}, nil
}

// GetUserSubscribeLogs 获取用户订阅日志
func (s *UserSubscribeService) GetUserSubscribeLogs(ctx context.Context, req *v1.GetUserSubscribeLogsRequest) (*v1.GetUserSubscribeLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.SubscribeLog, 0, len(list))
	for _, item := range list {
		protoItem := &v1.SubscribeLog{
			Id:              item.ID,
			UserId:          0, // 当前实现保持简化，与原项目数据结构一致
			UserSubscribeId: item.ObjectID,
			Content:         item.Content, // 返回原始JSON内容
			Timestamp:       item.CreatedAt.UnixMilli(),
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeLogsReply{
		Total: total,
		List:  protoList,
	}, nil
}

// GetUserSubscribeResetTrafficLogs 获取用户订阅重置流量日志
func (s *UserSubscribeService) GetUserSubscribeResetTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeResetTrafficLogsRequest) (*v1.GetUserSubscribeResetTrafficLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeResetTrafficLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoList := make([]*v1.ResetTrafficLog, 0, len(list))
	for _, item := range list {
		protoItem := &v1.ResetTrafficLog{
			Id:              item.ID,
			UserId:          0, // 当前实现保持简化，与原项目数据结构一致
			UserSubscribeId: item.ObjectID,
			Content:         item.Content, // 返回原始JSON内容
			Timestamp:       item.CreatedAt.UnixMilli(),
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeResetTrafficLogsReply{
		Total: total,
		List:  protoList,
	}, nil
}

// GetUserSubscribeTrafficLogs 获取用户订阅流量日志
func (s *UserSubscribeService) GetUserSubscribeTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeTrafficLogsRequest) (*v1.GetUserSubscribeTrafficLogsReply, error) {
	list, total, err := s.uc.GetUserSubscribeTrafficLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表（需要将traffic_log转为JSON格式）
	protoList := make([]*v1.SubscribeTrafficLog, 0, len(list))
	for _, item := range list {
		// 将traffic log转为JSON内容
		// Traffic log JSON序列化，当前实现与原项目保持一致
		// 构造content数据结构
		contentData := map[string]interface{}{
			"upload":    item.Upload,
			"download":  item.Download,
			"timestamp": item.Timestamp.Unix(),
		}
		contentBytes, _ := json.Marshal(contentData)
		content := string(contentBytes)

		protoItem := &v1.SubscribeTrafficLog{
			Id:              item.ID,
			UserId:          item.UserID,
			UserSubscribeId: 0, // traffic_log表中只有subscribe_id，需要关联查询得到user_subscribe_id
			Content:         content,
			Timestamp:       item.Timestamp.UnixMilli(),
		}

		protoList = append(protoList, protoItem)
	}

	return &v1.GetUserSubscribeTrafficLogsReply{
		Total: total,
		List:  protoList,
	}, nil
}
