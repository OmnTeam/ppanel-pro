package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuser"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	v1 "github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	// Log types matching ProxySystemLog schema
	LogTypeLogin    int8 = 30
	LogTypeRegister int8 = 31

	// Default JWT configuration
	DefaultJWTSecret = "your-secret-key-change-in-production"
	DefaultJWTExpire = 604800 // 7 days in seconds

	// Cache key prefixes
	SessionCacheKeyPrefix     = "session:"
	AuthCodeCacheKeyPrefix    = "auth_code:"
	AuthCodeTelCacheKeyPrefix = "auth_code_telephone:"

	// Verify types
	VerifyTypeRegister  = "register"
	VerifyTypeSecurity  = "security"
	VerifyTypeResetPass = "reset_password"
)

type authRepo struct {
	data   *Data
	log    *log.Helper
	config *conf.Application
}

// LoginLog represents login log data
type LoginLog struct {
	Method    string `json:"method"`
	LoginIP   string `json:"login_ip"`
	UserAgent string `json:"user_agent"`
	Success   bool   `json:"success"`
	Timestamp int64  `json:"timestamp"`
}

// RegisterLog represents register log data
type RegisterLog struct {
	AuthMethod string `json:"auth_method"`
	Identifier string `json:"identifier"`
	RegisterIP string `json:"register_ip"`
	UserAgent  string `json:"user_agent"`
	Timestamp  int64  `json:"timestamp"`
}

// NewAuthRepo creates a new auth repository
func NewAuthRepo(data *Data, config *conf.Application, logger log.Logger) v1.AuthRepo {
	return &authRepo{
		data:   data,
		config: config,
		log:    log.NewHelper(log.With(logger, "module", "data/auth")),
	}
}

// getJWTConfig returns JWT secret and expiry from environment or defaults
func (r *authRepo) getJWTConfig() (string, int64) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = DefaultJWTSecret
	}

	expireStr := os.Getenv("JWT_EXPIRE")
	expire := int64(DefaultJWTExpire)
	if expireStr != "" {
		if exp, err := strconv.ParseInt(expireStr, 10, 64); err == nil {
			expire = exp
		}
	}

	return secret, expire
}

// CheckUserExistByEmail checks if user exists by email
func (r *authRepo) CheckUserExistByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("email"),
			proxyuserauthmethod.AuthIdentifierEQ(email),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("CheckUserExistByEmail failed", "error", err, "email", email)
		return false, err
	}

	return count > 0, nil
}

// CheckUserExistByTelephone checks if user exists by telephone (E.164 format)
func (r *authRepo) CheckUserExistByTelephone(ctx context.Context, telephoneAreaCode, telephone string) (bool, error) {
	// Format phone to E.164
	phoneNumber, err := phone.FormatToE164(telephoneAreaCode, telephone)
	if err != nil {
		r.log.Errorw("CheckUserExistByTelephone: invalid phone", "error", err, "telephone", telephone)
		return false, fmt.Errorf("invalid phone number: %w", err)
	}

	count, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("mobile"),
			proxyuserauthmethod.AuthIdentifierEQ(phoneNumber),
		).
		Count(ctx)

	if err != nil {
		r.log.Errorw("CheckUserExistByTelephone failed", "error", err, "telephone", phoneNumber)
		return false, err
	}

	return count > 0, nil
}

