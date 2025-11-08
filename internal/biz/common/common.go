package common

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

// CommonRepo defines repository interface for common operations
type CommonRepo interface {
	// GetAdsList gets ads list by status
	GetAdsList(ctx context.Context, status int) ([]*Ads, error)
	// GetClientList gets subscribe application list
	GetClientList(ctx context.Context) ([]*SubscribeClient, error)
	// GetTosConfig gets TOS/Privacy config from system table
	GetTosConfig(ctx context.Context, key string) (string, error)
	// GetSystemConfigByCategory gets system config by category
	GetSystemConfigByCategory(ctx context.Context, category string) (map[string]string, error)
	// GetWebAdConfig gets WebAD config
	GetWebAdConfig(ctx context.Context) (bool, error)
	// GetEnabledAuthMethods gets enabled auth methods
	GetEnabledAuthMethods(ctx context.Context) ([]string, error)
	// GetStatistics gets system statistics (user count, node count, etc.)
	GetStatistics(ctx context.Context) (*Statistics, error)
	// SendEmailVerificationCode sends email verification code
	SendEmailVerificationCode(ctx context.Context, email string, verifyType int32) (code string, err error)
	// SendSmsVerificationCode sends SMS verification code
	SendSmsVerificationCode(ctx context.Context, telephone, telephoneArea string, verifyType int32) (code string, err error)
	// CheckVerificationCode checks verification code
	CheckVerificationCode(ctx context.Context, method, account, code string, verifyType int32) (bool, error)
}

// Ads contains ads information
type Ads struct {
	ID          int64
	Title       string
	Type        string
	Content     string
	Description string
	TargetURL   string
	StartTime   int64
	EndTime     int64
	Status      int
	CreatedAt   int64
	UpdatedAt   int64
}

// DownloadLink contains platform-specific download links
type DownloadLink struct {
	IOS     string
	Android string
	Windows string
	Mac     string
	Linux   string
	Harmony string
}

// SubscribeClient contains subscribe application information
type SubscribeClient struct {
	ID           int64
	Name         string
	Description  string
	Icon         string
	Scheme       string
	IsDefault    bool
	DownloadLink DownloadLink
}

// Statistics contains system statistics information
type Statistics struct {
	User     int64
	Node     int64
	Country  int64
	Protocol []string
}

// CommonUsecase handles common business logic
type CommonUsecase struct {
	repo CommonRepo
	conf *conf.Application
	log  *log.Helper
}

// NewCommonUsecase creates a new common usecase
func NewCommonUsecase(repo CommonRepo, c *conf.Application, logger log.Logger) *CommonUsecase {
	return &CommonUsecase{
		repo: repo,
		conf: c,
		log:  log.NewHelper(log.With(logger, "module", "biz/common")),
	}
}

// GetAds gets ads list
func (uc *CommonUsecase) GetAds(ctx context.Context, device, position string) ([]*Ads, error) {
	// 广告获取：当前实现获取所有活跃广告，与原项目保持一致
	adsList, err := uc.repo.GetAdsList(ctx, 1)
	if err != nil {
		uc.log.Errorw("GetAdsList error", "error", err)
		return nil, err
	}

	return adsList, nil
}

// GetClient gets subscribe client list
func (uc *CommonUsecase) GetClient(ctx context.Context) ([]*SubscribeClient, int64, error) {
	clientList, err := uc.repo.GetClientList(ctx)
	if err != nil {
		uc.log.Errorw("GetClientList error", "error", err)
		return nil, 0, err
	}

	return clientList, int64(len(clientList)), nil
}

// GetPrivacyPolicy gets privacy policy content
func (uc *CommonUsecase) GetPrivacyPolicy(ctx context.Context) (string, error) {
	content, err := uc.repo.GetTosConfig(ctx, "privacy_policy")
	if err != nil {
		uc.log.Errorw("GetPrivacyPolicy error", "error", err)
		return "", err
	}

	return content, nil
}

// GetTos gets terms of service content
func (uc *CommonUsecase) GetTos(ctx context.Context) (string, error) {
	content, err := uc.repo.GetTosConfig(ctx, "tos_content")
	if err != nil {
		uc.log.Errorw("GetTos error", "error", err)
		return "", err
	}

	return content, nil
}

