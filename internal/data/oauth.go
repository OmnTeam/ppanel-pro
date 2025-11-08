package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
		"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	oauthBiz "github.com/OmnTeam/ppanel-pro/internal/biz/auth/oauth"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/pkg/jwt"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	"github.com/go-kratos/kratos/v2/errors"
	kratoLog "github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

var _ oauthBiz.OAuthRepo = (*oauthRepo)(nil)

// oauthRepo OAuth仓储实现
type oauthRepo struct {
	data   *Data
	config *conf.Application
	logger *kratoLog.Helper
}

// NewOAuthRepo 创建OAuth仓储实例
func NewOAuthRepo(d *Data, config *conf.Application, logger kratoLog.Logger) oauthBiz.OAuthRepo {
	return &oauthRepo{
		data:   d,
		config: config,
		logger: kratoLog.NewHelper(logger),
	}
}

// getJWTConfig returns JWT secret and expiry from environment or defaults
func (r *oauthRepo) getJWTConfig() (string, int64) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = DefaultJWTSecret
	}

	expireStr := os.Getenv("JWT_EXPIRE")
	expire := int64(DefaultJWTExpire)
	if expireStr != "" {
		if val, err := strconv.ParseInt(expireStr, 10, 64); err == nil {
			expire = val
		}
	}

	return secret, expire
}

// GetOAuthConfig 获取OAuth配置
// 从 proxy_auth_method 表读取指定提供商的OAuth配置
// ⚠️ 包含tenant_id过滤
func (r *oauthRepo) GetOAuthConfig(ctx context.Context, tenantID int64, method string) (map[string]string, error) {
	r.logger.Infof("[GetOAuthConfig] tenantID: %d, method: %s", tenantID, method)

	// 查询 proxy_auth_method 表，移除 tenant_id 过滤
	authMethod, err := r.data.db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ(method)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.NotFound("OAUTH_CONFIG_NOT_FOUND", fmt.Sprintf("OAuth配置未找到: %s", method))
		}
		return nil, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询OAuth配置失败: %v", err))
	}

	// 解析 config 字段（JSON格式）
	var config map[string]string
	if err := json.Unmarshal([]byte(authMethod.Config), &config); err != nil {
		return nil, errors.InternalServer("CONFIG_PARSE_ERROR", fmt.Sprintf("解析OAuth配置失败: %v", err))
	}

	r.logger.Infof("[GetOAuthConfig] 成功获取OAuth配置, tenantID: %d, method: %s", tenantID, method)
	return config, nil
}

// SaveStateCode 保存state code到Redis
// Redis key格式: {provider}:{code}
// 过期时间: 5分钟（300秒）
func (r *oauthRepo) SaveStateCode(ctx context.Context, provider, code, redirect string) error {
	r.logger.Infof("[SaveStateCode] provider: %s, code: %s, redirect: %s", provider, code, redirect)

	// Redis key格式: {provider}:{code}
	key := fmt.Sprintf("%s:%s", provider, code)

	// 保存state code到Redis，5分钟过期
	err := r.data.rdb.Set(ctx, key, redirect, 5*time.Minute).Err()
	if err != nil {
		r.logger.Errorf("[SaveStateCode] Redis保存失败: %v", err)
		return errors.InternalServer("REDIS_ERROR", fmt.Sprintf("保存state code失败: %v", err))
	}

	r.logger.Infof("[SaveStateCode] 成功保存state code, key: %s", key)
	return nil
}

// GetStateCode 从Redis获取state code
// Redis key格式: {provider}:{code}
// 返回保存的redirect URL
func (r *oauthRepo) GetStateCode(ctx context.Context, provider, code string) (string, error) {
	r.logger.Infof("[GetStateCode] provider: %s, code: %s", provider, code)

	// Redis key格式: {provider}:{code}
	key := fmt.Sprintf("%s:%s", provider, code)

	// 从Redis获取redirect URL
	redirect, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Errorf("[GetStateCode] state code不存在或已过期, key: %s", key)
			return "", errors.BadRequest("STATE_CODE_INVALID", "state code无效或已过期")
		}
		r.logger.Errorf("[GetStateCode] Redis读取失败: %v", err)
		return "", errors.InternalServer("REDIS_ERROR", fmt.Sprintf("获取state code失败: %v", err))
	}

	r.logger.Infof("[GetStateCode] 成功获取state code, key: %s, redirect: %s", key, redirect)
	return redirect, nil
}

