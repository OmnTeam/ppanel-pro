package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/pkg/oauth/apple"
	"github.com/OmnTeam/ppanel-pro/pkg/oauth/google"
	"github.com/OmnTeam/ppanel-pro/pkg/oauth/telegram"
	"github.com/OmnTeam/ppanel-pro/pkg/random"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/oauth2"
)

const (
	// OAuth提供商常量
	OAuthGoogle   = "google"
	OAuthApple    = "apple"
	OAuthTelegram = "telegram"
	OAuthGitHub   = "github"
	OAuthFacebook = "facebook"

	// Telegram特殊常量
	TelegramDomain = "telegram.org"
	AuthExpire     = 86400 // 24小时（秒）
)

// oauthRequest OAuth回调请求结构
type oauthRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// oauthUseCase OAuth业务用例实现
type oauthUseCase struct {
	repo   OAuthRepo
	logger *log.Helper
}

// NewOAuthUseCase 创建OAuth业务用例实例
func NewOAuthUseCase(repo OAuthRepo, logger log.Logger) OAuthUseCase {
	return &oauthUseCase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// OAuthLogin 获取OAuth登录URL
// 完整复刻原项目 OAuthLoginLogic（oAuthLoginLogic.go）
// 包含：获取配置、生成state code、保存到Redis、构建OAuth URL
// ⚠️ 所有操作包含tenant_id
func (uc *oauthUseCase) OAuthLogin(ctx context.Context, params *OAuthParams) (*OAuthResult, error) {
	uc.logger.Infof("[OAuthLogin] tenantID: %d, method: %s", params.TenantID, params.Method)

	// 根据不同的OAuth提供商生成登录URL
	var uri string
	var err error

	switch params.Method {
	case OAuthGoogle:
		uri, err = uc.googleLogin(ctx, params)
	case OAuthApple:
		uri, err = uc.appleLogin(ctx, params)
	case OAuthTelegram:
		uri, err = uc.telegramLogin(ctx, params)
	case OAuthGitHub:
		uri, err = uc.githubLogin(ctx, params)
	case OAuthFacebook:
		uri, err = uc.facebookLogin(ctx, params)
	default:
		return nil, errors.BadRequest("UNSUPPORTED_OAUTH_METHOD", fmt.Sprintf("不支持的OAuth方法: %s", params.Method))
	}

	if err != nil {
		uc.logger.Errorf("[OAuthLogin] OAuth登录失败: %v", err)
		return nil, err
	}

	uc.logger.Infof("[OAuthLogin] OAuth登录URL生成成功, method: %s, redirect: %s", params.Method, uri)
	return &OAuthResult{
		Redirect: uri,
	}, nil
}

// googleLogin Google OAuth登录
// 完整复刻原项目 google() 函数（oAuthLoginLogic.go Line 60-84）
// ⚠️ 添加tenant_id过滤
func (uc *oauthUseCase) googleLogin(ctx context.Context, params *OAuthParams) (string, error) {
	uc.logger.Infof("[googleLogin] tenantID: %d, redirect: %s", params.TenantID, params.Redirect)

	// 1. 从数据库获取Google OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthGoogle)
	if err != nil {
		uc.logger.Errorf("[googleLogin] 获取Google配置失败: %v", err)
		return "", err
	}

	// 2. 创建Google OAuth客户端
	client := google.New(&google.Config{
		ClientID:     config["client_id"],
		ClientSecret: config["client_secret"],
		RedirectURL:  params.Redirect,
	})

	// 3. 生成8位随机state code
	stateCode := random.KeyNew(8, 1)
	uc.logger.Infof("[googleLogin] 生成state code: %s", stateCode)

	// 4. 保存state code到Redis（5分钟过期）
	err = uc.repo.SaveStateCode(ctx, OAuthGoogle, stateCode, params.Redirect)
	if err != nil {
		uc.logger.Errorf("[googleLogin] 保存state code失败: %v", err)
		return "", err
	}

	// 5. 构建Google OAuth授权URL
	authURL := client.AuthCodeURL(stateCode, oauth2.AccessTypeOffline)

	uc.logger.Infof("[googleLogin] Google OAuth URL生成成功: %s", authURL)
	return authURL, nil
}

// appleLogin Apple OAuth登录
// 完整复刻原项目 apple() 函数（oAuthLoginLogic.go Line 90-110）
// ⚠️ 添加tenant_id过滤
func (uc *oauthUseCase) appleLogin(ctx context.Context, params *OAuthParams) (string, error) {
	uc.logger.Infof("[appleLogin] tenantID: %d, redirect: %s", params.TenantID, params.Redirect)

	// 1. 从数据库获取Apple OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthApple)
	if err != nil {
		uc.logger.Errorf("[appleLogin] 获取Apple配置失败: %v", err)
		return "", err
	}

	// 2. 生成8位随机state code
	stateCode := random.KeyNew(8, 1)
	uc.logger.Infof("[appleLogin] 生成state code: %s", stateCode)

	// 3. 保存state code到Redis（5分钟过期）
	err = uc.repo.SaveStateCode(ctx, OAuthApple, stateCode, params.Redirect)
	if err != nil {
		uc.logger.Errorf("[appleLogin] 保存state code失败: %v", err)
		return "", err
	}

	// 4. 构建Apple OAuth授权URL
	// Apple回调URL格式: {redirect_url}/v1/auth/oauth/callback/apple
	callbackURL := fmt.Sprintf("%s/v1/auth/oauth/callback/apple", config["redirect_url"])
	authURL := fmt.Sprintf(
		"https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=name email&response_mode=form_post",
		config["client_id"],
		callbackURL,
		stateCode,
	)

	uc.logger.Infof("[appleLogin] Apple OAuth URL生成成功: %s", authURL)
	return authURL, nil
}

// telegramLogin Telegram OAuth登录
// 完整复刻原项目 telegram() 函数（oAuthLoginLogic.go Line 114-134）
// ⚠️ 添加tenant_id过滤
func (uc *oauthUseCase) telegramLogin(ctx context.Context, params *OAuthParams) (string, error) {
	uc.logger.Infof("[telegramLogin] tenantID: %d, redirect: %s", params.TenantID, params.Redirect)

	// 1. 从数据库获取Telegram OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthTelegram)
	if err != nil {
		uc.logger.Errorf("[telegramLogin] 获取Telegram配置失败: %v", err)
		return "", err
	}

	// 2. 生成8位随机state code
	stateCode := random.KeyNew(8, 1)
	uc.logger.Infof("[telegramLogin] 生成state code: %s", stateCode)

	// 3. 保存state code到Redis（5分钟过期）
	// 注意：原项目这里有个bug，使用了"apple"而不是"telegram"，但我们修正为正确的
	err = uc.repo.SaveStateCode(ctx, OAuthTelegram, stateCode, params.Redirect)
	if err != nil {
		uc.logger.Errorf("[telegramLogin] 保存state code失败: %v", err)
		return "", err
	}

	// 4. 生成Telegram OAuth URL
	authURL := telegram.GenerateTelegramOAuthURL(config["bot_token"], stateCode, params.Redirect)

	uc.logger.Infof("[telegramLogin] Telegram OAuth URL生成成功: %s", authURL)
	return authURL, nil
}

// githubLogin GitHub OAuth登录（未实现）
func (uc *oauthUseCase) githubLogin(ctx context.Context, params *OAuthParams) (string, error) {
	return "", errors.BadRequest("GITHUB_OAUTH_NOT_IMPLEMENTED", "GitHub OAuth未实现")
}

// facebookLogin Facebook OAuth登录（未实现）
func (uc *oauthUseCase) facebookLogin(ctx context.Context, params *OAuthParams) (string, error) {
	return "", errors.BadRequest("FACEBOOK_OAUTH_NOT_IMPLEMENTED", "Facebook OAuth未实现")
}

// OAuthLoginGetToken 处理OAuth回调并获取token
// 完整复刻原项目 OAuthLoginGetTokenLogic（oAuthLoginGetTokenLogic.go）
// 包含：验证state、换取token、获取用户信息、查找或创建用户、生成JWT、记录登录日志
// ⚠️ 所有操作包含tenant_id
func (uc *oauthUseCase) OAuthLoginGetToken(ctx context.Context, params *OAuthTokenParams) (*OAuthTokenResult, error) {
	uc.logger.Infof("[OAuthLoginGetToken] tenantID: %d, method: %s", params.TenantID, params.Method)

	// 初始化登录状态（默认为失败）
	var userID int64
	loginSuccess := false

	// defer记录登录日志（不管成功还是失败）
	// 完整复刻原项目 Line 67-69
	defer func() {
		// 如果userID为0（未获取到用户ID），使用0作为objectID
		logUserID := userID
		if logUserID == 0 {
			logUserID = 0
		}
		// 记录登录日志（失败不影响主流程）
		if err := uc.repo.RecordLoginLog(context.Background(), params.TenantID, logUserID, params.Method, params.IP, params.UserAgent, loginSuccess); err != nil {
			uc.logger.Errorf("[OAuthLoginGetToken] 记录登录日志失败: %v (不影响登录)", err)
		}
	}()

	// 根据不同的OAuth提供商处理token获取
	var err error

	switch params.Method {
	case OAuthGoogle:
		userID, err = uc.googleGetToken(ctx, params)
	case OAuthApple:
		userID, err = uc.appleGetToken(ctx, params)
	case OAuthTelegram:
		userID, err = uc.telegramGetToken(ctx, params)
	default:
		return nil, errors.BadRequest("UNSUPPORTED_OAUTH_METHOD", fmt.Sprintf("不支持的OAuth方法: %s", params.Method))
	}

	if err != nil {
		uc.logger.Errorf("[OAuthLoginGetToken] OAuth token获取失败: %v", err)
		return nil, err
	}

	// 生成JWT token（包含TenantId, UserId, SessionId）
	token, err := uc.repo.GenerateJWTToken(ctx, params.TenantID, userID)
	if err != nil {
		uc.logger.Errorf("[OAuthLoginGetToken] 生成JWT token失败: %v", err)
		return nil, err
	}

	// 登录成功，设置状态为true（defer中会记录）
	loginSuccess = true

	uc.logger.Infof("[OAuthLoginGetToken] OAuth登录成功, userID: %d, method: %s", userID, params.Method)
	return &OAuthTokenResult{
		Token: token,
	}, nil
}

// googleGetToken Google OAuth token获取
// 完整复刻原项目 google() 函数（oAuthLoginGetTokenLogic.go Line 85-161）
// ⚠️ 添加tenant_id到所有数据库操作
func (uc *oauthUseCase) googleGetToken(ctx context.Context, params *OAuthTokenParams) (int64, error) {
	uc.logger.Infof("[googleGetToken] tenantID: %d", params.TenantID)

	// 1. 解析callback数据
	var request oauthRequest
	if err := json.Unmarshal([]byte(params.Callback), &request); err != nil {
		uc.logger.Errorf("[googleGetToken] 解析callback数据失败: %v", err)
		return 0, errors.BadRequest("PARSE_CALLBACK_ERROR", fmt.Sprintf("解析callback数据失败: %v", err))
	}

	uc.logger.Infof("[googleGetToken] 验证state code: %s", request.State)

	// 2. 验证state code（从Redis获取redirect URL）
	redirect, err := uc.repo.GetStateCode(ctx, OAuthGoogle, request.State)
	if err != nil {
		uc.logger.Errorf("[googleGetToken] 验证state code失败: %v", err)
		return 0, err
	}

	// 3. 获取Google OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthGoogle)
	if err != nil {
		uc.logger.Errorf("[googleGetToken] 获取Google配置失败: %v", err)
		return 0, err
	}

	// 4. 创建Google OAuth客户端
	client := google.New(&google.Config{
		ClientID:     config["client_id"],
		ClientSecret: config["client_secret"],
		RedirectURL:  redirect,
	})

	uc.logger.Infof("[googleGetToken] 使用authorization code换取token")

	// 5. 使用authorization code换取access token
	token, err := client.Exchange(ctx, request.Code)
	if err != nil {
		uc.logger.Errorf("[googleGetToken] 换取token失败: %v", err)
		return 0, errors.InternalServer("EXCHANGE_TOKEN_ERROR", fmt.Sprintf("换取token失败: %v", err))
	}

	uc.logger.Infof("[googleGetToken] 获取Google用户信息")

	// 6. 获取Google用户信息
	googleUserInfo, err := client.GetUserInfo(token.AccessToken)
	if err != nil {
		uc.logger.Errorf("[googleGetToken] 获取用户信息失败: %v", err)
		return 0, errors.InternalServer("GET_USER_INFO_ERROR", fmt.Sprintf("获取用户信息失败: %v", err))
	}

	uc.logger.Infof("[googleGetToken] Google用户信息: openID=%s, email=%s", googleUserInfo.OpenID, googleUserInfo.Email)

	// 7. 查找或注册用户（包含tenant_id、ip、userAgent）
	userID, err := uc.findOrRegisterUser(ctx, params.TenantID, OAuthGoogle, googleUserInfo.OpenID, googleUserInfo.Email, googleUserInfo.Picture, params.IP, params.UserAgent)
	if err != nil {
		uc.logger.Errorf("[googleGetToken] 查找或注册用户失败: %v", err)
		return 0, err
	}

	uc.logger.Infof("[googleGetToken] Google OAuth登录成功, userID: %d", userID)
	return userID, nil
}

// appleGetToken Apple OAuth token获取
// 完整复刻原项目 apple() 函数（oAuthLoginGetTokenLogic.go Line 163-261）
// ⚠️ 添加tenant_id到所有数据库操作
func (uc *oauthUseCase) appleGetToken(ctx context.Context, params *OAuthTokenParams) (int64, error) {
	uc.logger.Infof("[appleGetToken] tenantID: %d", params.TenantID)

	// 1. 解析callback数据
	var callback map[string]interface{}
	if err := json.Unmarshal([]byte(params.Callback), &callback); err != nil {
		uc.logger.Errorf("[appleGetToken] 解析callback数据失败: %v", err)
		return 0, errors.BadRequest("PARSE_CALLBACK_ERROR", fmt.Sprintf("解析callback数据失败: %v", err))
	}

	state, _ := callback["state"].(string)
	code, _ := callback["code"].(string)

	uc.logger.Infof("[appleGetToken] 验证state code: %s", state)

	// 2. 验证state code（从Redis获取redirect URL）
	_, err := uc.repo.GetStateCode(ctx, OAuthApple, state)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 验证state code失败: %v", err)
		return 0, err
	}

	// 3. 获取Apple OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthApple)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 获取Apple配置失败: %v", err)
		return 0, err
	}

	// 4. 创建Apple OAuth客户端
	client, err := apple.New(apple.Config{
		ClientID:     config["client_id"],
		TeamID:       config["team_id"],
		KeyID:        config["key_id"],
		ClientSecret: config["client_secret"],
		RedirectURI:  config["redirect_url"],
	})
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 创建Apple客户端失败: %v", err)
		return 0, errors.InternalServer("CREATE_APPLE_CLIENT_ERROR", fmt.Sprintf("创建Apple客户端失败: %v", err))
	}

	uc.logger.Infof("[appleGetToken] 验证Apple web token")

	// 5. 验证Apple web token
	resp, err := client.VerifyWebToken(ctx, code)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 验证web token失败: %v", err)
		return 0, errors.InternalServer("VERIFY_WEB_TOKEN_ERROR", fmt.Sprintf("验证web token失败: %v", err))
	}

	if resp.Error != "" {
		uc.logger.Errorf("[appleGetToken] Apple返回错误: %s", resp.Error)
		return 0, errors.InternalServer("APPLE_ERROR", fmt.Sprintf("Apple返回错误: %s", resp.Error))
	}

	// 6. 获取Apple unique ID
	appleUnique, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 获取Apple unique ID失败: %v", err)
		return 0, errors.InternalServer("GET_UNIQUE_ID_ERROR", fmt.Sprintf("获取Apple unique ID失败: %v", err))
	}

	// 7. 获取Apple用户claims
	appleUserInfo, err := apple.GetClaims(resp.AccessToken)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 获取Apple用户信息失败: %v", err)
		return 0, errors.InternalServer("GET_APPLE_USER_INFO_ERROR", fmt.Sprintf("获取Apple用户信息失败: %v", err))
	}

	// 8. 提取email（可能为空）
	email := ""
	if emailVal, ok := (*appleUserInfo)["email"]; ok {
		email, _ = emailVal.(string)
	}

	uc.logger.Infof("[appleGetToken] Apple用户信息: uniqueID=%s, email=%s", appleUnique, email)

	// 9. 查找或注册用户（包含tenant_id）
	userID, err := uc.findOrRegisterUser(ctx, params.TenantID, OAuthApple, appleUnique, email, "", params.IP, params.UserAgent)
	if err != nil {
		uc.logger.Errorf("[appleGetToken] 查找或注册用户失败: %v", err)
		return 0, err
	}

	uc.logger.Infof("[appleGetToken] Apple OAuth登录成功, userID: %d", userID)
	return userID, nil
}

