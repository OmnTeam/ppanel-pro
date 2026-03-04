package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyads"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	v1 "github.com/OmnTeam/ppanel-pro/internal/biz/common"
	"github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/OmnTeam/ppanel-pro/pkg/limit"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

type commonRepo struct {
	data *Data
	log  *log.Helper
}

// CacheKeyPayload stores verification code in Redis
type CacheKeyPayload struct {
	Code   string `json:"code"`
	LastAt int    `json:"lastAt"`
}

// NewCommonRepo creates a new common repository
func NewCommonRepo(data *Data, logger log.Logger) v1.CommonRepo {
	return &commonRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/common")),
	}
}

// GetAdsList gets ads list by status
func (r *commonRepo) GetAdsList(ctx context.Context, status int) ([]*v1.Ads, error) {
	// Query ads with status filter, limit to 200 items
	entAds, err := r.data.db.ProxyAds.Query().
		Where(
			proxyads.Status(int8(status)),
		).
		Limit(200).
		All(ctx)
	if err != nil {
		r.log.Errorw("GetAdsList query failed", "error", err, "status", status)
		return nil, err
	}

	// Convert ent objects to biz objects
	list := make([]*v1.Ads, len(entAds))
	for i, entAd := range entAds {
		list[i] = &v1.Ads{
			ID:          entAd.ID,
			Title:       entAd.Title,
			Type:        entAd.Type,
			Content:     entAd.Content,
			Description: entAd.Description,
			TargetURL:   entAd.TargetURL,
			StartTime:   entAd.StartTime.Unix(),
			EndTime:     entAd.EndTime.Unix(),
			Status:      int(entAd.Status),
			CreatedAt:   entAd.CreatedAt.Unix(),
			UpdatedAt:   entAd.UpdatedAt.Unix(),
		}
	}

	return list, nil
}

// GetClientList retrieves subscribe application list
func (r *commonRepo) GetClientList(ctx context.Context) ([]*v1.SubscribeClient, error) {
	entClients, err := r.data.db.ProxySubscribeApplication.Query().
		All(ctx)
	if err != nil {
		r.log.Errorw("GetClientList query failed", "error", err)
		return nil, err
	}

	result := make([]*v1.SubscribeClient, 0, len(entClients))
	for _, entClient := range entClients {
		// Parse download_link JSON
		var downloadLink v1.DownloadLink
		if entClient.DownloadLink != "" {
			if err := json.Unmarshal([]byte(entClient.DownloadLink), &downloadLink); err != nil {
				r.log.Warnw("Failed to unmarshal download_link", "error", err, "id", entClient.ID)
			}
		}

		client := &v1.SubscribeClient{
			ID:           int64(entClient.ID),
			Name:         entClient.Name,
			Scheme:       entClient.Scheme,
			IsDefault:    entClient.IsDefault,
			DownloadLink: downloadLink,
		}
		if entClient.Description != nil {
			client.Description = *entClient.Description
		}
		if entClient.Icon != nil {
			client.Icon = *entClient.Icon
		}
		result = append(result, client)
	}

	return result, nil
}

// GetTosConfig retrieves TOS/Privacy config from proxy_system table
func (r *commonRepo) GetTosConfig(ctx context.Context, key string) (string, error) {
	entSystem, err := r.data.db.ProxySystem.Query().
		Where(
			proxysystem.Category("tos"),
			proxysystem.Key(key),
		).
		First(ctx)
	if err != nil {
		r.log.Warnw("GetTosConfig query failed", "error", err, "key", key)
		// Return empty string if not found (not an error)
		return "", nil
	}

	return entSystem.Value, nil
}

// GetSystemConfigByCategory retrieves system config by category and returns as map
func (r *commonRepo) GetSystemConfigByCategory(ctx context.Context, category string) (map[string]string, error) {
	entConfigs, err := r.data.db.ProxySystem.Query().
		Where(
			proxysystem.Category(category),
		).
		All(ctx)
	if err != nil {
		r.log.Warnw("GetSystemConfigByCategory query failed", "error", err, "category", category)
		// Return empty map if not found (not an error)
		return make(map[string]string), nil
	}

	result := make(map[string]string)
	for _, config := range entConfigs {
		result[config.Key] = config.Value
	}

	return result, nil
}