// UserLogin logs in user with email and password
func (r *authRepo) UserLogin(ctx context.Context, email, password, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	loginStatus := false

	// Deferred login logging
	defer func() {
		if userID != 0 {
			r.logLogin(ctx, int(userID), "email", ip, userAgent, loginStatus)
		}
	}()

	// Find user by email
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("email"),
			proxyuserauthmethod.AuthIdentifierEQ(email),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Warnw("UserLogin: user not found", "email", email)
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		r.log.Errorw("UserLogin: query auth method failed", "error", err, "email", email)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	userID = authMethod.UserID

	// Get user and verify password
	user, err := r.data.db.ProxyUser.Get(ctx, authMethod.UserID)
	if err != nil {
		r.log.Errorw("UserLogin: get user failed", "error", err, "user_id", authMethod.UserID)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Verify password
	if !tool.VerifyPassWord(password, user.Password) {
		r.log.Warnw("UserLogin: invalid password", "email", email)
		return nil, responsecode.NewKratosError(responsecode.ErrPasswordIncorrect)
	}

	// Check if user is enabled
	if !user.Enable {
		r.log.Warnw("UserLogin: user disabled", "user_id", user.ID)
		return nil, fmt.Errorf("user account is disabled")
	}

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"UserId":    user.ID,
		"SessionId": sessionID,
	}
	token, err := tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("UserLogin: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("UserLogin: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	loginStatus = true

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// TelephoneLogin logs in user with telephone and password or code
func (r *authRepo) TelephoneLogin(ctx context.Context, telephoneAreaCode, telephone, password, telephoneCode, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	loginStatus := false

	// Check if mobile login is enabled
	if r.config != nil && r.config.Mobile != nil && !r.config.Mobile.Enable {
		return nil, fmt.Errorf("mobile login is not enabled")
	}

	// Format phone to E.164
	phoneNumber, err := phone.FormatToE164(telephoneAreaCode, telephone)
	if err != nil {
		r.log.Errorw("TelephoneLogin: invalid phone", "error", err, "telephone", telephone)
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Deferred login logging
	defer func() {
		if userID != 0 {
			r.logLogin(ctx, int(userID), "mobile", ip, userAgent, loginStatus)
		}
	}()

	// Find user by phone
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("mobile"),
			proxyuserauthmethod.AuthIdentifierEQ(phoneNumber),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Warnw("TelephoneLogin: user not found", "phone", phoneNumber)
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		r.log.Errorw("TelephoneLogin: query auth method failed", "error", err, "phone", phoneNumber)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	userID = authMethod.UserID

	// Get user
	user, err := r.data.db.ProxyUser.Get(ctx, authMethod.UserID)
	if err != nil {
		r.log.Errorw("TelephoneLogin: get user failed", "error", err, "user_id", authMethod.UserID)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Verify password or telephone code
	if password == "" && telephoneCode == "" {
		return nil, fmt.Errorf("password or verification code is required")
	}

	if telephoneCode != "" {
		// Verify telephone code from Redis
		cacheKey := fmt.Sprintf("%s%s:%s", AuthCodeTelCacheKeyPrefix, VerifyTypeSecurity, phoneNumber)
		value, err := r.data.rdb.Get(ctx, cacheKey).Result()
		if err != nil {
			r.log.Warnw("TelephoneLogin: verification code not found", "cache_key", cacheKey)
			return nil, fmt.Errorf("invalid verification code")
		}

		var payload CacheKeyPayload
		if err := json.Unmarshal([]byte(value), &payload); err != nil {
			r.log.Errorw("TelephoneLogin: unmarshal code failed", "error", err)
			return nil, fmt.Errorf("invalid verification code")
		}

		if payload.Code != telephoneCode {
			r.log.Warnw("TelephoneLogin: code mismatch", "expected", payload.Code, "got", telephoneCode)
			return nil, fmt.Errorf("invalid verification code")
		}

		// Delete used code
		r.data.rdb.Del(ctx, cacheKey)
	} else {
		// Verify password
		if !tool.VerifyPassWord(password, user.Password) {
			r.log.Warnw("TelephoneLogin: invalid password", "phone", phoneNumber)
			return nil, responsecode.NewKratosError(responsecode.ErrPasswordIncorrect)
		}
	}

	// Check if user is enabled
	if !user.Enable {
		r.log.Warnw("TelephoneLogin: user disabled", "user_id", user.ID)
		return nil, fmt.Errorf("user account is disabled")
	}

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
	}
	token, err := tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("TelephoneLogin: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("TelephoneLogin: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	loginStatus = true

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// UserRegister registers a new user with email
func (r *authRepo) UserRegister(ctx context.Context, email, password, invite, code, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	var token string

	// Deferred logging
	defer func() {
		if userID != 0 && token != "" {
			r.logLogin(ctx, int(userID), "email", ip, userAgent, true)
			r.logRegister(ctx, int(userID), "email", email, ip, userAgent)
		}
	}()

	// Check if registration is stopped
	if r.config != nil && r.config.Register != nil && r.config.Register.StopRegister {
		return nil, fmt.Errorf("registration is currently disabled")
	}

	// Validate invite code if required
	var refererID *int64
	if invite != "" {
		referer, err := r.data.db.ProxyUser.Query().
			Where(proxyuser.ReferCodeEQ(invite)).
			Only(ctx)
		if err != nil {
			r.log.Warnw("UserRegister: invalid invite code", "invite", invite)
			return nil, fmt.Errorf("invalid invite code")
		}
		tempID := int64(referer.ID)
		refererID = &tempID
	} else {
		// Check if invite is forced
		if r.config != nil && r.config.Invite != nil && r.config.Invite.ForcedInvite {
			return nil, fmt.Errorf("invite code is required")
		}
	}

	// Verify email code if enabled
	if r.config != nil && r.config.Email != nil && r.config.Email.EnableVerify {
		cacheKey := fmt.Sprintf("%s%s:%s", AuthCodeCacheKeyPrefix, VerifyTypeRegister, email)
		value, err := r.data.rdb.Get(ctx, cacheKey).Result()
		if err != nil {
			r.log.Warnw("UserRegister: verification code not found", "cache_key", cacheKey)
			return nil, fmt.Errorf("invalid verification code")
		}

		var payload CacheKeyPayload
		if err := json.Unmarshal([]byte(value), &payload); err != nil {
			r.log.Errorw("UserRegister: unmarshal code failed", "error", err)
			return nil, fmt.Errorf("invalid verification code")
		}

		if payload.Code != code {
			r.log.Warnw("UserRegister: code mismatch", "expected", payload.Code, "got", code)
			return nil, fmt.Errorf("invalid verification code")
		}
	}

	// Check if user already exists
	exists, err := r.CheckUserExistByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Encode password
	encodedPassword := tool.EncodePassWord(password)

	// Get only_first_purchase config
	onlyFirstPurchase := true
	if r.config != nil && r.config.Invite != nil {
		onlyFirstPurchase = r.config.Invite.OnlyFirstPurchase
	}

	// Start transaction
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		r.log.Errorw("UserRegister: failed to start transaction", "error", err)
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Create user
	userCreate := tx.ProxyUser.Create().
		SetPassword(encodedPassword).
		SetAlgo("default").
		SetOnlyFirstPurchase(onlyFirstPurchase)

	if refererID != nil {
		userCreate = userCreate.SetNillableRefererID(refererID)
	}

	user, err := userCreate.Save(ctx)
	if err != nil {
		r.log.Errorw("UserRegister: failed to create user", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Generate and update refer code
	referCode := tool.GenerateReferCode(user.ID)
	user, err = tx.ProxyUser.UpdateOneID(user.ID).
		SetReferCode(referCode).
		Save(ctx)
	if err != nil {
		r.log.Errorw("UserRegister: failed to update refer code", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Create auth method record
	verified := false
	if r.config != nil && r.config.Email != nil {
		verified = r.config.Email.EnableVerify
	}

	_, err = tx.ProxyUserAuthMethod.Create().
		SetUserID(user.ID).
		SetAuthType("email").
		SetAuthIdentifier(email).
		SetVerified(verified).
		Save(ctx)

	if err != nil {
		r.log.Errorw("UserRegister: failed to create auth method", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		r.log.Errorw("UserRegister: failed to commit transaction", "error", err)
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	userID = user.ID

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
	}
	token, err = tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("UserRegister: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("UserRegister: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// TelephoneRegister registers a new user with telephone
func (r *authRepo) TelephoneRegister(ctx context.Context, telephoneAreaCode, telephone, password, invite, code, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	var token string

	// Check if mobile registration is enabled
	if r.config != nil && r.config.Mobile != nil && !r.config.Mobile.Enable {
		return nil, fmt.Errorf("mobile registration is not enabled")
	}

	// Check if registration is stopped
	if r.config != nil && r.config.Register != nil && r.config.Register.StopRegister {
		return nil, fmt.Errorf("registration is currently disabled")
	}

	// Format phone to E.164
	phoneNumber, err := phone.FormatToE164(telephoneAreaCode, telephone)
	if err != nil {
		r.log.Errorw("TelephoneRegister: invalid phone", "error", err, "telephone", telephone)
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Deferred logging
	defer func() {
		if userID != 0 && token != "" {
			r.logLogin(ctx, int(userID), "mobile", ip, userAgent, true)
			r.logRegister(ctx, int(userID), "mobile", phoneNumber, ip, userAgent)
		}
	}()

	// Verify telephone code (required for phone registration)
	cacheKey := fmt.Sprintf("%s%s:%s", AuthCodeTelCacheKeyPrefix, VerifyTypeRegister, phoneNumber)
	value, err := r.data.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		r.log.Warnw("TelephoneRegister: verification code not found", "cache_key", cacheKey)
		return nil, fmt.Errorf("invalid verification code")
	}

	var payload CacheKeyPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		r.log.Errorw("TelephoneRegister: unmarshal code failed", "error", err)
		return nil, fmt.Errorf("invalid verification code")
	}

	if payload.Code != code {
		r.log.Warnw("TelephoneRegister: code mismatch", "expected", payload.Code, "got", code)
		return nil, fmt.Errorf("invalid verification code")
	}

	// Delete used code
	r.data.rdb.Del(ctx, cacheKey)

	// Check if user already exists
	exists, err := r.CheckUserExistByTelephone(ctx, telephoneAreaCode, telephone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("phone number already registered")
	}

	// Validate invite code if required
	var refererID *int64
	if invite != "" {
		referer, err := r.data.db.ProxyUser.Query().
			Where(proxyuser.ReferCodeEQ(invite)).
			Only(ctx)
		if err != nil {
			r.log.Warnw("TelephoneRegister: invalid invite code", "invite", invite)
			return nil, fmt.Errorf("invalid invite code")
		}
		tempID := int64(referer.ID)
		refererID = &tempID
	} else {
		// Check if invite is forced
		if r.config != nil && r.config.Invite != nil && r.config.Invite.ForcedInvite {
			return nil, fmt.Errorf("invite code is required")
		}
	}

	// Encode password
	encodedPassword := tool.EncodePassWord(password)

	// Get only_first_purchase config
	onlyFirstPurchase := true
	if r.config != nil && r.config.Invite != nil {
		onlyFirstPurchase = r.config.Invite.OnlyFirstPurchase
	}

	// Start transaction
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		r.log.Errorw("TelephoneRegister: failed to start transaction", "error", err)
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Create user
	userCreate := tx.ProxyUser.Create().
		SetPassword(encodedPassword).
		SetAlgo("default").
		SetOnlyFirstPurchase(onlyFirstPurchase)

	if refererID != nil {
		userCreate = userCreate.SetNillableRefererID(refererID)
	}

	user, err := userCreate.Save(ctx)
	if err != nil {
		r.log.Errorw("TelephoneRegister: failed to create user", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Generate and update refer code
	referCode := tool.GenerateReferCode(user.ID)
	user, err = tx.ProxyUser.UpdateOneID(user.ID).
		SetReferCode(referCode).
		Save(ctx)
	if err != nil {
		r.log.Errorw("TelephoneRegister: failed to update refer code", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Create auth method record (phone is verified by SMS code)
	_, err = tx.ProxyUserAuthMethod.Create().
		SetUserID(user.ID).
		SetAuthType("mobile").
		SetAuthIdentifier(phoneNumber).
		SetVerified(true).
		Save(ctx)

	if err != nil {
		r.log.Errorw("TelephoneRegister: failed to create auth method", "error", err)
		tx.Rollback()
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		r.log.Errorw("TelephoneRegister: failed to commit transaction", "error", err)
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	userID = user.ID

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
	}
	token, err = tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("TelephoneRegister: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("TelephoneRegister: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// ResetPassword resets user password with email
func (r *authRepo) ResetPassword(ctx context.Context, email, password, code, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	loginStatus := false

	// Deferred login logging
	defer func() {
		if userID != 0 && loginStatus {
			r.logLogin(ctx, int(userID), "email", ip, userAgent, loginStatus)
		}
	}()

	// Verify email code
	cacheKey := fmt.Sprintf("%s%s:%s", AuthCodeCacheKeyPrefix, VerifyTypeSecurity, email)
	value, err := r.data.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		r.log.Warnw("ResetPassword: verification code not found", "cache_key", cacheKey)
		return nil, fmt.Errorf("invalid verification code")
	}

	var payload CacheKeyPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		r.log.Errorw("ResetPassword: unmarshal code failed", "error", err)
		return nil, fmt.Errorf("invalid verification code")
	}

	if payload.Code != code {
		r.log.Warnw("ResetPassword: code mismatch", "expected", payload.Code, "got", code)
		return nil, fmt.Errorf("invalid verification code")
	}

	// Find user by email
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("email"),
			proxyuserauthmethod.AuthIdentifierEQ(email),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Warnw("ResetPassword: user not found", "email", email)
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		r.log.Errorw("ResetPassword: query auth method failed", "error", err, "email", email)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	userID = authMethod.UserID

	// Get user
	user, err := r.data.db.ProxyUser.Get(ctx, authMethod.UserID)
	if err != nil {
		r.log.Errorw("ResetPassword: get user failed", "error", err, "user_id", authMethod.UserID)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	// Update password
	encodedPassword := tool.EncodePassWord(password)
	_, err = r.data.db.ProxyUser.UpdateOneID(user.ID).
		SetPassword(encodedPassword).
		Save(ctx)
	if err != nil {
		r.log.Errorw("ResetPassword: update password failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
	}
	token, err := tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("ResetPassword: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("ResetPassword: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	loginStatus = true

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// TelephoneResetPassword resets user password with telephone
func (r *authRepo) TelephoneResetPassword(ctx context.Context, telephoneAreaCode, telephone, password, code, ip, userAgent string) (*v1.LoginResult, error) {
	var userID int64
	loginStatus := false

	// Check if mobile is enabled
	if r.config != nil && r.config.Mobile != nil && !r.config.Mobile.Enable {
		return nil, fmt.Errorf("mobile authentication is not enabled")
	}

	// Format phone to E.164
	phoneNumber, err := phone.FormatToE164(telephoneAreaCode, telephone)
	if err != nil {
		r.log.Errorw("TelephoneResetPassword: invalid phone", "error", err, "telephone", telephone)
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Deferred login logging
	defer func() {
		if userID != 0 && loginStatus {
			r.logLogin(ctx, int(userID), "mobile", ip, userAgent, loginStatus)
		}
	}()

	// Verify telephone code
	cacheKey := fmt.Sprintf("%s%s:%s", AuthCodeTelCacheKeyPrefix, VerifyTypeSecurity, phoneNumber)
	value, err := r.data.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		r.log.Warnw("TelephoneResetPassword: verification code not found", "cache_key", cacheKey)
		return nil, fmt.Errorf("invalid verification code")
	}

	var payload CacheKeyPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		r.log.Errorw("TelephoneResetPassword: unmarshal code failed", "error", err)
		return nil, fmt.Errorf("invalid verification code")
	}

	if payload.Code != code {
		r.log.Warnw("TelephoneResetPassword: code mismatch", "expected", payload.Code, "got", code)
		return nil, fmt.Errorf("invalid verification code")
	}

	// Find user by phone
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthTypeEQ("mobile"),
			proxyuserauthmethod.AuthIdentifierEQ(phoneNumber),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Warnw("TelephoneResetPassword: user not found", "phone", phoneNumber)
			return nil, responsecode.NewKratosError(responsecode.ErrUserNotFound)
		}
		r.log.Errorw("TelephoneResetPassword: query auth method failed", "error", err, "phone", phoneNumber)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	userID = authMethod.UserID

	// Get user
	user, err := r.data.db.ProxyUser.Get(ctx, authMethod.UserID)
	if err != nil {
		r.log.Errorw("TelephoneResetPassword: get user failed", "error", err, "user_id", authMethod.UserID)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	// Update password
	encodedPassword := tool.EncodePassWord(password)
	_, err = r.data.db.ProxyUser.UpdateOneID(user.ID).
		SetPassword(encodedPassword).
		Save(ctx)
	if err != nil {
		r.log.Errorw("TelephoneResetPassword: update password failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("reset password failed: %w", err)
	}

	// Generate session ID and token
	sessionID := tool.GenerateUUID()
	secret, expire := r.getJWTConfig()

	claims := map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
	}
	token, err := tool.GenerateJWT(secret, expire, claims)
	if err != nil {
		r.log.Errorw("TelephoneResetPassword: token generation failed", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("%s%s", SessionCacheKeyPrefix, sessionID)
	if err := r.data.rdb.Set(ctx, sessionKey, user.ID, time.Duration(expire)*time.Second).Err(); err != nil {
		r.log.Errorw("TelephoneResetPassword: set session failed", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session storage failed: %w", err)
	}

	loginStatus = true

	return &v1.LoginResult{
		Token: token,
	}, nil
}

// logLogin logs login activity (deferred)
func (r *authRepo) logLogin(ctx context.Context, userID int, method, ip, userAgent string, success bool) {
	loginLog := LoginLog{
		Method:    method,
		LoginIP:   ip,
		UserAgent: userAgent,
		Success:   success,
		Timestamp: time.Now().UnixMilli(),
	}

	content, err := json.Marshal(loginLog)
	if err != nil {
		r.log.Errorw("logLogin: marshal failed", "error", err, "user_id", userID)
		return
	}

	_, err = r.data.db.ProxySystemLog.Create().
		SetType(LogTypeLogin).
		SetDate(time.Now().Format("2006-01-02")).
		SetObjectID(int64(userID)).
		SetContent(string(content)).
		Save(ctx)

	if err != nil {
		r.log.Errorw("logLogin: save failed", "error", err, "user_id", userID)
	}
}

// logRegister logs registration activity (deferred)
func (r *authRepo) logRegister(ctx context.Context, userID int, authMethod, identifier, ip, userAgent string) {
	registerLog := RegisterLog{
		AuthMethod: authMethod,
		Identifier: identifier,
		RegisterIP: ip,
		UserAgent:  userAgent,
		Timestamp:  time.Now().UnixMilli(),
	}

	content, err := json.Marshal(registerLog)
	if err != nil {
		r.log.Errorw("logRegister: marshal failed", "error", err, "user_id", userID)
		return
	}

	_, err = r.data.db.ProxySystemLog.Create().
		SetType(LogTypeRegister).
		SetDate(time.Now().Format("2006-01-02")).
		SetObjectID(int64(userID)).
		SetContent(string(content)).
		Save(ctx)

	if err != nil {
		r.log.Errorw("logRegister: save failed", "error", err, "user_id", userID)
	}
}