// telegramGetToken Telegram OAuth token获取
// 完整复刻原项目 telegram() 函数（oAuthLoginGetTokenLogic.go Line 263-324）
// ⚠️ 添加tenant_id到所有数据库操作
func (uc *oauthUseCase) telegramGetToken(ctx context.Context, params *OAuthTokenParams) (int64, error) {
	uc.logger.Infof("[telegramGetToken] tenantID: %d", params.TenantID)

	// 1. 获取Telegram OAuth配置（包含tenant_id过滤）
	config, err := uc.repo.GetOAuthConfig(ctx, params.TenantID, OAuthTelegram)
	if err != nil {
		uc.logger.Errorf("[telegramGetToken] 获取Telegram配置失败: %v", err)
		return 0, err
	}

	// 2. 解析callback数据
	var callback map[string]interface{}
	if err := json.Unmarshal([]byte(params.Callback), &callback); err != nil {
		uc.logger.Errorf("[telegramGetToken] 解析callback数据失败: %v", err)
		return 0, errors.BadRequest("PARSE_CALLBACK_ERROR", fmt.Sprintf("解析callback数据失败: %v", err))
	}

	encodeText, _ := callback["tgAuthResult"].(string)
	uc.logger.Infof("[telegramGetToken] 解析Telegram callback数据, 长度: %d", len(encodeText))

	// 3. 解析并验证Telegram callback数据
	callbackData, err := telegram.ParseAndValidateBase64([]byte(encodeText), config["bot_token"])
	if err != nil {
		uc.logger.Errorf("[telegramGetToken] 解析Telegram callback失败: %v", err)
		return 0, errors.BadRequest("PARSE_TELEGRAM_CALLBACK_ERROR", fmt.Sprintf("解析Telegram callback失败: %v", err))
	}

	// 4. 验证auth date（24小时内有效）
	if time.Now().Unix()-*callbackData.AuthDate > AuthExpire {
		uc.logger.Errorf("[telegramGetToken] Telegram auth date已过期, authDate: %d, now: %d",
			*callbackData.AuthDate, time.Now().Unix())
		return 0, errors.BadRequest("AUTH_DATE_EXPIRED", "Telegram认证时间已过期")
	}

	// 5. 构造用户ID和email
	userIDStr := fmt.Sprintf("%v", *callbackData.Id)
	email := fmt.Sprintf("%v@%s", *callbackData.Id, TelegramDomain)
	avatar := ""
	if callbackData.PhotoUrl != nil {
		avatar = *callbackData.PhotoUrl
	}

	uc.logger.Infof("[telegramGetToken] Telegram用户信息: userID=%s, email=%s", userIDStr, email)

	// 6. 查找或注册用户（包含tenant_id）
	userID, err := uc.findOrRegisterUser(ctx, params.TenantID, OAuthTelegram, userIDStr, email, avatar, params.IP, params.UserAgent)
	if err != nil {
		uc.logger.Errorf("[telegramGetToken] 查找或注册用户失败: %v", err)
		return 0, err
	}

	uc.logger.Infof("[telegramGetToken] Telegram OAuth登录成功, userID: %d", userID)
	return userID, nil
}

