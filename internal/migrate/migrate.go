package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/model/auth"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"

	"github.com/go-kratos/kratos/v2/log"
)

// Migrator 数据迁移器
type Migrator struct {
	client   *ent.Client
	logger   *log.Helper
	appConf  *conf.Application
}

// NewMigrator 创建新的迁移器
func NewMigrator(client *ent.Client, logger log.Logger, appConf *conf.Application) *Migrator {
	return &Migrator{
		client:  client,
		logger:  log.NewHelper(logger),
		appConf: appConf,
	}
}

// AutoMigrate 自动迁移数据库结构
func (m *Migrator) AutoMigrate(ctx context.Context) error {
	m.logger.Info("Starting auto migration...")

	if err := m.client.Schema.Create(ctx); err != nil {
		m.logger.Errorf("Failed to create schema: %v", err)
		return fmt.Errorf("failed to create schema: %w", err)
	}

	m.logger.Info("Auto migration completed successfully")
	return nil
}

// AutoMigrateWithData 自动迁移数据库结构并初始化数据
func (m *Migrator) AutoMigrateWithData(ctx context.Context) error {
	m.logger.Info("Starting auto migration with data initialization...")

	// 先执行数据库结构迁移
	if err := m.AutoMigrate(ctx); err != nil {
		return err
	}

	// 初始化基础数据
	if err := m.InitBasicData(ctx); err != nil {
		return fmt.Errorf("failed to initialize basic data: %w", err)
	}

	// 创建默认管理员用户
	if err := m.CreateDefaultAdminUser(ctx); err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	m.logger.Info("Auto migration with data completed successfully")
	return nil
}

// InitBasicData 初始化基础数据
func (m *Migrator) InitBasicData(ctx context.Context) error {
	m.logger.Info("Starting basic data initialization...")

	// 初始化认证方法
	if err := m.initAuthMethods(ctx); err != nil {
		return fmt.Errorf("failed to init auth methods: %w", err)
	}

	// 初始化默认支付方式
	if err := m.initPaymentMethods(ctx); err != nil {
		return fmt.Errorf("failed to init payment methods: %w", err)
	}

	// 初始化系统配置
	if err := m.initSystemConfig(ctx); err != nil {
		return fmt.Errorf("failed to init system config: %w", err)
	}

	m.logger.Info("Basic data initialization completed")
	return nil
}

// CreateDefaultAdminUser 创建默认管理员用户
func (m *Migrator) CreateDefaultAdminUser(ctx context.Context) error {
	// 从配置文件读取管理员凭据
	var email, password, algo string
	if m.appConf != nil && m.appConf.Admin != nil {
		email = m.appConf.Admin.Email
		password = m.appConf.Admin.Password
		algo = m.appConf.Admin.Algo
	}

	// 如果配置为空，使用默认值
	if email == "" {
		email = "admin@example.com"
	}
	if password == "" {
		password = "admin123456"
	}
	if algo == "" {
		algo = "default"
	}

	// 检查是否已存在该邮箱的认证方式
	exist, err := m.client.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.AuthType("email"),
			proxyuserauthmethod.AuthIdentifier(email),
		).
		Exist(ctx)
	if err != nil {
		return err
	}

	if exist {
		m.logger.Infof("Admin user with email %s already exists, skip creation", email)
		return nil
	}

	encodedPwd := tool.EncodePassWord(password)
	referCode := tool.KeyNew(6, 1)

	// 创建管理员用户
	user, err := m.client.ProxyUser.Create().
		SetPassword(encodedPwd).
		SetAlgo(algo).
		SetAvatar("").
		SetBalance(0).
		SetTelegram(0).
		SetReferCode(referCode).
		SetCommission(0).
		SetReferralPercentage(10).
		SetGiftAmount(0).
		SetEnable(true).
		SetIsAdmin(true).
		SetValidEmail(true).
		SetEnableEmailNotify(true).
		SetEnableTelegramNotify(false).
		SetEnableBalanceNotify(true).
		SetEnableLoginNotify(true).
		SetEnableSubscribeNotify(true).
		SetEnableTradeNotify(true).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// 创建邮箱认证方式
	_, err = m.client.ProxyUserAuthMethod.Create().
		SetUserID(user.ID).
		SetAuthType("email").
		SetAuthIdentifier(email).
		SetVerified(false).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create admin auth method: %w", err)
	}

	m.logger.Infof("Default admin user created successfully with email: %s", email)
	m.logger.Infof("Default admin credentials - Email: %s, Password: %s", email, password)
	m.logger.Infof("Please change the default password immediately after first login!")
	return nil
}

