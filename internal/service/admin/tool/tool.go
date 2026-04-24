package tool

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/tool/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/tool"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// ToolService tool service implementation
type ToolService struct {
	v1.UnimplementedToolServer

	uc *tool.ToolUseCase
}

// NewToolService create tool service
func NewToolService(uc *tool.ToolUseCase) *ToolService {
	return &ToolService{
		uc: uc,
	}
}

// GetSystemLog 获取系统日志
func (s *ToolService) GetSystemLog(ctx context.Context, req *v1.GetSystemLogRequest) (*v1.GetSystemLogReply, error) {
	logs, err := s.uc.GetSystemLog(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.GetSystemLogReply{
		Code:    int32(responsecode.AdminGetSystemLogSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetSystemLogSuccess],
		Data:    &v1.LogResponse{List: logs},
	}, nil
}

// RestartSystem 重启系统
func (s *ToolService) RestartSystem(ctx context.Context, req *v1.RestartSystemRequest) (*v1.RestartSystemReply, error) {
	if err := s.uc.RestartSystem(ctx, req); err != nil {
		return nil, err
	}

	return &v1.RestartSystemReply{
		Code:    int32(responsecode.AdminRestartSystemSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminRestartSystemSuccess],
	}, nil
}

// GetVersion 获取版本信息
func (s *ToolService) GetVersion(ctx context.Context, req *v1.GetVersionRequest) (*v1.GetVersionReply, error) {
	version, err := s.uc.GetVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.GetVersionReply{
		Code:    int32(responsecode.AdminGetVersionSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetVersionSuccess],
		Data:    version,
	}, nil
}

// QueryIPLocation 查询IP地理位置
func (s *ToolService) QueryIPLocation(ctx context.Context, req *v1.QueryIPLocationRequest) (*v1.QueryIPLocationReply, error) {
	location, err := s.uc.QueryIPLocation(ctx, req.Ip)
	if err != nil {
		return nil, err
	}

	return &v1.QueryIPLocationReply{
		Code:    int32(responsecode.AdminQueryIPLocationSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminQueryIPLocationSuccess],
		Data:    location,
	}, nil
}
