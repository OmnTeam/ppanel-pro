package types

// PaymentRequest 支付请求参数
type PaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	OrderID     string  `json:"order_id"`
	Description string  `json:"description"`
	ReturnURL   string  `json:"return_url"`
	NotifyURL   string  `json:"notify_url"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	Success    bool   `json:"success"`
	PaymentID  string `json:"payment_id"`
	PaymentURL string `json:"payment_url"`
	Message    string `json:"message"`
}

// SMSRequest 短信发送请求
type SMSRequest struct {
	Phone     string            `json:"phone"`
	Message   string            `json:"message"`
	Template  string            `json:"template"`
	Variables map[string]string `json:"variables"`
}

// SMSResponse 短信发送响应
type SMSResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Message   string `json:"message"`
}

// PlatformInfo 支付平台信息
type PlatformInfo struct {
	Platform                 string            `json:"platform"`
	PlatformUrl              string            `json:"platform_url"`
	PlatformFieldDescription map[string]string `json:"platform_field_description"`
}
