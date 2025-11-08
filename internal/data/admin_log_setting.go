package data

import (
	"context"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	logbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/log"
	"github.com/go-kratos/kratos/v2/log"
)

// 日志设置配置键（复刻原项目：updateLogSettingLogic.go）
// ⚠️ 注意：key名称必须与原项目保持一致（使用PascalCase，字段名即key）
const (
	LogSettingAutoClear = "AutoClear" // 是否自动清理日志
	LogSettingClearDays = "ClearDays" // 清理天数
)

type adminLogSettingRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminLogSettingRepo 创建日志设置仓库
func NewAdminLogSettingRepo(data *Data, logger log.Logger) logbiz.LogSettingRepo {
	return &adminLogSettingRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetLogSetting 获取日志设置
// ⚠️ 复刻原项目：getLogSettingLogic.go:27-37
func (r *adminLogSettingRepo) GetLogSetting(ctx context.Context, tenantID int64) (*v1.LogSetting, error) {
	// 查询日志清理相关配置（复刻原项目 line 28）
	configs, err := r.data.db.ProxySystem.
		Query().
		Where(
			proxysystem.CategoryEQ("log"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 构建配置 map
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.Key] = cfg.Value
	}

	// 解析 AutoClear（默认 true）
	autoClear := true
	if val, exists := configMap[LogSettingAutoClear]; exists {
		if b, err := strconv.ParseBool(val); err == nil {
			autoClear = b
		}
	}

	// 解析 ClearDays（默认 7）
	clearDays := int64(7)
	if val, exists := configMap[LogSettingClearDays]; exists {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			clearDays = i
		}
	}

	// 复刻原项目：使用反射映射到结构体（line 35）
	setting := &v1.LogSetting{
		AutoClear: autoClear,
		ClearDays: clearDays,
	}

	return setting, nil
}

// UpdateLogSetting 更新日志设置
// ⚠️ 完整复刻原项目：updateLogSettingLogic.go:33-63
func (r *adminLogSettingRepo) UpdateLogSetting(ctx context.Context, tenantID int64, setting *v1.LogSetting) error {
	// 准备配置项（复刻原项目：字段名即key）
	configs := map[string]string{
		LogSettingAutoClear: strconv.FormatBool(setting.AutoClear),
		LogSettingClearDays: strconv.FormatInt(setting.ClearDays, 10),
	}

	// ✅ 使用事务确保所有配置项更新的原子性（复刻原项目 line 37）
	err := r.data.db.TX(ctx, func(tx *ent.Tx) error {
		// 逐个更新配置项（复刻原项目 line 39-48）
		for key, value := range configs {
			// 原项目直接UPDATE，假设记录已存在
			// 新项目改进：先查询，支持UPSERT
			existing, err := tx.ProxySystem.
				Query().
				Where(
					proxysystem.CategoryEQ("log"),
					proxysystem.KeyEQ(key),
				).
				Only(ctx)

			if err != nil && !ent.IsNotFound(err) {
				r.log.Errorf("Failed to query log setting %s: %v", key, err)
				return err
			}

			if existing != nil {
				// 更新现有记录（复刻原项目 line 45）
				_, err = tx.ProxySystem.UpdateOneID(existing.ID).
					SetValue(value).
					Save(ctx)
				if err != nil {
					r.log.Errorf("Failed to update log setting %s: %v", key, err)
					return err
				}
			} else {
				// 创建新记录（新项目改进：支持首次创建）
				configType := "bool"
				if key == LogSettingClearDays {
					configType = "int"
				}

				_, err = tx.ProxySystem.
					Create().
					SetCategory("log").
					SetKey(key).
					SetValue(value).
					SetType(configType).
					SetDesc("Log setting").
					Save(ctx)
				if err != nil {
					r.log.Errorf("Failed to create log setting %s: %v", key, err)
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// TODO: 更新运行时配置（复刻原项目 line 57-60）
	// l.svcCtx.Config.Log = config.Log{
	//     AutoClear: setting.AutoClear,
	//     ClearDays: setting.ClearDays,
	// }
	// 注意：需要有全局配置对象才能实现此功能

	return nil
}