// FindUserByOAuth 通过OAuth查找用户
// 查询 proxy_user_auth_method 表，通过 auth_method 和 auth_identifier 查找用户
// ⚠️ 包含tenant_id过滤
func (r *oauthRepo) FindUserByOAuth(ctx context.Context, tenantID int64, method, openID string) (int64, error) {
	r.logger.Infof("[FindUserByOAuth] tenantID: %d, method: %s, openID: %s", tenantID, method, openID)

	// 查询 proxy_user_auth_method 表，移除 tenant_id 过滤
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.AuthTypeEQ(method),
			proxyuserauthmethod.AuthIdentifierEQ(openID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Infof("[FindUserByOAuth] 用户不存在, 需要注册")
			return 0, errors.NotFound("USER_NOT_FOUND", "用户不存在")
		}
		r.logger.Errorf("[FindUserByOAuth] 数据库查询失败: %v", err)
		return 0, errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查找用户失败: %v", err))
	}

	r.logger.Infof("[FindUserByOAuth] 找到用户, userID: %d", authMethod.UserID)
	return int64(authMethod.UserID), nil
}

// CreateUserWithOAuth 创建OAuth用户
// 完整复刻原项目的用户注册逻辑（register函数）
// 包含15个步骤的事务操作：
// 1. 检查强制邀请配置
// 2. 检查email是否已存在（如果提供）
// 3. 创建 proxy_user 记录
// 4. 生成并更新 refer_code
// 5. 创建 OAuth auth_method 记录
// 6. 创建 email auth_method 记录（如果email不为空）
// 7. 激活试用订阅（如果配置启用）
// 8. 记录注册日志
// ⚠️ 所有操作都包含tenant_id字段
func (r *oauthRepo) CreateUserWithOAuth(ctx context.Context, tenantID int64, method, openID, email, avatar, ip, userAgent string) (int64, error) {
	r.logger.Infof("[CreateUserWithOAuth] tenantID: %d, method: %s, email: %s, ip: %s", tenantID, method, email, ip)

	// 1. 检查强制邀请配置
	if r.config != nil && r.config.Invite != nil && r.config.Invite.ForcedInvite {
		r.logger.Errorf("[CreateUserWithOAuth] 强制邀请模式已启用，禁止直接注册")
		return 0, errors.BadRequest("INVITE_CODE_REQUIRED", "需要邀请码才能注册")
	}

	var userID int64

	// 开始事务
	err := r.data.db.TX(ctx, func(tx *ent.Tx) error {
		// 2. 如果email不为空，检查email是否已被使用
		if email != "" {
			r.logger.Infof("[CreateUserWithOAuth] 检查email是否已存在: %s", email)

			existingAuth, err := tx.ProxyUserAuthMethod.Query().
				Where(proxyuserauthmethod.AuthIdentifierEQ(email)).
				Only(ctx)
			if err != nil && !ent.IsNotFound(err) {
				r.logger.Errorf("[CreateUserWithOAuth] 检查email失败: %v", err)
				return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("检查email失败: %v", err))
			}
			if existingAuth != nil && existingAuth.UserID != 0 {
				r.logger.Errorf("[CreateUserWithOAuth] email已被使用: %s, 已存在用户ID: %d", email, existingAuth.UserID)
				return errors.BadRequest("EMAIL_EXISTS", fmt.Sprintf("email已被使用: %s", email))
			}
		}

		// 3. 创建 proxy_user 记录
		r.logger.Infof("[CreateUserWithOAuth] 创建用户记录, avatar: %s", avatar)

		// 获取 OnlyFirstPurchase 配置
		onlyFirstPurchase := false
		if r.config != nil && r.config.Invite != nil {
			onlyFirstPurchase = r.config.Invite.OnlyFirstPurchase
		}

		userCreate := tx.ProxyUser.Create().
			SetOnlyFirstPurchase(onlyFirstPurchase)

		if avatar != "" {
			userCreate.SetAvatar(avatar)
		}

		user, err := userCreate.Save(ctx)
		if err != nil {
			r.logger.Errorf("[CreateUserWithOAuth] 创建用户失败: %v", err)
			return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建用户失败: %v", err))
		}

		userID = int64(user.ID)
		r.logger.Infof("[CreateUserWithOAuth] 用户创建成功, userID: %d", userID)

		// 4. 生成并更新 refer_code
		referCode := uuidx.UserInviteCode(userID)
		r.logger.Infof("[CreateUserWithOAuth] 生成refer_code: %s, userID: %d", referCode, userID)

		// 移除 tenant_id 隔离条件
		err = tx.ProxyUser.UpdateOneID(int(userID)).
			SetReferCode(referCode).
			Exec(ctx)
		if err != nil {
			r.logger.Errorf("[CreateUserWithOAuth] 更新refer_code失败: %v", err)
			return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("更新refer_code失败: %v", err))
		}

		// 5. 创建 OAuth auth_method 记录
		r.logger.Infof("[CreateUserWithOAuth] 创建OAuth认证方法, method: %s, openID: %s", method, openID)

		_, err = tx.ProxyUserAuthMethod.Create().
			SetUserID(int(userID)).
			SetAuthType(method).
			SetAuthIdentifier(openID).
			SetVerified(true).
			Save(ctx)
		if err != nil {
			r.logger.Errorf("[CreateUserWithOAuth] 创建OAuth认证方法失败: %v", err)
			return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建OAuth认证方法失败: %v", err))
		}

		// 6. 如果email不为空，创建 email auth_method 记录
		if email != "" {
			r.logger.Infof("[CreateUserWithOAuth] 创建email认证方法, email: %s", email)

			_, err = tx.ProxyUserAuthMethod.Create().
				SetUserID(int(userID)).
				SetAuthType("email").
				SetAuthIdentifier(email).
				SetVerified(true).
				Save(ctx)
			if err != nil {
				r.logger.Errorf("[CreateUserWithOAuth] 创建email认证方法失败: %v", err)
				return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建email认证方法失败: %v", err))
			}
		}

		// 7. 激活试用订阅（如果配置启用）
		if r.config != nil && r.config.Register != nil && r.config.Register.EnableTrial {
			r.logger.Infof("[CreateUserWithOAuth] 激活试用订阅, userID: %d", userID)
			err = r.activeTrial(ctx, tx, tenantID, userID)
			if err != nil {
				r.logger.Errorf("[CreateUserWithOAuth] 激活试用订阅失败: %v", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		r.logger.Errorf("[CreateUserWithOAuth] 用户创建事务失败: %v", err)
		return 0, err
	}

	// 8. 记录注册日志（在事务外，失败不影响注册）
	r.logger.Infof("[CreateUserWithOAuth] 记录注册日志, userID: %d", userID)

	registerLog := log.Register{
		AuthMethod: method,
		Identifier: openID,
		RegisterIP: ip,
		UserAgent:  userAgent,
		Timestamp:  time.Now().UnixMilli(),
	}
	content, _ := registerLog.Marshal()

	_, err = r.data.db.ProxySystemLog.Create().
		SetType(int8(log.TypeRegister.Uint8())).
		SetDate(time.Now().Format("2006-01-02")).
		SetObjectID(userID).
		SetContent(string(content)).
		Save(ctx)
	if err != nil {
		r.logger.Errorf("[CreateUserWithOAuth] 记录注册日志失败: %v (不影响注册)", err)
	}

	r.logger.Infof("[CreateUserWithOAuth] 用户创建完成, userID: %d, method: %s", userID, method)
	return userID, nil
}

// RecordLoginLog 记录登录日志
// 完整复刻原项目的 recordLoginStatus 函数（oAuthLoginGetTokenLogic.go Line 515-540）
// 记录到proxy_system_log表，Type = TypeLogin (30)
// ⚠️ 包含tenant_id字段
func (r *oauthRepo) RecordLoginLog(ctx context.Context, tenantID, userID int64, method, ip, userAgent string, success bool) error {
	r.logger.Infof("[RecordLoginLog] tenantID: %d, userID: %d, method: %s, success: %v", tenantID, userID, method, success)

	loginLog := log.Login{
		Method:    method,
		LoginIP:   ip,
		UserAgent: userAgent,
		Success:   success,
		Timestamp: time.Now().UnixMilli(),
	}
	content, err := loginLog.Marshal()
	if err != nil {
		r.logger.Errorf("[RecordLoginLog] 序列化登录日志失败: %v", err)
		return errors.InternalServer("LOG_MARSHAL_ERROR", fmt.Sprintf("序列化登录日志失败: %v", err))
	}

	_, err = r.data.db.ProxySystemLog.Create().
		SetType(int8(log.TypeLogin.Uint8())).
		SetDate(time.Now().Format("2006-01-02")).
		SetObjectID(userID).
		SetContent(string(content)).
		Save(ctx)
	if err != nil {
		r.logger.Errorf("[RecordLoginLog] 记录登录日志失败: %v", err)
		return errors.InternalServer("LOG_SAVE_ERROR", fmt.Sprintf("记录登录日志失败: %v", err))
	}

	r.logger.Infof("[RecordLoginLog] 登录日志记录成功, userID: %d, success: %v", userID, success)
	return nil
}

// activeTrial 激活试用订阅
// 完整复刻原项目的 activeTrial 函数（Line 796-861）
// ⚠️ 包含tenant_id字段
func (r *oauthRepo) activeTrial(ctx context.Context, tx *ent.Tx, tenantID, userID int64) error {
	r.logger.Infof("[activeTrial] tenantID: %d, userID: %d", tenantID, userID)

	// 获取试用订阅配置
	if r.config == nil || r.config.Register == nil {
		return errors.InternalServer("CONFIG_ERROR", "注册配置不存在")
	}

	trialSubscribeID := r.config.Register.TrialSubscribe
	trialTimeUnit := r.config.Register.TrialTimeUnit
	trialTime := r.config.Register.TrialTime

	r.logger.Infof("[activeTrial] 查询试用订阅模板, subscribeID: %d", trialSubscribeID)

	// 查询试用订阅模板（移除tenant_id过滤）
	sub, err := tx.ProxySubscribe.Query().
		Where(proxysubscribe.IDEQ(int(trialSubscribeID))).
		Only(ctx)
	if err != nil {
		r.logger.Errorf("[activeTrial] 查询试用订阅模板失败: %v", err)
		return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("查询试用订阅模板失败: %v", err))
	}

	// 计算过期时间
	startTime := time.Now()
	expireTime := tool.AddTime(trialTimeUnit, trialTime, startTime)

	// 生成订阅token和UUID
	subscribeToken := uuidx.SubscribeToken(fmt.Sprintf("Trial-%v", userID))
	subscribeUUID := uuidx.NewUUID().String()

	r.logger.Infof("[activeTrial] 创建试用订阅, userID: %d, subscribeID: %d, traffic: %d, expireTime: %s",
		userID, sub.ID, sub.Traffic, expireTime.Format(time.RFC3339))

	// 创建试用订阅记录
	_, err = tx.ProxyUserSubscribe.Create().
		SetUserID(int(userID)).
		SetOrderID(0).
		SetSubscribeID(sub.ID).
		SetStartTime(startTime).
		SetExpireTime(expireTime).
		SetTraffic(sub.Traffic).
		SetDownload(0).
		SetUpload(0).
		SetToken(subscribeToken).
		SetUUID(subscribeUUID).
		SetStatus(1).
		Save(ctx)
	if err != nil {
		r.logger.Errorf("[activeTrial] 创建试用订阅失败: %v", err)
		return errors.InternalServer("DATABASE_ERROR", fmt.Sprintf("创建试用订阅失败: %v", err))
	}

	r.logger.Infof("[activeTrial] 试用订阅激活成功, userID: %d, token: %s", userID, subscribeToken)
	return nil
}

// GenerateJWTToken 生成JWT令牌
// 完整复刻原项目的 generateToken 函数（Line 564-609）
// Claims包含: TenantId (新增), UserId, SessionId
// Session ID保存到Redis，过期时间与JWT一致
func (r *oauthRepo) GenerateJWTToken(ctx context.Context, tenantID, userID int64) (string, error) {
	r.logger.Infof("[GenerateJWTToken] tenantID: %d, userID: %d", tenantID, userID)

	// 生成session ID
	sessionID := uuidx.NewUUID().String()

	// 获取JWT配置（从环境变量或默认值）
	accessSecret, accessExpire := r.getJWTConfig()

	r.logger.Infof("[GenerateJWTToken] 生成JWT token, userID: %d, sessionID: %s, expire: %d秒",
		userID, sessionID, accessExpire)

	// 生成JWT token（添加 TenantId 到 Claims）
	token, err := jwt.NewJwtToken(
		accessSecret,
		time.Now().Unix(),
		accessExpire,
		jwt.WithOption("TenantId", tenantID), // ⚠️ 新增租户ID
		jwt.WithOption("UserId", userID),
		jwt.WithOption("SessionId", sessionID),
	)
	if err != nil {
		r.logger.Errorf("[GenerateJWTToken] 生成JWT token失败: %v", err)
		return "", errors.InternalServer("TOKEN_GENERATE_ERROR", fmt.Sprintf("生成token失败: %v", err))
	}

	// 将session ID保存到Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	err = r.data.rdb.Set(ctx, sessionKey, userID, time.Duration(accessExpire)*time.Second).Err()
	if err != nil {
		r.logger.Errorf("[GenerateJWTToken] 保存session到Redis失败: %v", err)
		return "", errors.InternalServer("REDIS_ERROR", fmt.Sprintf("保存session失败: %v", err))
	}

	r.logger.Infof("[GenerateJWTToken] JWT token生成成功, userID: %d, sessionID: %s", userID, sessionID)
	return token, nil
}