// findOrRegisterUser 查找或注册用户
// 完整复刻原项目 findOrRegisterUser() 函数（oAuthLoginGetTokenLogic.go Line 743-794）
// ⚠️ 包含tenant_id过滤
func (uc *oauthUseCase) findOrRegisterUser(ctx context.Context, tenantID int64, authType, openID, email, avatar, ip, userAgent string) (int64, error) {
	uc.logger.Infof("[findOrRegisterUser] tenantID: %d, authType: %s, openID: %s, email: %s",
		tenantID, authType, openID, email)

	// 1. 通过OAuth查找用户（包含tenant_id过滤）
	userID, err := uc.repo.FindUserByOAuth(ctx, tenantID, authType, openID)
	if err != nil {
		// 用户不存在，需要注册
		if errors.IsNotFound(err) {
			uc.logger.Infof("[findOrRegisterUser] 用户不存在，开始注册")

			// 2. 创建新用户（包含tenant_id、ip、userAgent）
			userID, err = uc.repo.CreateUserWithOAuth(ctx, tenantID, authType, openID, email, avatar, ip, userAgent)
			if err != nil {
				uc.logger.Errorf("[findOrRegisterUser] 创建用户失败: %v", err)
				return 0, err
			}

			uc.logger.Infof("[findOrRegisterUser] 用户注册成功, userID: %d", userID)
			return userID, nil
		}

		// 其他数据库错误
		uc.logger.Errorf("[findOrRegisterUser] 查找用户失败: %v", err)
		return 0, err
	}

	// 用户已存在
	uc.logger.Infof("[findOrRegisterUser] 找到已存在用户, userID: %d", userID)
	return userID, nil
}

// AppleLoginCallback 处理Apple登录回调
// 完整复刻原项目 appleLoginCallbackLogic.go
// 注意：gRPC服务需要特殊处理HTTP重定向
func (uc *oauthUseCase) AppleLoginCallback(ctx context.Context, params *AppleCallbackParams) error {
	uc.logger.Infof("[AppleLoginCallback] state: %s, code: %s", params.State, params.Code)

	// 1. 验证state code（从Redis获取redirect URL）
	redirect, err := uc.repo.GetStateCode(ctx, OAuthApple, params.State)
	if err != nil {
		uc.logger.Errorf("[AppleLoginCallback] 验证state code失败: %v", err)
		return err
	}

	// 2. 构建重定向URL
	redirectURL := fmt.Sprintf("%s?method=apple&code=%s&state=%s", redirect, params.Code, params.State)

	uc.logger.Infof("[AppleLoginCallback] 重定向URL: %s", redirectURL)

	// gRPC服务返回重定向URL，HTTP重定向在网关层处理

	return errors.BadRequest("APPLE_CALLBACK_REDIRECT", "Apple回调重定向需要在HTTP层实现")
}
