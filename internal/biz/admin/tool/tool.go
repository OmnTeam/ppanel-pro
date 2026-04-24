package tool

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/tool/v1"
	systembiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/system"
	"github.com/OmnTeam/ppanel-pro/pkg/ip"
	ppanelLogger "github.com/OmnTeam/ppanel-pro/pkg/logger"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToolUseCase tool use case
type ToolUseCase struct {
	log      *log.Helper
	systemUC *systembiz.SystemUsecase
}

// NewToolUseCase creates a new tool use case
func NewToolUseCase(logger log.Logger, systemUC *systembiz.SystemUsecase) *ToolUseCase {
	return &ToolUseCase{
		log:      log.NewHelper(log.With(logger, "module", "biz/admin/tool")),
		systemUC: systemUC,
	}
}

// GetSystemLog gets system logs
func (uc *ToolUseCase) GetSystemLog(ctx context.Context, req *v1.GetSystemLogRequest) ([]*structpb.Struct, error) {
	lines, err := ppanelLogger.ReadLastNLines("./logs", 50)
	if err != nil {
		return nil, err
	}

	logs := make([]*structpb.Struct, 0, len(lines))
	for _, line := range lines {
		item, err := structpb.NewStruct(map[string]any{
			"line": line,
		})
		if err != nil {
			continue
		}
		logs = append(logs, item)
	}

	return logs, nil
}

// RestartSystem restarts the system
func (uc *ToolUseCase) RestartSystem(ctx context.Context, req *v1.RestartSystemRequest) error {
	uc.log.Infof("System restart requested")
	return nil
}

// GetVersion gets version information
func (uc *ToolUseCase) GetVersion(ctx context.Context) (*v1.VersionResponse, error) {
	version := "unknown"
	if uc.systemUC != nil {
		if module, err := uc.systemUC.GetSystemModule(ctx); err == nil && strings.TrimSpace(module.ServiceVersion) != "" {
			version = module.ServiceVersion
		}
	}

	return &v1.VersionResponse{
		Version: fmt.Sprintf("%s(%s) Develop", version, "unknown"),
	}, nil
}

// QueryIPLocation queries IP geolocation
func (uc *ToolUseCase) QueryIPLocation(ctx context.Context, queryIP string) (*v1.QueryIPLocationResponse, error) {
	location, err := ip.GetRegionByIp(queryIP)
	if err != nil {
		return nil, err
	}
	return &v1.QueryIPLocationResponse{
		Country: location.Country,
		Region:  location.Region,
		City:    location.City,
	}, nil
}