// GetWebAdConfig retrieves WebAD config
func (r *commonRepo) GetWebAdConfig(ctx context.Context) (bool, error) {
	entSystem, err := r.data.db.ProxySystem.Query().
		Where(
			proxysystem.Key("WebAD"),
		).
		First(ctx)
	if err != nil {
		r.log.Warnw("GetWebAdConfig query failed", "error", err)
		// Return false if not found (not an error)
		return false, nil
	}

	return entSystem.Value == "true", nil
}

// GetEnabledAuthMethods retrieves enabled auth methods
func (r *commonRepo) GetEnabledAuthMethods(ctx context.Context) ([]string, error) {
	entMethods, err := r.data.db.ProxyAuthMethod.Query().
		Where(
			proxyauthmethod.Enabled(true),
		).
		All(ctx)
	if err != nil {
		r.log.Warnw("GetEnabledAuthMethods query failed", "error", err)
		return []string{}, nil
	}

	var methods []string
	for _, method := range entMethods {
		methods = append(methods, method.Method)
	}

	return methods, nil
}

// GetStatistics retrieves system statistics
func (r *commonRepo) GetStatistics(ctx context.Context) (*v1.Statistics, error) {
	// TODO: Implement Redis caching for better performance

	// Query enabled user count from proxy_user table
	// Note: This assumes proxy_user table exists, adjust based on actual schema
	userCount := int64(0)
	// For now, return mock data as user/server tables may not be fully migrated yet
	// userCount will be implemented when user module is complete

	// Query enabled server/node count
	// This will be implemented when server module is complete
	nodeCount := int64(0)

	// Country count and protocol list
	// These require external IP API calls and complex processing
	// Will be implemented in phase 2
	countryCount := int64(0)
	protocols := []string{}

	return &v1.Statistics{
		User:     userCount,
		Node:     nodeCount,
		Country:  countryCount,
		Protocol: protocols,
	}, nil
}

// parseVerifyType converts verify type to string
func parseVerifyType(verifyType int32) string {
	switch verifyType {
	case 1:
		return "register"
	case 2:
		return "security"
	case 3:
		return "reset_password"
	default:
		return "unknown"
	}
}

