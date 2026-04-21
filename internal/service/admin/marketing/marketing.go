package marketing

import (
	"context"
	"strconv"
	"strings"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/marketing/v1"
	marketingbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/marketing"
	taskmodel "github.com/OmnTeam/ppanel-pro/internal/model/task"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func parseStringID(s string) (int, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return int(val), nil
}

// MarketingService 营销服务
type MarketingService struct {
	v1.UnimplementedMarketingServiceServer

	uc  *marketingbiz.MarketingUsecase
	log *log.Helper
}

// NewMarketingService 创建营销服务
func NewMarketingService(uc *marketingbiz.MarketingUsecase, logger log.Logger) *MarketingService {
	return &MarketingService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/admin/marketing")),
	}
}

// ========== Email Task Methods ==========

// CreateBatchSendEmailTask 创建批量发送邮件任务
func (s *MarketingService) CreateBatchSendEmailTask(ctx context.Context, req *v1.CreateBatchSendEmailTaskRequest) (*v1.CreateBatchSendEmailTaskReply, error) {
	if req.Subject == "" {
		return nil, responsecode.ErrInvalidParam()
	}
	if req.Content == "" {
		return nil, responsecode.ErrInvalidParam()
	}

	err := s.uc.CreateBatchSendEmailTask(ctx, req.Subject, req.Content, req.Scope,
		int64(req.RegisterStartTime), int64(req.RegisterEndTime), req.Additional, int(req.Scheduled), req.Interval, int(req.Limit))
	if err != nil {
		s.log.Errorw("msg", "create batch send email task failed", "error", err)
		return nil, responsecode.ErrTaskCreateFailed()
	}

	return &v1.CreateBatchSendEmailTaskReply{
		Code:    int32(responsecode.AdminCreateBatchSendEmailTaskSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateBatchSendEmailTaskSuccess],
		Data: &v1.CreateBatchSendEmailTaskData{
			Success: true,
		},
	}, nil
}

// GetBatchSendEmailTaskList 获取批量发送邮件任务列表
func (s *MarketingService) GetBatchSendEmailTaskList(ctx context.Context, req *v1.GetBatchSendEmailTaskListRequest) (*v1.GetBatchSendEmailTaskListReply, error) {
	var scope, status *int32
	if req.Scope != nil {
		scope = req.Scope
	}
	if req.Status != nil {
		status = req.Status
	}

	tasks, total, err := s.uc.GetBatchSendEmailTaskList(ctx, int32(req.Page), int32(req.Size), scope, status)
	if err != nil {
		s.log.Errorw("msg", "get batch send email task list failed", "error", err)
		return nil, responsecode.ErrTaskListFailed()
	}

	// 转换为 proto message
	list := make([]*v1.BatchSendEmailTask, 0, len(tasks))
	for _, task := range tasks {
		// 解析 scope
		scopeInfo, err := taskmodel.UnmarshalEmailScope(task.Scope)
		if err != nil {
			s.log.Errorw("msg", "unmarshal email scope failed", "error", err)
			continue
		}

		// 解析 content
		contentInfo, err := taskmodel.UnmarshalEmailContent(task.Content)
		if err != nil {
			s.log.Errorw("msg", "unmarshal email content failed", "error", err)
			continue
		}

		list = append(list, &v1.BatchSendEmailTask{
			Id:                formatInt64(int64(task.ID)),
			Subject:           contentInfo.Subject,
			Content:           contentInfo.Content,
			Recipients:        strings.Join(scopeInfo.Recipients, "\n"),
			Scope:             int32(scopeInfo.Type),
			RegisterStartTime: scopeInfo.RegisterStartTime,
			RegisterEndTime:   scopeInfo.RegisterEndTime,
			Additional:        strings.Join(scopeInfo.Additional, "\n"),
			Scheduled:         int64(scopeInfo.Scheduled),
			Interval:          int32(scopeInfo.Interval),
			Limit:             int64(scopeInfo.Limit),
			Status:            int32(task.Status),
			Errors:            task.Errors,
			Total:             int64(task.Total),
			Current:           int64(task.Current),
			CreatedAt:         task.CreatedAt.UnixMilli(),
			UpdatedAt:         task.UpdatedAt.UnixMilli(),
		})
	}

	return &v1.GetBatchSendEmailTaskListReply{
		Code:    int32(responsecode.AdminGetBatchSendEmailTaskListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetBatchSendEmailTaskListSuccess],
		Data: &v1.GetBatchSendEmailTaskListData{
			Total: total,
			List:  list,
		},
	}, nil
}

