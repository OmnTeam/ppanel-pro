package oauth

import (
	"context"
)

// OAuthParams OAuth登录参数
type OAuthParams struct {
	Method   string // OAuth提供商: google, apple, telegram, github, facebook
	Redirect string // 登录成功后的重定向URL
}

// OAuthResult OAuth登录结果
type OAuthResult struct {
	Redirect string // OAuth提供商的授权URL
}

// OAuthTokenParams OAuth获取令牌参数
type OAuthTokenParams struct {
	Method    string // OAuth提供商
	Callback  string // OAuth回调数据（JSON格式）
	IP        string // 客户端IP
	UserAgent string // 用户代理
}

// OAuthTokenResult OAuth令牌结果
type OAuthTokenResult struct {
	Token string // JWT访问令牌
}

// AppleCallbackParams Apple回调参数
type AppleCallbackParams struct {
	Code    string // Authorization code
	IDToken string // ID token
	State   string // State code
}

// OAuthUseCase OAuth业务用例接口
type OAuthUseCase interface {
	// OAuthLogin 获取OAuth登录URL
	OAuthLogin(ctx context.Context, params *OAuthParams) (*OAuthResult, error)

	// OAuthLoginGetToken 处理OAuth回调并获取token
	OAuthLoginGetToken(ctx context.Context, params *OAuthTokenParams) (*OAuthTokenResult, error)

	// AppleLoginCallback 处理Apple登录回调
	AppleLoginCallback(ctx context.Context, params *AppleCallbackParams) error
}

// OAuthRepo OAuth数据仓储接口
type OAuthRepo interface {
	// GetOAuthConfig 获取OAuth配置
	GetOAuthConfig(ctx context.Context, method string) (map[string]string, error)

	// SaveStateCode 保存state code到缓存
	SaveStateCode(ctx context.Context, provider, code, redirect string) error

	// GetStateCode 从缓存获取state code
	GetStateCode(ctx context.Context, provider, code string) (string, error)

	// FindUserByOAuth 通过OAuth查找用户
	FindUserByOAuth(ctx context.Context, method, openID string) (int, error)

	// CreateUserWithOAuth 创建OAuth用户
	CreateUserWithOAuth(ctx context.Context, method, openID, email, avatar, ip, userAgent string) (int, error)

	// GenerateJWTToken 生成JWT令牌
	GenerateJWTToken(ctx context.Context, userID int) (string, error)

	// RecordLoginLog 记录登录日志
	RecordLoginLog(ctx context.Context, userID int, method, ip, userAgent string, success bool) error
}
