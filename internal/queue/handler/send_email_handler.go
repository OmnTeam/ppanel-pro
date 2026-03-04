package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/OmnTeam/ppanel-pro/pkg/email"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// SendEmailSystemConfig holds email configuration for verification emails from ProxySystem table
type SendEmailSystemConfig struct {
	Platform                   string `json:"platform"`
	PlatformConfig             string `json:"platform_config"`
	VerifyEmailTemplate        string `json:"verify_email_template"`
	ExpirationEmailTemplate    string `json:"expiration_email_template"`
	MaintenanceEmailTemplate   string `json:"maintenance_email_template"`
	TrafficExceedEmailTemplate string `json:"traffic_exceed_email_template"`
}

// SendEmailHandler handles immediate email sending tasks
type SendEmailHandler struct {
	db     *ent.Client
	logger *log.Helper
}

// NewSendEmailHandler creates a new email sending handler
// 所有配置从数据库根据租户ID获取
func NewSendEmailHandler(db *ent.Client, logger log.Logger) *SendEmailHandler {
	return &SendEmailHandler{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask processes email sending task - implements asynq.Handler interface
func (h *SendEmailHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload queueTypes.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Unmarshal payload failed: %v, payload: %s",
			err, string(task.Payload()))
		return nil // Return nil to skip retry
	}

	// Load email config from ProxySystem table based on tenant ID
	emailConfig, err := h.loadEmailConfig(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Load email config failed: %v", err)
		return nil
	}

	// Prepare message log
	messageLog := logmodel.Message{
		Platform: emailConfig.Platform,
		To:       payload.Email,
		Subject:  payload.Subject,
		Content:  payload.Content,
		Status:   2, // Default to failed
	}

	// Load site config from database
	siteConfig, err := h.loadSiteConfig(ctx)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Load site config failed: %v", err)
		h.logMessage(ctx, &messageLog)
		return nil
	}

	// Create email sender
	sender, err := email.NewSender(emailConfig.Platform, emailConfig.PlatformConfig, siteConfig.SiteName)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] NewSender failed: %v", err)
		h.logMessage(ctx, &messageLog)
		return nil
	}

	// Process email content based on type
	var content string
	switch payload.Type {
	case queueTypes.EmailTypeVerify:
		tpl, err := template.New("verify").Parse(emailConfig.VerifyEmailTemplate)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Parse verify template failed: %v", err)
			h.logMessage(ctx, &messageLog)
			return nil
		}

		var result bytes.Buffer
		// Type conversion for template execution
		if typeVal, ok := payload.Content["Type"].(float64); ok {
			payload.Content["Type"] = uint8(typeVal)
		}

		if err = tpl.Execute(&result, payload.Content); err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Execute verify template failed: %v, data: %+v",
				err, payload.Content)
			h.logMessage(ctx, &messageLog)
			return nil
		}
		content = result.String()

	case queueTypes.EmailTypeMaintenance:
		tpl, err := template.New("maintenance").Parse(emailConfig.MaintenanceEmailTemplate)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Parse maintenance template failed: %v", err)
			h.logMessage(ctx, &messageLog)
			return nil
		}

		var result bytes.Buffer
		if err = tpl.Execute(&result, payload.Content); err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Execute maintenance template failed: %v, template: %s, data: %+v",
				err, emailConfig.MaintenanceEmailTemplate, payload.Content)
			h.logMessage(ctx, &messageLog)
			return nil
		}
		content = result.String()

	case queueTypes.EmailTypeExpiration:
		tpl, err := template.New("expiration").Parse(emailConfig.ExpirationEmailTemplate)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Parse expiration template failed: %v", err)
			h.logMessage(ctx, &messageLog)
			return nil
		}

		var result bytes.Buffer
		if err = tpl.Execute(&result, payload.Content); err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Execute expiration template failed: %v, template: %s, data: %+v",
				err, emailConfig.ExpirationEmailTemplate, payload.Content)
			h.logMessage(ctx, &messageLog)
			return nil
		}
		content = result.String()

	case queueTypes.EmailTypeTrafficExceed:
		tpl, err := template.New("traffic_exceed").Parse(emailConfig.TrafficExceedEmailTemplate)
		if err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Parse traffic_exceed template failed: %v", err)
			h.logMessage(ctx, &messageLog)
			return nil
		}

		var result bytes.Buffer
		if err = tpl.Execute(&result, payload.Content); err != nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Execute traffic_exceed template failed: %v, template: %s, data: %+v",
				err, emailConfig.TrafficExceedEmailTemplate, payload.Content)
			h.logMessage(ctx, &messageLog)
			return nil
		}
		content = result.String()

	case queueTypes.EmailTypeCustom:
		if payload.Content == nil {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Custom email content is empty, payload: %+v", payload)
			h.logMessage(ctx, &messageLog)
			return nil
		}

		if contentStr, ok := payload.Content["content"].(string); !ok {
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Custom email content is not a string, payload: %+v", payload)
			h.logMessage(ctx, &messageLog)
			return nil
		} else {
			content = contentStr
		}

	default:
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Unsupported email type: %s, payload: %+v",
			payload.Type, payload)
		h.logMessage(ctx, &messageLog)
		return nil
	}

	// Send email
	err = sender.Send([]string{payload.Email}, payload.Subject, content)
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Send email failed: %v", err)
		h.logMessage(ctx, &messageLog)
		return nil
	}

	// Mark as sent successfully
	messageLog.Status = 1
	emailLog, err := messageLog.Marshal()
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Marshal message log failed: %v, messageLog: %+v",
			err, messageLog)
		return nil
	}

	// Insert log to database
	if err = h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeEmailMessage)).
		SetDate(time.Now().Format(time.DateOnly)).
		SetObjectID(0).
		SetContent(string(emailLog)).
		Exec(ctx); err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Insert email log failed: %v, emailLog: %s",
			err, string(emailLog))
		return nil
	}

	return nil
}

