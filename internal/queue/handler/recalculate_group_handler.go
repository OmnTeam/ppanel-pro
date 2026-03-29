package handler

import (
	"context"
	"strings"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// RecalculateGroupHandler 分组重算任务处理器
type RecalculateGroupHandler struct {
	db                *ent.Client
	groupRecalculator groupRecalculator
	logger            *log.Helper
}

// NewRecalculateGroupHandler 创建分组重算任务处理器
func NewRecalculateGroupHandler(db *ent.Client, groupRecalculator groupRecalculator, logger log.Logger) *RecalculateGroupHandler {
	return &RecalculateGroupHandler{
		db:                db,
		groupRecalculator: groupRecalculator,
		logger:            log.NewHelper(logger),
	}
}

// ProcessTask 处理分组重算任务
func (h *RecalculateGroupHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	h.logger.Info("[RecalculateGroup] Starting scheduled group recalculation")

	// 1. 检查分组管理是否启用
	enabledConfig, err := h.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("group"),
			proxysystem.KeyEQ("enabled"),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			h.logger.Debug("[RecalculateGroup] Group enabled config not found, skipping")
			return nil
		}
		h.logger.Errorf("[RecalculateGroup] Failed to read group enabled config: %v", err)
		return err
	}

	// 如果未启用，跳过执行
	enabledValue := strings.TrimSpace(enabledConfig.Value)
	if enabledValue != "true" && enabledValue != "1" {
		h.logger.Debug("[RecalculateGroup] Group management is not enabled, skipping")
		return nil
	}

	// 2. 获取分组模式
	modeConfig, err := h.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("group"),
			proxysystem.KeyEQ("mode"),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			h.logger.Debug("[RecalculateGroup] Group mode config not found, using default 'average'")
			return nil
		}
		h.logger.Errorf("[RecalculateGroup] Failed to read group mode config: %v", err)
		return err
	}

	mode := strings.TrimSpace(modeConfig.Value)
	if mode == "" {
		mode = "average" // 默认模式
	}

	// 3. 只在 traffic 模式下执行
	if mode != "traffic" {
		h.logger.Debugf("[RecalculateGroup] Group mode is not 'traffic' (current: %s), skipping", mode)
		return nil
	}

	if h.groupRecalculator == nil {
		h.logger.Warn("[RecalculateGroup] Group recalculator is not configured, skipping")
		return nil
	}

	h.logger.Info("[RecalculateGroup] Executing traffic-based grouping")

	historyID, err := h.groupRecalculator.RecalculateGroup(ctx, mode, "scheduled")
	if err != nil {
		h.logger.Errorf("[RecalculateGroup] Failed to execute traffic grouping: %v", err)
		return err
	}

	h.logger.Infof("[RecalculateGroup] Successfully completed traffic-based grouping: history_id=%d", historyID)
	return nil
}