// StopBatchSendEmailTask 停止批量发送邮件任务
func (s *MarketingService) StopBatchSendEmailTask(ctx context.Context, req *v1.StopBatchSendEmailTaskRequest) (*v1.StopBatchSendEmailTaskReply, error) {
	if req.Id == "" {
		return nil, responsecode.ErrInvalidParam()
	}

	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.uc.StopBatchSendEmailTask(ctx, id)
	if err != nil {
		s.log.Errorw("msg", "stop batch send email task failed", "error", err)
		return nil, responsecode.ErrTaskUpdateFailed()
	}

	return &v1.StopBatchSendEmailTaskReply{
		Code:    int32(responsecode.AdminStopBatchSendEmailTaskSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminStopBatchSendEmailTaskSuccess],
		Data: &v1.StopBatchSendEmailTaskData{
			Success: true,
		},
	}, nil
}

// GetPreSendEmailCount 获取预发送邮件数量
func (s *MarketingService) GetPreSendEmailCount(ctx context.Context, req *v1.GetPreSendEmailCountRequest) (*v1.GetPreSendEmailCountReply, error) {
	count, err := s.uc.GetPreSendEmailCount(ctx, req.Scope, int(int64(req.RegisterStartTime)), int(int64(req.RegisterEndTime)))
	if err != nil {
		s.log.Errorw("msg", "get pre send email count failed", "error", err)
		return nil, responsecode.ErrTaskQueryFailed()
	}

	return &v1.GetPreSendEmailCountReply{
		Code:    int32(responsecode.AdminGetPreSendEmailCountSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetPreSendEmailCountSuccess],
		Data: &v1.GetPreSendEmailCountData{
			Count: count,
		},
	}, nil
}

// GetBatchSendEmailTaskStatus 获取批量发送邮件任务状态
func (s *MarketingService) GetBatchSendEmailTaskStatus(ctx context.Context, req *v1.GetBatchSendEmailTaskStatusRequest) (*v1.GetBatchSendEmailTaskStatusReply, error) {
	if req.Id == "" {
		return nil, responsecode.ErrInvalidParam()
	}

	id, err := parseStringID(req.Id)
	if err != nil {
		return nil, err
	}

	task, err := s.uc.GetBatchSendEmailTaskStatus(ctx, id)
	if err != nil {
		s.log.Errorw("msg", "get batch send email task status failed", "error", err)
		return nil, responsecode.ErrTaskQueryFailed()
	}

	return &v1.GetBatchSendEmailTaskStatusReply{
		Code:    int32(responsecode.AdminGetBatchSendEmailTaskStatusSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetBatchSendEmailTaskStatusSuccess],
		Data: &v1.GetBatchSendEmailTaskStatusData{
			Status:  int32(task.Status),
			Current: int64(task.Current),
			Total:   int64(task.Total),
			Errors:  task.Errors,
		},
	}, nil
}

// ========== Quota Task Methods ==========

// CreateQuotaTask 创建配额任务
func (s *MarketingService) CreateQuotaTask(ctx context.Context, req *v1.CreateQuotaTaskRequest) (*v1.CreateQuotaTaskReply, error) {
	var isActive *bool
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	// 转换subscribers从[]int64到[]int
	subscribersInt := make([]int, len(req.Subscribers))
	for i, v := range req.Subscribers {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		subscribersInt[i] = parsed
	}

	err := s.uc.CreateQuotaTask(ctx, subscribersInt, isActive,
		req.StartTime, req.EndTime, req.ResetTraffic, req.Days, req.GiftType, int(req.GiftValue))
	if err != nil {
		s.log.Errorw("msg", "create quota task failed", "error", err)
		return nil, responsecode.ErrTaskCreateFailed()
	}

	return &v1.CreateQuotaTaskReply{
		Code:    int32(responsecode.AdminCreateQuotaTaskSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateQuotaTaskSuccess],
		Data: &v1.CreateQuotaTaskData{
			Success: true,
		},
	}, nil
}

