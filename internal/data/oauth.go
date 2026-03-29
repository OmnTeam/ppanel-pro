package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	oauthBiz "github.com/OmnTeam/ppanel-pro/internal/biz/auth/oauth"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
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

// GetOAuthConfig 获取OAuth配置
// 从 proxy_auth_method 表读取指定提供商的OAuth配置
func (r *oauthRepo) GetOAuthConfig(ctx context.Context, method string) (map[string]string, error) {
	r.logger.Infof("[GetOAuthConfig] method: %s", method)

	// 查询 proxy_auth_method 表
	authMethod, err := r.data.db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ(method)).
		Only(ctx)
	if err != nil {
		return nil, responsecode.NewDatabaseQueryError()
	}

	// 解析 config 字段（JSON格式）
	var config map[string]string
	if err := json.Unmarshal([]byte(authMethod.Config), &config); err != nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	r.logger.Infof("[GetOAuthConfig] 成功获取OAuth配置, method: %s", method)
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
		return responsecode.NewKratosError(responsecode.ErrInternalError)
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
			return "", responsecode.NewKratosError(responsecode.ErrInternalError)
		}
		r.logger.Errorf("[GetStateCode] Redis读取失败: %v", err)
		return "", responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	r.logger.Infof("[GetStateCode] 成功获取state code, key: %s, redirect: %s", key, redirect)
	return redirect, nil
}

