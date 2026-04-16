package data

import "encoding/json"

// StripeConfig Stripe支付配置
type StripeConfig struct {
	PublicKey     string `json:"public_key"`
	SecretKey     string `json:"secret_key"`
	WebhookSecret string `json:"webhook_secret"`
	Payment       string `json:"payment"` // card/alipay/wechat_pay
}

func (c *StripeConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

// AlipayF2FConfig 支付宝当面付配置
type AlipayF2FConfig struct {
	AppId       string `json:"app_id"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
	InvoiceName string `json:"invoice_name"`
	Sandbox     bool   `json:"sandbox"`
}

func (c *AlipayF2FConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

// EPayConfig EPay支付配置
type EPayConfig struct {
	Pid  string `json:"pid"`
	Url  string `json:"url"`
	Key  string `json:"key"`
	Type string `json:"type"`
}

func (c *EPayConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

// CryptoSaaSConfig CryptoSaaS加密货币支付配置
type CryptoSaaSConfig struct {
	Endpoint  string `json:"endpoint"`
	AccountID string `json:"account_id"`
	SecretKey string `json:"secret_key"`
	Type      string `json:"type"`
}

func (c *CryptoSaaSConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}