// loadEmailConfig loads email configuration from ProxySystem table based on tenant ID
func (h *SendEmailHandler) loadEmailConfig(ctx context.Context) (*SendEmailSystemConfig, error) {
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
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Email config not found, category: email, key: config")
			return nil, fmt.Errorf("email config not found")
		}
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Failed to query email config, error: %v", err)
		return nil, err
	}

	var emailConfig SendEmailSystemConfig
	if err := json.Unmarshal([]byte(config.Value), &emailConfig); err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Failed to unmarshal email config, error: %v", err)
		return nil, err
	}

	return &emailConfig, nil
}

// SendEmailSiteConfig holds site configuration
type SendEmailSiteConfig struct {
	SiteName string `json:"site_name"`
}

// loadSiteConfig loads site configuration from ProxySystem table
func (h *SendEmailHandler) loadSiteConfig(ctx context.Context) (*SendEmailSiteConfig, error) {
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
			h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Site config not found, category: site, key: config")
			return nil, fmt.Errorf("site config not found")
		}
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Failed to query site config, error: %v", err)
		return nil, err
	}

	var siteConfig SendEmailSiteConfig
	if err := json.Unmarshal([]byte(config.Value), &siteConfig); err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Failed to unmarshal site config, error: %v", err)
		return nil, err
	}

	return &siteConfig, nil
}

// logMessage logs the email message to ProxySystemLog
func (h *SendEmailHandler) logMessage(ctx context.Context, messageLog *logmodel.Message) {
	emailLog, err := messageLog.Marshal()
	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Marshal message log failed: %v, messageLog: %+v",
			err, messageLog)
		return
	}

	// Insert log to database
	_, err = h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeEmailMessage)).
		SetDate(time.Now().Format(time.DateOnly)).
		SetObjectID(0).
		SetContent(string(emailLog)).
		Save(ctx)

	if err != nil {
		h.logger.WithContext(ctx).Errorf("[SendEmailHandler] Insert email log failed: %v, emailLog: %s",
			err, string(emailLog))
	}
}