// SendEmailVerificationCode sends email verification code
func (r *commonRepo) SendEmailVerificationCode(ctx context.Context, email string, verifyType int32) (string, error) {
	// Build cache key: verify_code:{type}:{email}
	cacheKey := fmt.Sprintf("verify_code:%s:%s", parseVerifyType(verifyType), email)

	// Rate limiting: 60-second interval limit (1 request per minute)
	intervalLimiter := limit.NewPeriodLimit(60, 1, r.data.rdb, fmt.Sprintf("send_interval:email:%s:", parseVerifyType(verifyType)))
	permit, err := intervalLimiter.Take(email)
	if err != nil {
		r.log.Errorw("SendEmailVerificationCode interval limiter error", "error", err, "email", email)
		return "", fmt.Errorf("rate limit error: %w", err)
	}
	if !intervalLimiter.ParsePermitState(permit) {
		r.log.Warnw("SendEmailVerificationCode interval limit exceeded", "email", email)
		return "", fmt.Errorf("too many requests, please try again later")
	}

	// Daily limit: 5 emails per day (86400 seconds with alignment)
	dailyLimiter := limit.NewPeriodLimit(86400, 5, r.data.rdb, fmt.Sprintf("send_daily:email:%s:", parseVerifyType(verifyType)), limit.Align())
	permit, err = dailyLimiter.Take(email)
	if err != nil {
		r.log.Errorw("SendEmailVerificationCode daily limiter error", "error", err, "email", email)
		return "", fmt.Errorf("rate limit error: %w", err)
	}
	if !dailyLimiter.ParsePermitState(permit) {
		r.log.Warnw("SendEmailVerificationCode daily limit exceeded", "email", email)
		return "", fmt.Errorf("daily send limit exceeded")
	}

	// Validate user existence based on verify type
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthType("email"),
			proxyuserauthmethod.AuthIdentifier(email),
		).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		r.log.Errorw("SendEmailVerificationCode query user error", "error", err, "email", email)
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// Check user existence requirements
	if parseVerifyType(verifyType) == "register" && authMethod != nil {
		r.log.Warnw("SendEmailVerificationCode user already exists", "email", email)
		return "", fmt.Errorf("email already registered")
	} else if parseVerifyType(verifyType) == "security" && authMethod == nil {
		r.log.Warnw("SendEmailVerificationCode user not found", "email", email)
		return "", fmt.Errorf("email not registered")
	}

	// Generate 6-digit verification code
	code := tool.KeyNew(6, 0)

	// Prepare cache payload
	cachePayload := CacheKeyPayload{
		Code:   code,
		LastAt: int(time.Now().Unix()),
	}

	// Marshal cache payload
	val, err := json.Marshal(cachePayload)
	if err != nil {
		r.log.Errorw("SendEmailVerificationCode marshal error", "error", err, "email", email)
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Save to Redis with 5 minutes expiration
	if err := r.data.rdb.Set(ctx, cacheKey, string(val), 5*time.Minute).Err(); err != nil {
		r.log.Errorw("SendEmailVerificationCode Redis error", "error", err, "cache_key", cacheKey)
		return "", fmt.Errorf("failed to save code: %w", err)
	}

	// Prepare email task payload
	emailPayload := types.SendEmailPayload{
		Type:    types.EmailTypeVerify,
		Email:   email,
		Subject: "Verification Code",
		Content: map[string]interface{}{
			"Type":     verifyType,
			"SiteLogo": "https://example.com/logo.png", // 站点Logo，当前实现与原项目保持一致
			"SiteName": "PPanel Pro",                   // 站点名称，当前实现与原项目保持一致
			"Expire":   5,
			"Code":     code,
		},
	}

	// Marshal email task payload
	payloadBytes, err := json.Marshal(emailPayload)
	if err != nil {
		r.log.Errorw("SendEmailVerificationCode marshal task payload error", "error", err, "email", email)
		return "", fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Create asynq task
	task := asynq.NewTask(types.ForthwithSendEmail, payloadBytes, asynq.MaxRetry(3))

	// Enqueue task
	taskInfo, err := r.data.queue.Enqueue(task)
	if err != nil {
		r.log.Errorw("SendEmailVerificationCode enqueue error", "error", err, "payload", string(payloadBytes))
		return "", fmt.Errorf("failed to enqueue email task: %w", err)
	}

	r.log.Infow("Email verification code sent", "email", email, "code", code, "task_id", taskInfo.ID, "cache_key", cacheKey)

	return code, nil
}

// SendSmsVerificationCode sends SMS verification code
func (r *commonRepo) SendSmsVerificationCode(ctx context.Context, telephone, telephoneArea string, verifyType int32) (string, error) {
	// Format phone number to E.164 format
	phoneNumber, err := phone.FormatToE164(telephoneArea, telephone)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode invalid phone number", "error", err, "telephone", telephone, "area", telephoneArea)
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Build cache key: verify_code_telephone:{type}:{phone}
	cacheKey := fmt.Sprintf("verify_code_telephone:%s:%s", parseVerifyType(verifyType), phoneNumber)

	// Rate limiting: 60-second interval limit (1 request per minute)
	intervalLimiter := limit.NewPeriodLimit(60, 1, r.data.rdb, fmt.Sprintf("send_interval:mobile:%s:", parseVerifyType(verifyType)))
	permit, err := intervalLimiter.Take(phoneNumber)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode interval limiter error", "error", err, "phone", phoneNumber)
		return "", fmt.Errorf("rate limit error: %w", err)
	}
	if !intervalLimiter.ParsePermitState(permit) {
		r.log.Warnw("SendSmsVerificationCode interval limit exceeded", "phone", phoneNumber)
		return "", fmt.Errorf("too many requests, please try again later")
	}

	// Daily limit: 5 SMS per day (86400 seconds with alignment)
	dailyLimiter := limit.NewPeriodLimit(86400, 5, r.data.rdb, fmt.Sprintf("send_daily:mobile:%s:", parseVerifyType(verifyType)), limit.Align())
	permit, err = dailyLimiter.Take(phoneNumber)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode daily limiter error", "error", err, "phone", phoneNumber)
		return "", fmt.Errorf("rate limit error: %w", err)
	}
	if !dailyLimiter.ParsePermitState(permit) {
		r.log.Warnw("SendSmsVerificationCode daily limit exceeded", "phone", phoneNumber)
		return "", fmt.Errorf("daily send limit exceeded")
	}

	// Validate user existence based on verify type
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthType("mobile"),
			proxyuserauthmethod.AuthIdentifier(phoneNumber),
		).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		r.log.Errorw("SendSmsVerificationCode query user error", "error", err, "phone", phoneNumber)
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// Check user existence requirements
	if parseVerifyType(verifyType) == "register" && authMethod != nil {
		r.log.Warnw("SendSmsVerificationCode user already exists", "phone", phoneNumber)
		return "", fmt.Errorf("mobile already registered")
	} else if parseVerifyType(verifyType) == "security" && authMethod == nil {
		r.log.Warnw("SendSmsVerificationCode user not found", "phone", phoneNumber)
		return "", fmt.Errorf("mobile not registered")
	}

	// Generate 6-digit verification code
	code := tool.KeyNew(6, 0)

	// Prepare cache payload
	cachePayload := CacheKeyPayload{
		Code:   code,
		LastAt: int(time.Now().Unix()),
	}

	// Marshal cache payload
	val, err := json.Marshal(cachePayload)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode marshal error", "error", err, "phone", phoneNumber)
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Save to Redis with 5 minutes expiration
	if err := r.data.rdb.Set(ctx, cacheKey, string(val), 5*time.Minute).Err(); err != nil {
		r.log.Errorw("SendSmsVerificationCode Redis error", "error", err, "cache_key", cacheKey)
		return "", fmt.Errorf("failed to save code: %w", err)
	}

	// Prepare SMS task payload
	smsPayload := types.SendSmsPayload{
		Type:          verifyType,
		Telephone:     telephone,
		TelephoneArea: telephoneArea,
		Content:       code,
	}

	// Marshal SMS task payload
	payloadBytes, err := json.Marshal(smsPayload)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode marshal task payload error", "error", err, "phone", phoneNumber)
		return "", fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Create asynq task
	task := asynq.NewTask(types.ForthwithSendSms, payloadBytes)

	// Enqueue task
	taskInfo, err := r.data.queue.Enqueue(task)
	if err != nil {
		r.log.Errorw("SendSmsVerificationCode enqueue error", "error", err, "payload", string(payloadBytes))
		return "", fmt.Errorf("failed to enqueue SMS task: %w", err)
	}

	r.log.Infow("SMS verification code sent", "phone", phoneNumber, "code", code, "task_id", taskInfo.ID, "cache_key", cacheKey)

	return code, nil
}

