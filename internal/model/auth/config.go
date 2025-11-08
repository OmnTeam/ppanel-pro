package auth

import (
	"encoding/json"
)

// AppleAuthConfig Apple 认证配置
type AppleAuthConfig struct {
	TeamID       string `json:"team_id"`
	KeyID        string `json:"key_id"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

func (l *AppleAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(AppleAuthConfig))
	}
	return string(bytes)
}

func (l *AppleAuthConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// GoogleAuthConfig Google 认证配置
type GoogleAuthConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

func (l *GoogleAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(GoogleAuthConfig))
	}
	return string(bytes)
}

func (l *GoogleAuthConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// GithubAuthConfig Github 认证配置
type GithubAuthConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

func (l *GithubAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(GithubAuthConfig))
	}
	return string(bytes)
}

func (l *GithubAuthConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// FacebookAuthConfig Facebook 认证配置
type FacebookAuthConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

func (l *FacebookAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(FacebookAuthConfig))
	}
	return string(bytes)
}

func (l *FacebookAuthConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// TelegramAuthConfig Telegram 认证配置
type TelegramAuthConfig struct {
	BotToken      string `json:"bot_token"`
	EnableNotify  bool   `json:"enable_notify"`
	WebHookDomain string `json:"webhook_domain"`
}

func (l *TelegramAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(TelegramAuthConfig))
	}
	return string(bytes)
}

func (l *TelegramAuthConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// EmailAuthConfig 邮箱认证配置
type EmailAuthConfig struct {
	Platform                   string      `json:"platform"`
	PlatformConfig             interface{} `json:"platform_config"`
	EnableVerify               bool        `json:"enable_verify"`
	EnableNotify               bool        `json:"enable_notify"`
	EnableDomainSuffix         bool        `json:"enable_domain_suffix"`
	DomainSuffixList           string      `json:"domain_suffix_list"`
	VerifyEmailTemplate        string      `json:"verify_email_template"`
	ExpirationEmailTemplate    string      `json:"expiration_email_template"`
	MaintenanceEmailTemplate   string      `json:"maintenance_email_template"`
	TrafficExceedEmailTemplate string      `json:"traffic_exceed_email_template"`
}

func (l *EmailAuthConfig) Marshal() string {
	// 设置默认模板
	if l.ExpirationEmailTemplate == "" {
		l.ExpirationEmailTemplate = DefaultExpirationEmailTemplate
	}
	if l.MaintenanceEmailTemplate == "" {
		l.MaintenanceEmailTemplate = DefaultMaintenanceEmailTemplate
	}
	if l.TrafficExceedEmailTemplate == "" {
		l.TrafficExceedEmailTemplate = DefaultTrafficExceedEmailTemplate
	}
	if l.VerifyEmailTemplate == "" {
		l.VerifyEmailTemplate = DefaultEmailVerifyTemplate
	}

	bytes, err := json.Marshal(l)
	if err != nil {
		config := &EmailAuthConfig{
			Platform:                   "smtp",
			PlatformConfig:             new(SMTPConfig),
			EnableVerify:               true,
			EnableNotify:               true,
			EnableDomainSuffix:         false,
			DomainSuffixList:           "",
			VerifyEmailTemplate:        DefaultEmailVerifyTemplate,
			ExpirationEmailTemplate:    DefaultExpirationEmailTemplate,
			MaintenanceEmailTemplate:   DefaultMaintenanceEmailTemplate,
			TrafficExceedEmailTemplate: DefaultTrafficExceedEmailTemplate,
		}
		bytes, _ = json.Marshal(config)
	}
	return string(bytes)
}

func (l *EmailAuthConfig) Unmarshal(data string) {
	err := json.Unmarshal([]byte(data), &l)
	if err != nil {
		config := &EmailAuthConfig{
			Platform:                   "smtp",
			PlatformConfig:             new(SMTPConfig),
			EnableVerify:               true,
			EnableNotify:               true,
			EnableDomainSuffix:         false,
			DomainSuffixList:           "",
			VerifyEmailTemplate:        DefaultEmailVerifyTemplate,
			ExpirationEmailTemplate:    DefaultExpirationEmailTemplate,
			MaintenanceEmailTemplate:   DefaultMaintenanceEmailTemplate,
			TrafficExceedEmailTemplate: DefaultTrafficExceedEmailTemplate,
		}
		_ = json.Unmarshal([]byte(config.Marshal()), &l)
	}
}

// SMTPConfig Email SMTP 配置
type SMTPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	From string `json:"from"`
	SSL  bool   `json:"ssl"`
}

func (l *SMTPConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(SMTPConfig))
	}
	return string(bytes)
}

func (l *SMTPConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), &l)
}

// MobileAuthConfig 手机认证配置
type MobileAuthConfig struct {
	Platform        string      `json:"platform"`
	PlatformConfig  interface{} `json:"platform_config"`
	EnableWhitelist bool        `json:"enable_whitelist"`
	Whitelist       []string    `json:"whitelist"`
}

func (l *MobileAuthConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		config := &MobileAuthConfig{
			Platform:        "alibaba_cloud",
			PlatformConfig:  new(AlibabaCloudConfig),
			EnableWhitelist: false,
			Whitelist:       []string{},
		}
		bytes, _ = json.Marshal(config)
	}
	return string(bytes)
}

func (l *MobileAuthConfig) Unmarshal(data string) {
	err := json.Unmarshal([]byte(data), &l)
	if err != nil {
		config := &MobileAuthConfig{
			Platform:        "alibaba_cloud",
			PlatformConfig:  new(AlibabaCloudConfig),
			EnableWhitelist: false,
			Whitelist:       []string{},
		}
		_ = json.Unmarshal([]byte(config.Marshal()), &l)
	}
}

// AlibabaCloudConfig 阿里云配置
type AlibabaCloudConfig struct {
	Access       string `json:"access"`
	Secret       string `json:"secret"`
	SignName     string `json:"sign_name"`
	Endpoint     string `json:"endpoint"`
	TemplateCode string `json:"template_code"`
}

func (l *AlibabaCloudConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(AlibabaCloudConfig))
	}
	return string(bytes)
}

func (l *AlibabaCloudConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// SmsbaoConfig 短信宝配置
type SmsbaoConfig struct {
	Access   string `json:"access"`
	Secret   string `json:"secret"`
	Template string `json:"template"`
}

func (l *SmsbaoConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(SmsbaoConfig))
	}
	return string(bytes)
}

func (l *SmsbaoConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// AbosendConfig Abosend 配置
type AbosendConfig struct {
	ApiDomain string `json:"api_domain"`
	Access    string `json:"access"`
	Secret    string `json:"secret"`
	Template  string `json:"template"`
}

func (l *AbosendConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(AbosendConfig))
	}
	return string(bytes)
}

func (l *AbosendConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// TwilioConfig Twilio 配置
type TwilioConfig struct {
	Access      string `json:"access"`
	Secret      string `json:"secret"`
	PhoneNumber string `json:"phone_number"`
	Template    string `json:"template"`
}

func (l *TwilioConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(TwilioConfig))
	}
	return string(bytes)
}

func (l *TwilioConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// DeviceConfig 设备认证配置
type DeviceConfig struct {
	ShowAds        bool   `json:"show_ads"`
	OnlyRealDevice bool   `json:"only_real_device"`
	EnableSecurity bool   `json:"enable_security"`
	SecuritySecret string `json:"security_secret"`
}

func (l *DeviceConfig) Marshal() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		bytes, _ = json.Marshal(new(DeviceConfig))
	}
	return string(bytes)
}

func (l *DeviceConfig) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// 默认邮件模板
const (
	DefaultEmailVerifyTemplate = `
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>验证您的邮箱</h2>
    <p>您的验证码是: <strong>{{.Code}}</strong></p>
    <p>此验证码将在 {{.ExpireMinutes}} 分钟后过期。</p>
</div>`

	DefaultExpirationEmailTemplate = `
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>订阅即将过期</h2>
    <p>您的订阅将在 {{.Days}} 天后过期。</p>
    <p>请及时续费以继续使用服务。</p>
</div>`

	DefaultMaintenanceEmailTemplate = `
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>系统维护通知</h2>
    <p>系统将在 {{.Time}} 进行维护。</p>
    <p>维护期间服务将暂时不可用，请您谅解。</p>
</div>`

	DefaultTrafficExceedEmailTemplate = `
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>流量使用提醒</h2>
    <p>您的流量使用已达到 {{.Percentage}}%。</p>
    <p>剩余流量: {{.Remaining}}，请合理使用。</p>
</div>`
)
