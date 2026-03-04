package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxytask"
	taskmodel "github.com/OmnTeam/ppanel-pro/internal/model/task"
	"github.com/OmnTeam/ppanel-pro/pkg/email"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// EmailSystemConfig holds email configuration from ProxySystem table
type EmailSystemConfig struct {
	Platform       string `json:"platform"`
	PlatformConfig string `json:"platform_config"`
}

// SiteConfig holds site configuration from ProxySystem table
type SiteConfig struct {
	SiteName string `json:"site_name"`
}

// EmailErrorInfo stores email sending error information
type EmailErrorInfo struct {
	Error string `json:"error"`
	Email string `json:"email"`
	Time  int64  `json:"time"`
}

// BatchEmailHandler 批量邮件任务处理器
type BatchEmailHandler struct {
	db  *ent.Client
	log *log.Helper
}

// NewBatchEmailHandler 创建批量邮件任务处理器
// 所有配置从数据库根据租户ID获取
func NewBatchEmailHandler(db *ent.Client, logger log.Logger) *BatchEmailHandler {
	return &BatchEmailHandler{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// ProcessTask 处理批量邮件任务 - 实现 asynq.Handler 接口
func (h *BatchEmailHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// 解析任务负载
	payload := task.Payload()
	if len(payload) == 0 {
		h.log.Error("[BatchEmailHandler] ProcessTask failed: empty payload")
		return asynq.SkipRetry
	}

	// 转换获取任务ID
	taskID, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] ProcessTask failed: invalid task ID, error: %v, payload: %s",
			err, string(payload))
		return asynq.SkipRetry
	}

	// 查询任务信息
	taskInfo, err := h.db.ProxyTask.Query().
		Where(proxytask.ID(taskID)).
		Only(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] ProcessTask failed: find task error, taskID: %d, error: %v",
			taskID, err)
		return asynq.SkipRetry
	}

	// 检查任务状态 - 如果不是待处理状态则跳过
	if taskInfo.Status != int8(taskmodel.StatusPending) {
		h.log.WithContext(ctx).Infof("[BatchEmailHandler] ProcessTask skipped: task already processed, taskID: %d, status: %d",
			taskID, taskInfo.Status)
		return nil
	}

	h.log.WithContext(ctx).Infof("[BatchEmailHandler] Starting to process email task %d", taskID)

	// 解析scope
	scope, err := taskmodel.UnmarshalEmailScope(taskInfo.Scope)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to unmarshal scope: %v", err)
		return asynq.SkipRetry
	}

	// 解析content
	content, err := taskmodel.UnmarshalEmailContent(taskInfo.Content)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to unmarshal content: %v", err)
		return asynq.SkipRetry
	}

	// 合并收件人列表
	var recipients []string
	if len(scope.Recipients) > 0 {
		recipients = append(recipients, scope.Recipients...)
	}
	if len(scope.Additional) > 0 {
		recipients = append(recipients, scope.Additional...)
	}
	// 去重
	recipients = tool.RemoveDuplicateElements(recipients...)

	if len(recipients) == 0 {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] No valid recipients found for task %d", taskID)
		// 标记任务为完成
		_ = h.db.ProxyTask.UpdateOneID(taskID).
			SetStatus(int8(taskmodel.StatusCompleted)).
			Exec(ctx)
		return nil
	}

	// 加载邮件配置
	emailConfig, err := h.loadEmailConfig(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to load email config: %v", err)
		return asynq.SkipRetry
	}

	// 从数据库加载站点配置
	siteConfig, err := h.loadSiteConfig(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to load site config: %v", err)
		return asynq.SkipRetry
	}
	siteName := siteConfig.SiteName

	// 使用租户配置创建邮件发送器
	sender, err := email.NewSender(emailConfig.Platform, emailConfig.PlatformConfig, siteName)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to create email sender: %v", err)
		return asynq.SkipRetry
	}

	// 更新任务状态为进行中
	err = h.db.ProxyTask.UpdateOneID(taskID).
		SetStatus(int8(taskmodel.StatusInProgress)).
		Exec(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to update task status to InProgress: %v", err)
		return err
	}

	// 设置发送间隔时间
	var intervalTime time.Duration
	if scope.Interval == 0 {
		intervalTime = 1 * time.Second
	} else {
		intervalTime = time.Duration(scope.Interval) * time.Second
	}

	// 批量发送邮件
	var errors []EmailErrorInfo
	var count uint64

	for _, recipient := range recipients {
		select {
		case <-ctx.Done():
			h.log.WithContext(ctx).Infof("[BatchEmailHandler] Worker stopped by context cancellation, taskID: %d", taskID)
			// 更新任务状态
			_ = h.db.ProxyTask.UpdateOneID(taskID).
				SetCurrent(uint32(count)).
				Exec(ctx)
			return nil
		default:
		}

		// 发送邮件
		if err := sender.Send([]string{recipient}, content.Subject, content.Content); err != nil {
			h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to send email to %s: %v", recipient, err)
			errors = append(errors, EmailErrorInfo{
				Error: err.Error(),
				Email: recipient,
				Time:  time.Now().Unix(),
			})
		}

		count++

		// 更新任务进度
		errorJSON := ""
		if len(errors) > 0 {
			errBytes, _ := json.Marshal(errors)
			errorJSON = string(errBytes)
		}

		err = h.db.ProxyTask.UpdateOneID(taskID).
			SetCurrent(uint32(count)).
			SetErrors(errorJSON).
			Exec(ctx)
		if err != nil {
			h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to update task progress: %v", err)
		}

		// 发送间隔
		time.Sleep(intervalTime)
	}

	// 标记任务为完成
	err = h.db.ProxyTask.UpdateOneID(taskID).
		SetStatus(int8(taskmodel.StatusCompleted)).
		SetCurrent(uint32(count)).
		Exec(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to update task status to Completed: %v", err)
		return err
	}

	h.log.WithContext(ctx).Infof("[BatchEmailHandler] Successfully completed email task %d, sent %d/%d emails",
		taskID, count-uint64(len(errors)), count)
	return nil
}

// loadEmailConfig loads email configuration from ProxySystem table
func (h *BatchEmailHandler) loadEmailConfig(ctx context.Context) (*EmailSystemConfig, error) {
	// Query email config from ProxySystem table
	// Category: "email", Key: "config"
	config, err := h.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("email"),
			proxysystem.KeyEQ("config"),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Email config not found, category: email, key: config")
			return nil, fmt.Errorf("email config not found")
		}
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to query email config, error: %v", err)
		return nil, err
	}

	var emailConfig EmailSystemConfig
	if err := json.Unmarshal([]byte(config.Value), &emailConfig); err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to unmarshal email config, error: %v", err)
		return nil, err
	}

	return &emailConfig, nil
}

// loadSiteConfig loads site configuration from ProxySystem table
func (h *BatchEmailHandler) loadSiteConfig(ctx context.Context) (*SiteConfig, error) {
	// Query site config from ProxySystem table
	// Category: "site", Key: "config"
	config, err := h.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("site"),
			proxysystem.KeyEQ("config"),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Site config not found, category: site, key: config")
			return nil, fmt.Errorf("site config not found")
		}
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to query site config, error: %v", err)
		return nil, err
	}

	var siteConfig SiteConfig
	if err := json.Unmarshal([]byte(config.Value), &siteConfig); err != nil {
		h.log.WithContext(ctx).Errorf("[BatchEmailHandler] Failed to unmarshal site config, error: %v", err)
		return nil, err
	}

	return &siteConfig, nil
}
