package log

import (
	"context"
	"encoding/json"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	logbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/log"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

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
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterBalanceLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter balance log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.BalanceLog, len(logs))
	for _, l := range logs {
		var content logmodel.Balance
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal balance log failed", "error", err)
			continue
		}

		list = append(list, &v1.BalanceLog{
			Type:      int32(content.Type),
			UserId:    formatInt64(int64(l.ObjectID)),
			Amount:    formatInt64(content.Amount),
			OrderNo:   content.OrderNo,
			Balance:   formatInt64(content.Balance),
			Timestamp: formatInt64(content.Timestamp),
		})
	}

	return &v1.FilterBalanceLogReply{
		Code:    int32(responsecode.FilterBalanceLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterBalanceLogSuccess],
		Data: &v1.FilterBalanceLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterCommissionLog 过滤佣金日志
func (s *LogService) FilterCommissionLog(ctx context.Context, req *v1.FilterCommissionLogRequest) (*v1.FilterCommissionLogReply, error) {
	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterCommissionLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter commission log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.CommissionLog, len(logs))
	for _, l := range logs {
		var content logmodel.Commission
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal commission log failed", "error", err)
			continue
		}

		list = append(list, &v1.CommissionLog{
			Type:      int32(content.Type),
			UserId:    formatInt64(int64(l.ObjectID)),
			Amount:    formatInt64(content.Amount),
			OrderNo:   content.OrderNo,
			Timestamp: formatInt64(content.Timestamp),
		})
	}

	return &v1.FilterCommissionLogReply{
		Code:    int32(responsecode.FilterCommissionLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterCommissionLogSuccess],
		Data: &v1.FilterCommissionLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterEmailLog 过滤邮件日志
func (s *LogService) FilterEmailLog(ctx context.Context, req *v1.FilterEmailLogRequest) (*v1.FilterEmailLogReply, error) {

	logs, total, err := s.systemLogUc.FilterEmailLog(ctx, int32(req.Page), int32(req.Size), req.Date, req.Search)
	if err != nil {
		s.log.Errorw("msg", "filter email log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.EmailLog, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal email log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.EmailLog{
			Id:        formatInt64(int64(l.ID)),
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterEmailLogReply{
		Code:    int32(responsecode.FilterEmailLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterEmailLogSuccess],
		Data: &v1.FilterEmailLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterGiftLog 过滤赠送日志
func (s *LogService) FilterGiftLog(ctx context.Context, req *v1.FilterGiftLogRequest) (*v1.FilterGiftLogReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterGiftLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter gift log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.GiftLog, len(logs))
	for _, l := range logs {
		var content logmodel.Gift
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal gift log failed", "error", err)
			continue
		}

		list = append(list, &v1.GiftLog{
			Type:        int32(content.Type),
			UserId:      formatInt64(int64(l.ObjectID)),
			OrderNo:     content.OrderNo,
			SubscribeId: formatInt64(content.SubscribeId),
			Amount:      formatInt64(content.Amount),
			Balance:     formatInt64(content.Balance),
			Remark:      content.Remark,
			Timestamp:   formatInt64(content.Timestamp),
		})
	}

	return &v1.FilterGiftLogReply{
		Code:    int32(responsecode.FilterGiftLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterGiftLogSuccess],
		Data: &v1.FilterGiftLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterLoginLog 过滤登录日志
func (s *LogService) FilterLoginLog(ctx context.Context, req *v1.FilterLoginLogRequest) (*v1.FilterLoginLogReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterLoginLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter login log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.LoginLog, len(logs))
	for _, l := range logs {
		var content logmodel.Login
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal login log failed", "error", err)
			continue
		}

		list = append(list, &v1.LoginLog{
			UserId:    formatInt64(int64(l.ObjectID)),
			Method:    content.Method,
			LoginIp:   content.LoginIP,
			UserAgent: content.UserAgent,
			Success:   content.Success,
			Timestamp: formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterLoginLogReply{
		Code:    int32(responsecode.FilterLoginLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterLoginLogSuccess],
		Data: &v1.FilterLoginLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// GetMessageLogList 获取消息日志列表
func (s *LogService) GetMessageLogList(ctx context.Context, req *v1.GetMessageLogListRequest) (*v1.GetMessageLogListReply, error) {

	logs, total, err := s.systemLogUc.GetMessageLogList(ctx, int32(req.Page), int32(req.Size), req.Type, req.Search)
	if err != nil {
		s.log.Errorw("msg", "get message log list failed", "error", err)
		return nil, err
	}

	list := make([]*v1.MessageLog, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal message log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.MessageLog{
			Id:        formatInt64(int64(l.ID)),
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.GetMessageLogListReply{
		Code:    int32(responsecode.GetMessageLogListSuccess),
		Message: responsecode.CodeMessages[responsecode.GetMessageLogListSuccess],
		Data: &v1.GetMessageLogListData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterMobileLog 过滤手机日志
func (s *LogService) FilterMobileLog(ctx context.Context, req *v1.FilterMobileLogRequest) (*v1.FilterMobileLogReply, error) {

	logs, total, err := s.systemLogUc.FilterMobileLog(ctx, int32(req.Page), int32(req.Size), req.Date, req.Search)
	if err != nil {
		s.log.Errorw("msg", "filter mobile log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.MobileLog, len(logs))
	for _, l := range logs {
		var content logmodel.Message
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal mobile log failed", "error", err)
			continue
		}

		// Convert content map to JSON string
		contentJSON, _ := json.Marshal(content.Content)

		list = append(list, &v1.MobileLog{
			Id:        formatInt64(int64(l.ID)),
			Type:      int32(l.Type),
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   string(contentJSON),
			Status:    int32(content.Status),
			CreatedAt: formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterMobileLogReply{
		Code:    int32(responsecode.FilterMobileLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterMobileLogSuccess],
		Data: &v1.FilterMobileLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterRegisterLog 过滤注册日志
func (s *LogService) FilterRegisterLog(ctx context.Context, req *v1.FilterRegisterLogRequest) (*v1.FilterRegisterLogReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterRegisterLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter register log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.RegisterLog, len(logs))
	for _, l := range logs {
		var content logmodel.Register
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal register log failed", "error", err)
			continue
		}

		list = append(list, &v1.RegisterLog{
			UserId:     formatInt64(int64(l.ObjectID)),
			AuthMethod: content.AuthMethod,
			Identifier: content.Identifier,
			RegisterIp: content.RegisterIP,
			UserAgent:  content.UserAgent,
			Timestamp:  formatInt64(content.Timestamp),
		})
	}

	return &v1.FilterRegisterLogReply{
		Code:    int32(responsecode.FilterRegisterLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterRegisterLogSuccess],
		Data: &v1.FilterRegisterLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterServerTrafficLog 过滤服务器流量日志
func (s *LogService) FilterServerTrafficLog(ctx context.Context, req *v1.FilterServerTrafficLogRequest) (*v1.FilterServerTrafficLogReply, error) {

	var serverID *int64
	if req.ServerId != "" {
		parsedID := parseInt64(req.ServerId)
		serverID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterServerTrafficLog(ctx, int32(req.Page), int32(req.Size), req.Date, serverID)
	if err != nil {
		s.log.Errorw("msg", "filter server traffic log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.ServerTrafficLog, len(logs))
	for _, l := range logs {
		var content logmodel.Traffic
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal server traffic log failed", "error", err)
			continue
		}

		list = append(list, &v1.ServerTrafficLog{
			ServerId:  formatInt64(int64(l.ObjectID)),
			Upload:    formatInt64(content.Upload),
			Download:  formatInt64(content.Download),
			Timestamp: formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterServerTrafficLogReply{
		Code:    int32(responsecode.FilterServerTrafficLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterServerTrafficLogSuccess],
		Data: &v1.FilterServerTrafficLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterSubscribeLog 过滤订阅日志
func (s *LogService) FilterSubscribeLog(ctx context.Context, req *v1.FilterSubscribeLogRequest) (*v1.FilterSubscribeLogReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterSubscribeLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter subscribe log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.SubscribeLog, len(logs))
	for _, l := range logs {
		var content logmodel.Subscribe
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal subscribe log failed", "error", err)
			continue
		}

		list = append(list, &v1.SubscribeLog{
			UserId:          formatInt64(int64(l.ObjectID)),
			Token:           content.Token,
			UserAgent:       content.UserAgent,
			ClientIp:        content.ClientIP,
			UserSubscribeId: int32(content.UserSubscribeId),
			Timestamp:       formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterSubscribeLogReply{
		Code:    int32(responsecode.FilterSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterSubscribeLogSuccess],
		Data: &v1.FilterSubscribeLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterResetSubscribeLog 过滤重置订阅日志
func (s *LogService) FilterResetSubscribeLog(ctx context.Context, req *v1.FilterResetSubscribeLogRequest) (*v1.FilterResetSubscribeLogReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterResetSubscribeLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID)
	if err != nil {
		s.log.Errorw("msg", "filter reset subscribe log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.ResetSubscribeLog, len(logs))
	for _, l := range logs {
		var content logmodel.ResetSubscribe
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal reset subscribe log failed", "error", err)
			continue
		}

		list = append(list, &v1.ResetSubscribeLog{
			Type:            int32(content.Type),
			UserId:          formatInt64(content.UserId),
			UserSubscribeId: int32(content.UserSubscribeId),
			OrderNo:         content.OrderNo,
			Timestamp:       formatInt64(content.Timestamp),
		})
	}

	return &v1.FilterResetSubscribeLogReply{
		Code:    int32(responsecode.FilterResetSubscribeLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterResetSubscribeLogSuccess],
		Data: &v1.FilterResetSubscribeLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// FilterUserSubscribeTrafficLog 过滤用户订阅流量日志
func (s *LogService) FilterUserSubscribeTrafficLog(ctx context.Context, req *v1.FilterUserSubscribeTrafficLogRequest) (*v1.FilterUserSubscribeTrafficLogReply, error) {

	var userID, subscribeID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}
	if req.SubscribeId != "" {
		parsedID := parseInt64(req.SubscribeId)
		subscribeID = &parsedID
	}

	logs, total, err := s.systemLogUc.FilterUserSubscribeTrafficLog(ctx, int32(req.Page), int32(req.Size), req.Date, userID, subscribeID)
	if err != nil {
		s.log.Errorw("msg", "filter user subscribe traffic log failed", "error", err)
		return nil, err
	}

	list := make([]*v1.UserSubscribeTrafficLog, len(logs))
	for _, l := range logs {
		var content logmodel.Traffic
		if err := json.Unmarshal([]byte(l.Content), &content); err != nil {
			s.log.Warnw("msg", "unmarshal user subscribe traffic log failed", "error", err)
			continue
		}

		list = append(list, &v1.UserSubscribeTrafficLog{
			UserId:      formatInt64(int64(l.ObjectID)),
			SubscribeId: "", // 需要从content解析
			Upload:      formatInt64(content.Upload),
			Download:    formatInt64(content.Download),
			Timestamp:   formatInt64(l.CreatedAt.UnixMilli()),
		})
	}

	return &v1.FilterUserSubscribeTrafficLogReply{
		Code:    int32(responsecode.FilterUserSubscribeTrafficLogSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterUserSubscribeTrafficLogSuccess],
		Data: &v1.FilterUserSubscribeTrafficLogData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// ========== Traffic Log APIs ==========

// FilterTrafficLogDetails 过滤流量日志详情
func (s *LogService) FilterTrafficLogDetails(ctx context.Context, req *v1.FilterTrafficLogDetailsRequest) (*v1.FilterTrafficLogDetailsReply, error) {

	var serverID, userID, subscribeID *int64
	if req.ServerId != "" {
		parsedID := parseInt64(req.ServerId)
		serverID = &parsedID
	}
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}
	if req.SubscribeId != "" {
		parsedID := parseInt64(req.SubscribeId)
		subscribeID = &parsedID
	}

	// date字段留空,使用start_time/end_time需要在data层处理
	logs, total, err := s.trafficLogUc.FilterTrafficLogDetails(ctx, int32(req.Page), int32(req.Size), "", serverID, userID, subscribeID)
	if err != nil {
		s.log.Errorw("msg", "filter traffic log details failed", "error", err)
		return nil, err
	}

	list := make([]*v1.TrafficLogDetail, len(logs))
	for _, l := range logs {
		list = append(list, &v1.TrafficLogDetail{
			Id:          formatInt64(int64(l.ID)),
			ServerId:    formatInt64(int64(l.ServerID)),
			UserId:      formatInt64(int64(l.UserID)),
			SubscribeId: formatInt64(int64(l.SubscribeID)),
			Download:    formatInt64(int64(l.Download)),
			Upload:      formatInt64(int64(l.Upload)),
			Timestamp:   formatInt64(l.Timestamp.UnixMilli()),
		})
	}

	return &v1.FilterTrafficLogDetailsReply{
		Code:    int32(responsecode.FilterTrafficLogDetailsSuccess),
		Message: responsecode.CodeMessages[responsecode.FilterTrafficLogDetailsSuccess],
		Data: &v1.FilterTrafficLogDetailsData{
			Total: int32(total),
			List:  list,
		},
	}, nil
}

// ========== Log Setting APIs ==========

// GetLogSetting 获取日志设置
func (s *LogService) GetLogSetting(ctx context.Context, req *v1.GetLogSettingRequest) (*v1.GetLogSettingReply, error) {
	setting, err := s.logSettingUc.GetLogSetting(ctx)
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

	err := s.logSettingUc.UpdateLogSetting(ctx, setting)
	if err != nil {
		s.log.Errorw("msg", "update log setting failed", "error", err)
		return nil, err
	}

	return &v1.UpdateLogSettingReply{
		Code:    int32(responsecode.UpdateLogSettingSuccess),
		Message: responsecode.CodeMessages[responsecode.UpdateLogSettingSuccess],
	}, nil
}
