package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/tool/v1"
	systembiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/system"
	"github.com/OmnTeam/npanel-pro/pkg/ip"
	npanelLogger "github.com/OmnTeam/npanel-pro/pkg/logger"
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
	lines, err := npanelLogger.ReadLastNLines("./logs", 50)
	if err != nil {
		return nil, err
	}

	logs := make([]*structpb.Struct, 0, len(lines))
	for _, line := range lines {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		item, err := structpb.NewStruct(payload)
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
	buildTime := "unknown"

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if candidate := strings.TrimSpace(buildInfo.Main.Version); candidate != "" && candidate != "(devel)" {
			version = candidate
		}
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, strings.TrimSpace(setting.Value)); err == nil {
					buildTime = t.Format("2006-01-02 15:04:05")
				}
			case "vcs.revision":
				if version == "unknown" {
					value := strings.TrimSpace(setting.Value)
					if value != "" {
						version = value
					}
				}
			}
		}
	}

	if uc.systemUC != nil {
		if module, err := uc.systemUC.GetSystemModule(ctx); err == nil && strings.TrimSpace(module.ServiceVersion) != "" {
			version = module.ServiceVersion
		}
	}

	version = strings.TrimSpace(version)
	if version == "" || version == "unknown version" {
		version = "unknown"
	}
	if strings.HasPrefix(version, "v") {
		return &v1.VersionResponse{
			Version: fmt.Sprintf("%s(%s)", strings.TrimPrefix(version, "v"), buildTime),
		}, nil
	}

	return &v1.VersionResponse{
		Version: fmt.Sprintf("%s(%s) Develop", version, buildTime),
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
