package log

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	"github.com/go-kratos/kratos/v2/log"
)

// LogSettingRepo 日志设置仓库接口
type LogSettingRepo interface {
	// GetLogSetting 获取日志设置
	GetLogSetting(ctx context.Context) (*v1.LogSetting, error)

	// UpdateLogSetting 更新日志设置
	UpdateLogSetting(ctx context.Context, setting *v1.LogSetting) error
}

// LogSettingUsecase 日志设置用例
type LogSettingUsecase struct {
	repo LogSettingRepo
	log  *log.Helper
}

// NewLogSettingUsecase 创建日志设置用例
func NewLogSettingUsecase(repo LogSettingRepo, logger log.Logger) *LogSettingUsecase {
	return &LogSettingUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/log/setting")),
	}
}

// GetLogSetting 获取日志设置
func (uc *LogSettingUsecase) GetLogSetting(ctx context.Context) (*v1.LogSetting, error) {
	return uc.repo.GetLogSetting(ctx)
}

// UpdateLogSetting 更新日志设置
func (uc *LogSettingUsecase) UpdateLogSetting(ctx context.Context, setting *v1.LogSetting) error {
	return uc.repo.UpdateLogSetting(ctx, setting)
}