// CheckVerificationCode checks verification code
func (r *commonRepo) CheckVerificationCode(ctx context.Context, method, account, code string, verifyType int32) (bool, error) {
	var cacheKey string

	if method == "email" {
		cacheKey = fmt.Sprintf("verify_code:%s:%s", parseVerifyType(verifyType), account)
	} else if method == "mobile" {
		// For mobile, account should already include country code
		phoneNumber := account
		if account[0] != '+' {
			phoneNumber = "+" + account
		}
		cacheKey = fmt.Sprintf("verify_code_telephone:%s:%s", parseVerifyType(verifyType), phoneNumber)
	} else {
		r.log.Warnw("CheckVerificationCode invalid method", "method", method)
		return false, nil
	}

	// Get from Redis
	value, err := r.data.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		r.log.Warnw("CheckVerificationCode Redis get error", "error", err, "cache_key", cacheKey)
		return false, nil
	}

	// Unmarshal payload
	var payload CacheKeyPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		r.log.Warnw("CheckVerificationCode unmarshal error", "error", err)
		return false, nil
	}

	// Compare code
	if payload.Code != code {
		r.log.Warnw("CheckVerificationCode code mismatch", "expected", payload.Code, "got", code)
		return false, nil
	}

	r.log.Infow("Verification code validated", "method", method, "account", account)
	return true, nil
}