// GlobalConfig represents complete global configuration
type GlobalConfig struct {
	Site         *conf.Site
	Verify       *conf.Verify
	Auth         *AuthConfig
	Invite       *conf.Invite
	Currency     map[string]string
	Subscribe    *conf.Subscribe
	VerifyCode   map[string]string
	OAuthMethods []string
	WebAd        bool
}

// AuthConfig combines auth-related configurations
type AuthConfig struct {
	Mobile   *conf.MobileAuth
	Email    *conf.EmailAuth
	Register *conf.Register
}

// GetGlobalConfig gets global configuration
func (uc *CommonUsecase) GetGlobalConfig(ctx context.Context) (*GlobalConfig, error) {
	// Query currency config from database
	currency, err := uc.repo.GetSystemConfigByCategory(ctx, "currency")
	if err != nil {
		uc.log.Errorw("GetSystemConfigByCategory currency error", "error", err)
		// Use empty map if not found
		currency = make(map[string]string)
	}

	// Query verify code config from database
	verifyCode, err := uc.repo.GetSystemConfigByCategory(ctx, "verify_code")
	if err != nil {
		uc.log.Errorw("GetSystemConfigByCategory verify_code error", "error", err)
		// Use empty map if not found
		verifyCode = make(map[string]string)
	}

	// Get enabled auth methods from database
	oauthMethods, err := uc.repo.GetEnabledAuthMethods(ctx)
	if err != nil {
		uc.log.Errorw("GetEnabledAuthMethods error", "error", err)
		// Not critical, continue with empty list
		oauthMethods = []string{}
	}

	// Get WebAD config from database
	webAd, err := uc.repo.GetWebAdConfig(ctx)
	if err != nil {
		uc.log.Errorw("GetWebAdConfig error", "error", err)
		webAd = false
	}

	// Combine config file data with database data
	return &GlobalConfig{
		Site:   uc.conf.Site,
		Verify: uc.conf.Verify,
		Auth: &AuthConfig{
			Mobile:   uc.conf.Mobile,
			Email:    uc.conf.Email,
			Register: uc.conf.Register,
		},
		Invite:       uc.conf.Invite,
		Currency:     currency,
		Subscribe:    uc.conf.Subscribe,
		VerifyCode:   verifyCode,
		OAuthMethods: oauthMethods,
		WebAd:        webAd,
	}, nil
}

// GetStat gets system statistics
func (uc *CommonUsecase) GetStat(ctx context.Context) (*Statistics, error) {
	stat, err := uc.repo.GetStatistics(ctx)
	if err != nil {
		uc.log.Errorw("GetStatistics error", "error", err)
		return nil, err
	}

	return stat, nil
}

// SendEmailCode sends email verification code
func (uc *CommonUsecase) SendEmailCode(ctx context.Context, email string, verifyType int32) (string, error) {
	code, err := uc.repo.SendEmailVerificationCode(ctx, email, verifyType)
	if err != nil {
		uc.log.Errorw("SendEmailVerificationCode error", "error", err, "email", email)
		return "", err
	}

	return code, nil
}

// SendSmsCode sends SMS verification code
func (uc *CommonUsecase) SendSmsCode(ctx context.Context, telephone, telephoneArea string, verifyType int32) (string, error) {
	code, err := uc.repo.SendSmsVerificationCode(ctx, telephone, telephoneArea, verifyType)
	if err != nil {
		uc.log.Errorw("SendSmsVerificationCode error", "error", err, "telephone", telephone)
		return "", err
	}

	return code, nil
}

// CheckVerificationCode checks verification code
func (uc *CommonUsecase) CheckVerificationCode(ctx context.Context, method, account, code string, verifyType int32) (bool, error) {
	valid, err := uc.repo.CheckVerificationCode(ctx, method, account, code, verifyType)
	if err != nil {
		uc.log.Errorw("CheckVerificationCode error", "error", err, "method", method, "account", account)
		return false, err
	}

	return valid, nil
}
