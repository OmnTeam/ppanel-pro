package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/npanel-pro/ent"
	logmodel "github.com/OmnTeam/npanel-pro/internal/model/log"
	queueTypes "github.com/OmnTeam/npanel-pro/internal/queue/types"
	"github.com/OmnTeam/npanel-pro/pkg/constant"
	"github.com/OmnTeam/npanel-pro/pkg/sms"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// MobileSystemConfig holds mobile/SMS configuration from ProxySystem table
type MobileSystemConfig struct {
	Platform       string `json:"platform"`
	PlatformConfig string `json:"platform_config"`
}

// SendSmsHandler handles immediate SMS sending tasks
type SendSmsHandler struct {
	db     *ent.Client
	logger *log.Helper
}

// NewSendSmsHandler creates a new SMS sending handler.
func NewSendSmsHandler(db *ent.Client, logger log.Logger) *SendSmsHandler {
	return &SendSmsHandler{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask processes SMS sending task - implements asynq.Handler interface
func (h *SendSmsHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload queueTypes.SendSmsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Unmarshal payload failed: %v, payload: %s",
			err, string(task.Payload()))
		return nil // Return nil to skip retry
	}

	// Load mobile config from ProxySystem table
	mobileConfig, err := h.loadMobileConfig(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Load mobile config failed: %v", err)
		return nil
	}

	// Create SMS sender
	client, err := sms.NewSender(mobileConfig.Platform, mobileConfig.PlatformConfig)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] New send sms client failed: %v, payload: %+v",
			err, payload)

		// Create failed message log
		createSms := &logmodel.Message{

			Platform: mobileConfig.Platform,
			To:       fmt.Sprintf("+%s%s", payload.TelephoneArea, payload.Telephone),
			Subject:  constant.ParseVerifyType(payload.Type).String(),
			Content: map[string]interface{}{
				"content": payload.Content,
			},
			Status: 2, // Failed
		}
		h.logMessage(ctx, createSms)
		return nil
	}

	// Prepare message log
	createSms := &logmodel.Message{

		Platform: mobileConfig.Platform,
		To:       fmt.Sprintf("+%s%s", payload.TelephoneArea, payload.Telephone),
		Subject:  constant.ParseVerifyType(payload.Type).String(),
		Content: map[string]interface{}{
			"content": client.GetSendCodeContent(payload.Content),
		},
		Status: 2, // Default to failed
	}

	// Send SMS
	err = client.SendCode(payload.TelephoneArea, payload.Telephone, payload.Content)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Send sms failed: %v, payload: %+v",
			err, payload)

		// Log failed message
		createSms.Status = 2 // Failed
		h.logMessage(ctx, createSms)
		return nil
	}

	// Mark as sent successfully
	createSms.Status = 1
	h.logger.WithContext(ctx).Infof("[SendSmsHandler] Send sms successfully, telephone: %s, content: %+v",
		payload.Telephone, createSms.Content)

	content, err := createSms.Marshal()
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Marshal message log failed: %v, messageLog: %+v",
			err, createSms)
		return nil
	}

	err = h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeMobileMessage)).
		SetDate(time.Now().Format(time.DateOnly)).
		SetObjectID(0).
		SetContent(string(content)).
		Exec(ctx)

	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Insert sms log failed: %v, content: %s",
			err, string(content))
		return nil
	}

	return nil
}

// loadMobileConfig loads mobile/SMS configuration from ProxySystem table.
func (h *SendSmsHandler) loadMobileConfig(ctx context.Context) (*MobileSystemConfig, error) {
	config, err := loadQueueMobileConfig(ctx, h.db)
	if err != nil {
		if ent.IsNotFound(err) {
			h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Mobile auth method config not found")
			return nil, fmt.Errorf("mobile config not found")
		}
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Failed to load mobile config, error: %v", err)
		return nil, err
	}
	return &MobileSystemConfig{
		Platform:       config.Platform,
		PlatformConfig: config.PlatformConfig,
	}, nil
}

// logMessage logs the SMS message to ProxySystemLog
func (h *SendSmsHandler) logMessage(ctx context.Context, messageLog *logmodel.Message) {
	content, err := messageLog.Marshal()
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Marshal message log failed: %v, messageLog: %+v",
			err, messageLog)
		return
	}

	// Insert log to database
	_, err = h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeMobileMessage)).
		SetDate(time.Now().Format(time.DateOnly)).
		SetObjectID(0).
		SetContent(string(content)).
		Save(ctx)

	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendSmsHandler] Insert sms log failed: %v, content: %s",
			err, string(content))
	}
}