// QueryQuotaTaskPreCount 查询配额任务预计数量
func (s *MarketingService) QueryQuotaTaskPreCount(ctx context.Context, req *v1.QueryQuotaTaskPreCountRequest) (*v1.QueryQuotaTaskPreCountReply, error) {
	var isActive *bool
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	// 转换subscribers从[]int64到[]int
	subscribersInt := make([]int, len(req.Subscribers))
	for i, v := range req.Subscribers {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		subscribersInt[i] = parsed
	}

	count, err := s.uc.QueryQuotaTaskPreCount(ctx, subscribersInt, isActive, int(req.StartTime), int(req.EndTime))
	if err != nil {
		s.log.Errorw("msg", "query quota task pre count failed", "error", err)
		return nil, responsecode.ErrTaskQueryFailed()
	}

	return &v1.QueryQuotaTaskPreCountReply{
		Code:    int32(responsecode.AdminQueryQuotaTaskPreCountSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminQueryQuotaTaskPreCountSuccess],
		Data: &v1.QueryQuotaTaskPreCountData{
			Count: count,
		},
	}, nil
}

// QueryQuotaTaskList 查询配额任务列表
func (s *MarketingService) QueryQuotaTaskList(ctx context.Context, req *v1.QueryQuotaTaskListRequest) (*v1.QueryQuotaTaskListReply, error) {
	var status *int32
	if req.Status != nil {
		status = req.Status
	}

	tasks, total, err := s.uc.QueryQuotaTaskList(ctx, int32(req.Page), int32(req.Size), status)
	if err != nil {
		s.log.Errorw("msg", "query quota task list failed", "error", err)
		return nil, responsecode.ErrTaskListFailed()
	}

	// 转换为 proto message
	list := make([]*v1.QuotaTask, 0, len(tasks))
	for _, task := range tasks {
		// 解析 scope
		scopeInfo, err := taskmodel.UnmarshalQuotaScope(task.Scope)
		if err != nil {
			s.log.Errorw("msg", "unmarshal quota scope failed", "error", err)
			continue
		}

		// 解析 content
		contentInfo, err := taskmodel.UnmarshalQuotaContent(task.Content)
		if err != nil {
			s.log.Errorw("msg", "unmarshal quota content failed", "error", err)
			continue
		}

		list = append(list, &v1.QuotaTask{
			Id:           formatInt64(int64(task.ID)),
			Subscribers:  formatInt64Slice(scopeInfo.Subscribers),
			IsActive:     scopeInfo.IsActive,
			StartTime:    scopeInfo.StartTime,
			EndTime:      scopeInfo.EndTime,
			ResetTraffic: contentInfo.ResetTraffic,
			Days:         int64(contentInfo.Days),
			GiftType:     int32(contentInfo.GiftType),
			GiftValue:    int64(contentInfo.GiftValue),
			Objects:      formatInt64Slice(scopeInfo.Objects),
			Status:       int32(task.Status),
			Total:        int64(task.Total),
			Current:      int64(task.Current),
			Errors:       task.Errors,
			CreatedAt:    task.CreatedAt.UnixMilli(),
			UpdatedAt:    task.UpdatedAt.UnixMilli(),
		})
	}

	return &v1.QueryQuotaTaskListReply{
		Code:    int32(responsecode.AdminQueryQuotaTaskListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminQueryQuotaTaskListSuccess],
		Data: &v1.QueryQuotaTaskListData{
			Total: total,
			List:  list,
		},
	}, nil
}

func formatInt64Slice(values []int64) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strconv.FormatInt(value, 10))
	}
	return result
}