// FindUserByOAuth 通过OAuth查找用户
// 查询 proxy_user_auth_method 表，通过 auth_method 和 auth_identifier 查找用户
func (r *oauthRepo) FindUserByOAuth(ctx context.Context, method, openID string) (int, error) {
	r.logger.Infof("[FindUserByOAuth] method: %s, openID: %s", method, openID)

	// 查询 proxy_user_auth_method 表
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
		return 0, responsecode.NewDatabaseQueryError()
	}

	r.logger.Infof("[FindUserByOAuth] 找到用户, userID: %d", authMethod.UserID)
	return int(authMethod.UserID), nil
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
func (r *oauthRepo) CreateUserWithOAuth(ctx context.Context, method, openID, email, avatar, ip, userAgent string) (int, error) {
	r.logger.Infof("[CreateUserWithOAuth] method: %s, email: %s, ip: %s", method, email, ip)

	// 1. 检查强制邀请配置
	if r.config != nil && r.config.Invite != nil && r.config.Invite.ForcedInvite {
		r.logger.Errorf("[CreateUserWithOAuth] 强制邀请模式已启用，禁止直接注册")
		return 0, responsecode.NewKratosError(responsecode.ErrInviteCodeError)
	}

	var userID int
	var trialSubscribe *ent.ProxyUserSubscribe

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
				return responsecode.NewDatabaseQueryError()
			}
			if existingAuth != nil && existingAuth.UserID != 0 {
				r.logger.Errorf("[CreateUserWithOAuth] email已被使用: %s, 已存在用户ID: %d", email, existingAuth.UserID)
				return responsecode.NewKratosError(responsecode.ErrDuplicateEmail)
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
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}

		userID = int(user.ID)
		r.logger.Infof("[CreateUserWithOAuth] 用户创建成功, userID: %d", userID)

		// 4. 生成并更新 refer_code
		referCode := uuidx.UserInviteCode(int64(userID))
		r.logger.Infof("[CreateUserWithOAuth] 生成refer_code: %s, userID: %d", referCode, userID)

		// 单库模型下直接查询当前记录
		err = tx.ProxyUser.UpdateOneID(int64(userID)).
			SetReferCode(referCode).
			Exec(ctx)
		if err != nil {
			r.logger.Errorf("[CreateUserWithOAuth] 更新refer_code失败: %v", err)
			return responsecode.NewDatabaseUpdateError()
		}

		// 5. 创建 OAuth auth_method 记录
		r.logger.Infof("[CreateUserWithOAuth] 创建OAuth认证方法, method: %s, openID: %s", method, openID)

		_, err = tx.ProxyUserAuthMethod.Create().
			SetUserID(int64(userID)).
			SetAuthType(method).
			SetAuthIdentifier(openID).
			SetVerified(true).
			Save(ctx)
		if err != nil {
			r.logger.Errorf("[CreateUserWithOAuth] 创建OAuth认证方法失败: %v", err)
			return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
		}

		// 6. 如果email不为空，创建 email auth_method 记录
		if email != "" {
			r.logger.Infof("[CreateUserWithOAuth] 创建email认证方法, email: %s", email)

			_, err = tx.ProxyUserAuthMethod.Create().
				SetUserID(int64(userID)).
				SetAuthType("email").
				SetAuthIdentifier(email).
				SetVerified(true).
				Save(ctx)
			if err != nil {
				r.logger.Errorf("[CreateUserWithOAuth] 创建email认证方法失败: %v", err)
				return responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
			}
		}

		// 7. 激活试用订阅（如果配置启用）
		if r.config != nil && r.config.Register != nil && r.config.Register.EnableTrial {
			r.logger.Infof("[CreateUserWithOAuth] 激活试用订阅, userID: %d", userID)
			trialSubscribe, err = r.activeTrial(ctx, tx, userID)
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
	if trialSubscribe != nil {
		legacyAuthRepo := &authRepo{
			data:   r.data,
			config: r.config,
			log:    r.logger,
		}
		legacyAuthRepo.clearTrialCaches(ctx, trialSubscribe)
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
		SetObjectID(int64(userID)).
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
func (r *oauthRepo) RecordLoginLog(ctx context.Context, userID int, method, ip, userAgent string, success bool) error {
	r.logger.Infof("[RecordLoginLog] userID: %d, method: %s, success: %v", userID, method, success)

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
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	_, err = r.data.db.ProxySystemLog.Create().
		SetType(int8(log.TypeLogin.Uint8())).
		SetDate(time.Now().Format("2006-01-02")).
		SetObjectID(int64(userID)).
		SetContent(string(content)).
		Save(ctx)
	if err != nil {
		r.logger.Errorf("[RecordLoginLog] 记录登录日志失败: %v", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	r.logger.Infof("[RecordLoginLog] 登录日志记录成功, userID: %d, success: %v", userID, success)
	return nil
}

// activeTrial 激活试用订阅
// 完整复刻原项目的 activeTrial 函数（Line 796-861）
func (r *oauthRepo) activeTrial(ctx context.Context, tx *ent.Tx, userID int) (*ent.ProxyUserSubscribe, error) {
	r.logger.Infof("[activeTrial] userID: %d", userID)

	// 获取试用订阅配置
	if r.config == nil || r.config.Register == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	trialSubscribeID := r.config.Register.TrialSubscribe
	trialTimeUnit := r.config.Register.TrialTimeUnit
	trialTime := r.config.Register.TrialTime

	r.logger.Infof("[activeTrial] 查询试用订阅模板, subscribeID: %d", trialSubscribeID)

	// 查询试用订阅模板
	sub, err := tx.ProxySubscribe.Query().
		Where(proxysubscribe.IDEQ(trialSubscribeID)).
		Only(ctx)
	if err != nil {
		r.logger.Errorf("[activeTrial] 查询试用订阅模板失败: %v", err)
		return nil, responsecode.NewDatabaseQueryError()
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
	userSubscribe, err := tx.ProxyUserSubscribe.Create().
		SetUserID(int64(userID)).
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
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
	}

	r.logger.Infof("[activeTrial] 试用订阅激活成功, userID: %d, token: %s", userID, subscribeToken)
	return userSubscribe, nil
}

// GenerateJWTToken 生成JWT令牌
// 完整复刻原项目的 generateToken 函数（Line 564-609）
// Claims包含: UserId, SessionId
// Session ID保存到Redis，过期时间与JWT一致
func (r *oauthRepo) GenerateJWTToken(ctx context.Context, userID int) (string, error) {
	r.logger.Infof("[GenerateJWTToken] userID: %d", userID)

	token, err := r.data.issueSessionToken(ctx, int64(userID), sessionTokenOptions{})
	if err != nil {
		r.logger.Errorf("[GenerateJWTToken] 生成JWT token失败: %v", err)
		return "", responsecode.NewKratosError(responsecode.ErrInternalError)
	}
	r.logger.Infof("[GenerateJWTToken] JWT token生成成功, userID: %d", userID)
	return token, nil
}
