package authmethod

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/internal/model/auth"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/email"
	"github.com/OmnTeam/ppanel-pro/pkg/sms"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
)

// AuthMethod 认证方法
type AuthMethod struct {
	ID        int64
	Method    string
	Config    string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpdateAuthMethodRequest 更新认证方法请求
type UpdateAuthMethodRequest struct {
	Method  string
	Config  interface{}
	Enabled bool
}

// AuthMethodRepo 认证方法仓储接口
type AuthMethodRepo interface {
	FindByMethod(ctx context.Context, method string) (*AuthMethod, error)
	Update(ctx context.Context, auth *AuthMethod) (*AuthMethod, error)
	FindAll(ctx context.Context) ([]*AuthMethod, error)
}

// AuthMethodUsecase 认证方法用例
type AuthMethodUsecase struct {
	repo   AuthMethodRepo
	logger *log.Helper
}

// NewAuthMethodUsecase 创建认证方法用例
func NewAuthMethodUsecase(repo AuthMethodRepo, logger log.Logger) *AuthMethodUsecase {
	return &AuthMethodUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// GetAuthMethodConfig 获取认证方法配置
// 对应原项目 getAuthMethodConfigLogic.go 第30-46行
func (uc *AuthMethodUsecase) GetAuthMethodConfig(ctx context.Context, method string) (*AuthMethod, error) {
	// 第31行：FindOneByMethod
	authMethod, err := uc.repo.FindByMethod(ctx, method)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("find one by method failed: method=%s, error=%v", method, err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 第37-38行：tool.DeepCopy(resp, method)
	// 第39-44行：解析 Config
	// 这些在 service 层完成
	return authMethod, nil
}

// UpdateAuthMethodConfig 更新认证方法配置
// 完全对应原项目 updateAuthMethodConfigLogic.go 第34-86行
func (uc *AuthMethodUsecase) UpdateAuthMethodConfig(ctx context.Context, req *UpdateAuthMethodRequest) (*AuthMethod, error) {
	// 第35-39行：FindOneByMethod
	method, err := uc.repo.FindByMethod(ctx, req.Method)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("find one by method failed: method=%s, error=%v", req.Method, err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 第41行：tool.DeepCopy(method, req) - 关键！将 req 的字段复制到 method
	method = tool.DeepCopy(method, req)

	// 第42-69行：处理 Config
	if req.Config != nil {
		// 第43行：检查是否是 map[string]interface{}
		_, exist := req.Config.(map[string]interface{})
		if !exist {
			// 第45行：如果不是 map，使用初始化配置
			req.Config = auth.InitializePlatformConfig(req.Method).(string)
		}

		// 第47-51行：特殊处理 email
		if req.Method == "email" {
			configs, _ := json.Marshal(req.Config)
			emailConfig := new(auth.EmailAuthConfig)
			emailConfig.Unmarshal(string(configs))
			req.Config = emailConfig
		}

		// 第54-58行：特殊处理 mobile
		if req.Method == "mobile" {
			configs, _ := json.Marshal(req.Config)
			mobileConfig := new(auth.MobileAuthConfig)
			mobileConfig.Unmarshal(string(configs))
			req.Config = mobileConfig
		}

		// 第61-65行：序列化 Config
		bytes, err := json.Marshal(req.Config)
		if err != nil {
			uc.logger.WithContext(ctx).Errorf("marshal config failed: %v", err)
			return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		method.Config = string(bytes)
	} else {
		// 第67-68行：初始化平台配置
		method.Config = auth.InitializePlatformConfig(req.Method).(string)
	}

	// 第70-73行：Update
	_, err = uc.repo.Update(ctx, method)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("update auth method failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	// 第75-82行：构造响应（在 service 层完成）
	// 第84行：defer UpdateGlobal
	defer uc.UpdateGlobal(ctx, method.Method)

	// 第85行：return
	return method, nil
}

// UpdateGlobal 更新全局配置
// 对应原项目第88-95行
func (uc *AuthMethodUsecase) UpdateGlobal(ctx context.Context, method string) {
	// 第89-91行
	if method == "email" {
		uc.logger.WithContext(ctx).Info("updating global email config")
		// initialize.Email(l.svcCtx)
	}
	// 第92-94行
	if method == "mobile" {
		uc.logger.WithContext(ctx).Info("updating global mobile config")
		// initialize.Mobile(l.svcCtx)
	}
}

// GetAuthMethodList 获取认证方法列表
// 对应原项目 getAuthMethodListLogic.go 第30-49行
func (uc *AuthMethodUsecase) GetAuthMethodList(ctx context.Context) ([]*AuthMethod, error) {
	// 第31-35行：FindAll
	methods, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("find all failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 第36-47行：DeepCopy 和 Unmarshal Config（在 service 层完成）
	return methods, nil
}

// GetEmailPlatforms 获取邮件平台列表
// 对应原项目 getEmailPlatformLogic.go 第28-32行
func (uc *AuthMethodUsecase) GetEmailPlatforms(ctx context.Context) []types.PlatformInfo {
	// 第30行：email.GetSupportedPlatforms()
	return email.GetSupportedPlatforms()
}

// GetSmsPlatforms 获取短信平台列表
// 对应原项目 getSmsPlatformLogic.go 第28-32行
func (uc *AuthMethodUsecase) GetSmsPlatforms(ctx context.Context) []types.PlatformInfo {
	// 第30行：sms.GetSupportedPlatforms()
	return sms.GetSupportedPlatforms()
}

// TestEmailSend 测试邮件发送
// 对应原项目 testEmailSendLogic.go 第30-41行
func (uc *AuthMethodUsecase) TestEmailSend(ctx context.Context, emailAddr string) (bool, string, error) {
	// 第31-34行：获取 email 配置
	authMethod, err := uc.repo.FindByMethod(ctx, "email")
	if err != nil || authMethod == nil {
		uc.logger.WithContext(ctx).Errorf("find email auth method failed: %v", err)
		return false, "", responsecode.NewKratosError(responsecode.ErrAuthMethodNotFound)
	}

	// 解析 email 配置
	var emailConfig auth.EmailAuthConfig
	if authMethod.Config != "" {
		if err := json.Unmarshal([]byte(authMethod.Config), &emailConfig); err != nil {
			uc.logger.WithContext(ctx).Errorf("unmarshal email config failed: %v", err)
			return false, "", responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
	}

	// 获取 platform_config 的 JSON 字符串
	platformConfigJSON, err := json.Marshal(emailConfig.PlatformConfig)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("marshal platform config failed: %v", err)
		return false, "", responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 第31-34行：创建 email sender
	client, err := email.NewSender(emailConfig.Platform, string(platformConfigJSON), "ProxyService")
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("new email sender err: %v", err)
		return false, fmt.Sprintf("send email err: %v", err), responsecode.NewKratosError(responsecode.ErrEmailSendFailed)
	}

	// 第36-40行：发送测试邮件
	err = client.Send([]string{emailAddr}, "Test Email Send", "this a test email send by kratos-service")
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("send email err: %v", err)
		return false, fmt.Sprintf("send email err: %v", err), responsecode.NewKratosError(responsecode.ErrEmailSendFailed)
	}

	uc.logger.WithContext(ctx).Infof("test email send to: %s success", emailAddr)
	return true, "邮件发送测试成功", nil
}

// TestSmsSend 测试短信发送
// 对应原项目 testSmsSendLogic.go 第30-42行
func (uc *AuthMethodUsecase) TestSmsSend(ctx context.Context, mobile string) (bool, string, error) {
	// 第31-34行：获取 mobile 配置
	authMethod, err := uc.repo.FindByMethod(ctx, "mobile")
	if err != nil || authMethod == nil {
		uc.logger.WithContext(ctx).Errorf("find mobile auth method failed: %v", err)
		return false, "", responsecode.NewKratosError(responsecode.ErrAuthMethodNotFound)
	}

	// 解析 mobile 配置
	var mobileConfig auth.MobileAuthConfig
	if authMethod.Config != "" {
		if err := json.Unmarshal([]byte(authMethod.Config), &mobileConfig); err != nil {
			uc.logger.WithContext(ctx).Errorf("unmarshal mobile config failed: %v", err)
			return false, "", responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
	}

	// 获取 platform_config 的 JSON 字符串
	platformConfigJSON, err := json.Marshal(mobileConfig.PlatformConfig)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("marshal platform config failed: %v", err)
		return false, "", responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 第31-34行：创建 sms sender
	client, err := sms.NewSender(mobileConfig.Platform, string(platformConfigJSON))
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("new sms sender err: %v", err)
		return false, fmt.Sprintf("send sms err: %v", err), responsecode.NewKratosError(responsecode.ErrSMSSendFailed)
	}

	// 第36-41行：发送测试短信（使用测试验证码 123456）
	err = client.SendCode("", mobile, "123456")
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("send sms err: %v", err)
		return false, fmt.Sprintf("send sms err: %v", err), responsecode.NewKratosError(responsecode.ErrSMSSendFailed)
	}

	uc.logger.WithContext(ctx).Infof("test sms send to: %s success", mobile)
	return true, "短信发送测试成功", nil
}
