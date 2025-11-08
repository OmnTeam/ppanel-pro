package log

import (
	"context"
	"encoding/json"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	logbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/log"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// LogService 日志服务（统一所有日志相关 API）
type LogService struct {
	v1.UnimplementedLogServiceServer
	systemLogUc  *logbiz.SystemLogUsecase
	trafficLogUc *logbiz.TrafficLogUsecase
	logSettingUc *logbiz.LogSettingUsecase
	log          *log.Helper
}

// NewLogService 创建统一的日志服务
func NewLogService(
	systemLogUc *logbiz.SystemLogUsecase,
	trafficLogUc *logbiz.TrafficLogUsecase,
	logSettingUc *logbiz.LogSettingUsecase,
	logger log.Logger,
) *LogService {
	return &LogService{
		systemLogUc:  systemLogUc,
		trafficLogUc: trafficLogUc,
		logSettingUc: logSettingUc,
		log:          log.NewHelper(log.With(logger, "module", "service/admin/log")),
	}
}

// ========== System Log APIs ==========

// FilterBalanceLog 过滤余额日志
func (s *LogService) FilterBalanceLog(ctx context.Context, req *v1.FilterBalanceLogRequest) (*v1.FilterBalanceLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterBalanceLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter balance log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.BalanceLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Balance
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal balance log failed", "error", err)
			continue
		}

		list = append(list, &v1.BalanceLog{
			Type:      int32(content.Type),
			UserId:    l.ObjectID,
			Amount:    content.Amount,
			OrderNo:   content.OrderNo,
			Balance:   content.Balance,
			Timestamp: content.Timestamp,
		})
	}

	return &v1.FilterBalanceLogReply{
		Code:    int32(responsecode.FilterBalanceLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterBalanceLogSuccess],
		Data: &v1.FilterBalanceLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterCommissionLog 过滤佣金日志
func (s *LogService) FilterCommissionLog(ctx context.Context, req *v1.FilterCommissionLogRequest) (*v1.FilterCommissionLogReply, error) {
	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterCommissionLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter commission log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.CommissionLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Commission
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal commission log failed", "error", err)
			continue
		}

		list = append(list, &v1.CommissionLog{
			Type:      int32(content.Type),
			UserId:    l.ObjectID,
			Amount:    content.Amount,
			OrderNo:   content.OrderNo,
			Timestamp: content.Timestamp,
		})
	}

	return &v1.FilterCommissionLogReply{
		Code:    int32(responsecode.FilterCommissionLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterCommissionLogSuccess],
		Data: &v1.FilterCommissionLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterEmailLog 过滤邮件日志
func (s *LogService) FilterEmailLog(ctx context.Context, req *v1.FilterEmailLogRequest) (*v1.FilterEmailLogReply, error) {

	logs, total, err := s.systemLogUc.FilterEmailLog(ctx, req.Page, req.Size, req.Date, req.Search)
	if err != nil {
		s.log.Errorw("msg", "filter email log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.EmailLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal email log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.EmailLog{
			Id:        l.ID,
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterEmailLogReply{
		Code:    int32(responsecode.FilterEmailLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterEmailLogSuccess],
		Data: &v1.FilterEmailLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterGiftLog 过滤赠送日志
func (s *LogService) FilterGiftLog(ctx context.Context, req *v1.FilterGiftLogRequest) (*v1.FilterGiftLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterGiftLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter gift log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.GiftLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Gift
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal gift log failed", "error", err)
			continue
		}

		list = append(list, &v1.GiftLog{
			Type:        int32(content.Type),
			UserId:      l.ObjectID,
			OrderNo:     content.OrderNo,
			SubscribeId: content.SubscribeId,
			Amount:      content.Amount,
			Balance:     content.Balance,
			Remark:      content.Remark,
			Timestamp:   content.Timestamp,
		})
	}

	return &v1.FilterGiftLogReply{
		Code:    int32(responsecode.FilterGiftLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterGiftLogSuccess],
		Data: &v1.FilterGiftLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterLoginLog 过滤登录日志
func (s *LogService) FilterLoginLog(ctx context.Context, req *v1.FilterLoginLogRequest) (*v1.FilterLoginLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterLoginLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter login log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.LoginLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Login
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal login log failed", "error", err)
			continue
		}

		list = append(list, &v1.LoginLog{
			UserId:    l.ObjectID,
			Method:    content.Method,
			LoginIp:   content.LoginIP,
			UserAgent: content.UserAgent,
			Success:   content.Success,
			Timestamp: l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterLoginLogReply{
		Code:    int32(responsecode.FilterLoginLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterLoginLogSuccess],
		Data: &v1.FilterLoginLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// GetMessageLogList 获取消息日志列表
func (s *LogService) GetMessageLogList(ctx context.Context, req *v1.GetMessageLogListRequest) (*v1.GetMessageLogListReply, error) {

	logs, total, err := s.systemLogUc.GetMessageLogList(ctx, req.Page, req.Size, req.Type, req.Search)
	if err != nil {
		s.log.Errorw("msg", "get message log list failed", "error", err)
		return nil, err
	}

	list := make([]*v1.MessageLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal message log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.MessageLog{
			Id:        l.ID,
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.GetMessageLogListReply{
		Code:    int32(responsecode.GetMessageLogListSuccess),
		Message: responsecode.CodeMessages[responsecode.GetMessageLogListSuccess],
		Data: &v1.GetMessageLogListData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterMobileLog 过滤手机日志
func (s *LogService) FilterMobileLog(ctx context.Context, req *v1.FilterMobileLogRequest) (*v1.FilterMobileLogReply, error) {

	logs, total, err := s.systemLogUc.FilterMobileLog(ctx, req.Page, req.Size, req.Date, req.Search)
	if err != nil {
		s.log.Errorw("msg", "filter mobile log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.MobileLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal mobile log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.MobileLog{
			Id:        l.ID,
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterMobileLogReply{
		Code:    int32(responsecode.FilterMobileLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterMobileLogSuccess],
		Data: &v1.FilterMobileLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterRegisterLog 过滤注册日志
func (s *LogService) FilterRegisterLog(ctx context.Context, req *v1.FilterRegisterLogRequest) (*v1.FilterRegisterLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterRegisterLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter register log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.RegisterLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Register
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal register log failed", "error", err)
			continue
		}

		list = append(list, &v1.RegisterLog{
			UserId:     l.ObjectID,
			AuthMethod: content.AuthMethod,
			Identifier: content.Identifier,
			RegisterIp: content.RegisterIP,
			UserAgent:  content.UserAgent,
			Timestamp:  content.Timestamp,
		})
	}

	return &v1.FilterRegisterLogReply{
		Code:    int32(responsecode.FilterRegisterLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterRegisterLogSuccess],
		Data: &v1.FilterRegisterLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterServerTrafficLog 过滤服务器流量日志
func (s *LogService) FilterServerTrafficLog(ctx context.Context, req *v1.FilterServerTrafficLogRequest) (*v1.FilterServerTrafficLogReply, error) {

	var serverID *int64
	if req.ServerId > 0 {
		serverID = &req.ServerId
	}

	logs, total, err := s.systemLogUc.FilterServerTrafficLog(ctx, req.Page, req.Size, req.Date, serverID)
	if err != nil {
		s.log.Errorw("msg", "filter server traffic log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.ServerTrafficLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Traffic
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal server traffic log failed", "error", err)
			continue
		}

		list = append(list, &v1.ServerTrafficLog{
			ServerId:  l.ObjectID,
			Upload:    content.Upload,
			Download:  content.Download,
			Timestamp: l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterServerTrafficLogReply{
		Code:    int32(responsecode.FilterServerTrafficLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterServerTrafficLogSuccess],
		Data: &v1.FilterServerTrafficLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterSubscribeLog 过滤订阅日志
func (s *LogService) FilterSubscribeLog(ctx context.Context, req *v1.FilterSubscribeLogRequest) (*v1.FilterSubscribeLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterSubscribeLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter subscribe log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.SubscribeLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Subscribe
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal subscribe log failed", "error", err)
			continue
		}

		list = append(list, &v1.SubscribeLog{
			UserId:          l.ObjectID,
			Token:           content.Token,
			UserAgent:       content.UserAgent,
			ClientIp:        content.ClientIP,
			UserSubscribeId: content.UserSubscribeId,
			Timestamp:       l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterSubscribeLogReply{
		Code:    int32(responsecode.FilterSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterSubscribeLogSuccess],
		Data: &v1.FilterSubscribeLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterResetSubscribeLog 过滤重置订阅日志
func (s *LogService) FilterResetSubscribeLog(ctx context.Context, req *v1.FilterResetSubscribeLogRequest) (*v1.FilterResetSubscribeLogReply, error) {

	var userID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}

	logs, total, err := s.systemLogUc.FilterResetSubscribeLog(ctx, req.Page, req.Size, req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter reset subscribe log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.ResetSubscribeLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.ResetSubscribe
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal reset subscribe log failed", "error", err)
			continue
		}

		list = append(list, &v1.ResetSubscribeLog{
			Type:            int32(content.Type),
			UserId:          content.UserId,
			UserSubscribeId: content.UserSubscribeId,
			OrderNo:         content.OrderNo,
			Timestamp:       content.Timestamp,
		})
	}

	return &v1.FilterResetSubscribeLogReply{
		Code:    int32(responsecode.FilterResetSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterResetSubscribeLogSuccess],
		Data: &v1.FilterResetSubscribeLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// FilterUserSubscribeTrafficLog 过滤用户订阅流量日志
func (s *LogService) FilterUserSubscribeTrafficLog(ctx context.Context, req *v1.FilterUserSubscribeTrafficLogRequest) (*v1.FilterUserSubscribeTrafficLogReply, error) {

	var userID, subscribeID *int64
	if req.UserId > 0 {
		userID = &req.UserId
	}
	if req.SubscribeId > 0 {
		subscribeID = &req.SubscribeId
	}

	logs, total, err := s.systemLogUc.FilterUserSubscribeTrafficLog(ctx, req.Page, req.Size, req.Date, userID, subscribeID)
	if err != nil {
		s.log.Errorw("msg", "filter user subscribe traffic log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.UserSubscribeTrafficLog, 0, len(logs))
	for _, l := range logs {
		var content logmodel.Traffic
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal user subscribe traffic log failed", "error", err)
			continue
		}

		list = append(list, &v1.UserSubscribeTrafficLog{
			UserId:      l.ObjectID,
			SubscribeId: 0, // 需要从content解析
			Upload:      content.Upload,
			Download:    content.Download,
			Timestamp:   l.CreatedAt.UnixMilli(),
		})
	}

	return &v1.FilterUserSubscribeTrafficLogReply{
		Code:    int32(responsecode.FilterUserSubscribeTrafficLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterUserSubscribeTrafficLogSuccess],
		Data: &v1.FilterUserSubscribeTrafficLogData{
			Total: total,
			List:  list,
		},
	}, nil
}

// ========== Traffic Log APIs ==========

// FilterTrafficLogDetails 过滤流量日志详情
func (s *LogService) FilterTrafficLogDetails(ctx context.Context, req *v1.FilterTrafficLogDetailsRequest) (*v1.FilterTrafficLogDetailsReply, error) {

	var serverID, userID, subscribeID *int64
	if req.ServerId > 0 {
		serverID = &req.ServerId
	}
	if req.UserId > 0 {
		userID = &req.UserId
	}
	if req.SubscribeId > 0 {
		subscribeID = &req.SubscribeId
	}

	// date字段留空,使用start_time/end_time需要在data层处理
	logs, total, err := s.trafficLogUc.FilterTrafficLogDetails(ctx, 0, req.Page, req.Size, "", serverID, userID, subscribeID)
	if err != nil {
		s.log.Errorw("msg", "filter traffic log details failed", "error", err)
		return nil, err
	}

	list := make([]*v1.TrafficLogDetail, 0, len(logs))
	for _, l := range logs {
		list = append(list, &v1.TrafficLogDetail{
			Id:          l.ID,
			ServerId:    l.ServerID,
			UserId:      l.UserID,
			SubscribeId: l.SubscribeID,
			Download:    l.Download,
			Upload:      l.Upload,
			Timestamp:   l.Timestamp.UnixMilli(),
		})
	}

	return &v1.FilterTrafficLogDetailsReply{
		Code:    int32(responsecode.FilterTrafficLogDetailsSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterTrafficLogDetailsSuccess],
		Data: &v1.FilterTrafficLogDetailsData{
			Total: total,
			List:  list,
		},
	}, nil
}

// ========== Log Setting APIs ==========

// GetLogSetting 获取日志设置
func (s *LogService) GetLogSetting(ctx context.Context, req *v1.GetLogSettingRequest) (*v1.GetLogSettingReply, error) {
	setting, err := s.logSettingUc.GetLogSetting(ctx, 0)
	if err != nil {
		s.log.Errorw("msg", "get log setting failed", "error", err)
		return nil, err
	}

	return &v1.GetLogSettingReply{
		Code:    int32(responsecode.GetLogSettingSuccess),
		Message: responsecode.CodeMessages[responsecode.GetLogSettingSuccess],
		Data:    setting,
	}, nil
}

// UpdateLogSetting 更新日志设置
func (s *LogService) UpdateLogSetting(ctx context.Context, req *v1.UpdateLogSettingRequest) (*v1.UpdateLogSettingReply, error) {
	// 构建 LogSetting 对象
	setting := &v1.LogSetting{
		AutoClear: req.AutoClear,
		ClearDays: req.ClearDays,
	}

	err := s.logSettingUc.UpdateLogSetting(ctx, 0, setting)
	if err != nil {
		s.log.Errorw("msg", "update log setting failed", "error", err)
		return nil, err
	}

	return &v1.UpdateLogSettingReply{
		Code:    int32(responsecode.UpdateLogSettingSuccess),
		Message: responsecode.CodeMessages[responsecode.UpdateLogSettingSuccess],
	}, nil
}