// initAuthMethods 初始化认证方法
func (m *Migrator) initAuthMethods(ctx context.Context) error {
	// 邮件认证配置
	emailConfig := auth.EmailAuthConfig{
		Platform:         "smtp",
		EnableVerify:     false,
		EnableNotify:     false,
		EnableDomainSuffix: false,
		DomainSuffixList: "",
		PlatformConfig: auth.SMTPConfig{
			Host:     "",
			Port:     587,
			User:     "",
			Pass:     "",
			From:     "",
			SSL:      false,
		},
		VerifyEmailTemplate:     "",
		ExpirationEmailTemplate: "",
		MaintenanceEmailTemplate: "",
		TrafficExceedEmailTemplate: "",
	}
	emailConfigJSON, _ := json.Marshal(emailConfig)

	// 手机认证配置 - 使用abosend平台
	mobileConfig := auth.MobileAuthConfig{
		Platform:        "abosend",
		EnableWhitelist: false,
		Whitelist:       []string{},
		PlatformConfig: auth.AbosendConfig{
			ApiDomain: "https://smsapi.abosend.com",
			Access:    "UVTtbbTz",
			Secret:    "CVRZQVJLTJWTBDXDWSYSOITEWLUMBRCO",
			Template:  "Your verification code is: {{.code}}",
		},
	}
	mobileConfigJSON, _ := json.Marshal(mobileConfig)

	authMethods := []struct {
		method  string
		config  string
		enabled bool
	}{
		{"email", string(emailConfigJSON), true},
		{"mobile", string(mobileConfigJSON), true},
		{"apple", `{"team_id":"","key_id":"","client_id":"","client_secret":"","redirect_url":""}`, false},
		{"google", `{"client_id":"","client_secret":"","redirect_url":""}`, false},
		{"github", `{"client_id":"","client_secret":"","redirect_url":""}`, false},
		{"telegram", `{"bot_token":"","enable_notify":false,"webhook_domain":""}`, false},
		{"device", `{"show_ads":false,"only_real_device":false,"enable_security":false,"security_secret":""}`, false},
	}

	// 创建认证方法，根据method查询是否已存在
	createdCount := 0
	for _, authMethod := range authMethods {
		// 检查是否已存在该认证方法
		exist, err := m.client.ProxyAuthMethod.Query().
			Where(proxyauthmethod.Method(authMethod.method)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check auth method %s: %w", authMethod.method, err)
		}

		if exist {
			m.logger.Infof("Auth method %s already exists, skip creation", authMethod.method)
			continue
		}

		// 创建认证方法
		_, err = m.client.ProxyAuthMethod.Create().
			SetMethod(authMethod.method).
			SetConfig(authMethod.config).
			SetEnabled(authMethod.enabled).
			SetCreatedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create auth method %s: %w", authMethod.method, err)
		}
		createdCount++
	}

	m.logger.Infof("Successfully initialized %d new auth methods", createdCount)
	return nil
}

// initPaymentMethods 初始化支付方式
func (m *Migrator) initPaymentMethods(ctx context.Context) error {
	// 默认支付方式 - 简化版本，暂时只创建余额支付
	paymentName := "余额支付"
	paymentPlatform := "balance"

	// 检查是否已存在该支付方式
	exist, err := m.client.ProxyPayment.Query().
		Where(proxypayment.Name(paymentName)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check payment method %s: %w", paymentName, err)
	}

	if exist {
		m.logger.Infof("Payment method %s already exists, skip creation", paymentName)
		return nil
	}

	// 创建支付方式
	_, err = m.client.ProxyPayment.Create().
		SetName(paymentName).
		SetPlatform(paymentPlatform).
		SetDescription("使用账户余额进行支付").
		SetIcon("").
		SetDomain("").
		SetConfig("{}").
		SetFeeMode(0).
		SetFeePercent(0).
		SetFeeAmount(0).
		SetEnable(true).
		SetToken("").
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create balance payment: %w", err)
	}

	m.logger.Infof("Successfully created payment method: %s", paymentName)
	return nil
}

// initSystemConfig 初始化系统配置
func (m *Migrator) initSystemConfig(ctx context.Context) error {
	// 简化的系统配置
	systemConfigs := []struct {
		category string
		key      string
		value    string
		typ      string
		desc     string
	}{
		{"site", "site_name", "PPanel Pro", "string", "站点名称"},
		{"site", "site_desc", "Professional Panel Management System", "string", "站点描述"},
		{"currency", "default_currency", "CNY", "string", "默认货币"},
		{"currency", "currency_symbol", "¥", "string", "货币符号"},
	}

	// 创建系统配置，根据category和key查询是否已存在
	createdCount := 0
	for _, config := range systemConfigs {
		// 检查是否已存在该系统配置
		exist, err := m.client.ProxySystem.Query().
			Where(
				proxysystem.Category(config.category),
				proxysystem.Key(config.key),
			).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check system config %s.%s: %w", config.category, config.key, err)
		}

		if exist {
			m.logger.Infof("System config %s.%s already exists, skip creation", config.category, config.key)
			continue
		}

		// 创建系统配置
		_, err = m.client.ProxySystem.Create().
			SetCategory(config.category).
			SetKey(config.key).
			SetValue(config.value).
			SetType(config.typ).
			SetDesc(config.desc).
			SetCreatedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create system config %s.%s: %w", config.category, config.key, err)
		}
		createdCount++
	}

	m.logger.Infof("Successfully created %d new system config items", createdCount)
	return nil
}

