package payment

// 支付平台定义
type Platform string

const (
	Stripe      Platform = "Stripe"
	AlipayF2F   Platform = "Alipay"
	EPay        Platform = "EPay"
	CryptoSaaS  Platform = "CryptoSaaS"
	Balance     Platform = "Balance"
	UNSUPPORTED Platform = "Unsupported"
)

// ParsePlatform 解析支付平台
func ParsePlatform(platform string) Platform {
	switch platform {
	case "Stripe":
		return Stripe
	case "Alipay":
		return AlipayF2F
	case "EPay":
		return EPay
	case "CryptoSaaS":
		return CryptoSaaS
	case "Balance":
		return Balance
	default:
		return UNSUPPORTED
	}
}

// GetSupportedPlatforms 获取支持的支付平台列表
func GetSupportedPlatforms() []map[string]string {
	return []map[string]string{
		{"name": "Stripe", "label": "Stripe"},
		{"name": "Alipay", "label": "支付宝当面付"},
		{"name": "EPay", "label": "易支付"},
		{"name": "CryptoSaaS", "label": "CryptoSaaS"},
	}
}
