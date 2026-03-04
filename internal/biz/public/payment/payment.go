package payment

import (
	"context"
	"net/url"

	"github.com/go-kratos/kratos/v2/log"
)

// PaymentRepo Public Payment数据仓库接口
type PaymentRepo interface {
	// GetAvailablePaymentMethods 获取可用支付方式
	GetAvailablePaymentMethods(ctx context.Context) ([]*PaymentMethod, error)
	// GetPaymentConfigByToken 根据token获取支付配置
	GetPaymentConfigByToken(ctx context.Context, token string) (*PaymentConfig, error)
	// ActivateOrder 激活订单（支付成功后调用）
	ActivateOrder(ctx context.Context, orderNo string, platform string, tradeNo string, amount int64) error
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID          int64
	Name        string
	Platform    string
	Description string
	Icon        string
	FeeMode     int32
	FeePercent  int64
	FeeAmount   int64
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	ID            int64
	Platform      string
	AppID         string
	PrivateKey    string
	PublicKey     string
	WebhookSecret string
	InvoiceName   string
	Sandbox       bool
	EPayPid       string
	EPayKey       string
	EPayURL       string
	NotifyURL     string
}

// PaymentUseCase Public Payment用例
type PaymentUseCase struct {
	repo PaymentRepo
	log  *log.Helper
}

// NewPaymentUseCase 创建Public Payment用例
func NewPaymentUseCase(repo PaymentRepo, logger log.Logger) *PaymentUseCase {
	return &PaymentUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (uc *PaymentUseCase) GetAvailablePaymentMethods(ctx context.Context) ([]*PaymentMethod, error) {
	return uc.repo.GetAvailablePaymentMethods(ctx)
}

// AlipayNotify 处理支付宝回调
func (uc *PaymentUseCase) AlipayNotify(ctx context.Context, token string, params url.Values) (bool, string, error) {
	// 获取支付配置
	config, err := uc.repo.GetPaymentConfigByToken(ctx, token)
	if err != nil {
		uc.log.Errorf("GetPaymentConfigByToken failed: %v", err)
		return false, "failure", err
	}

	// 验证平台是否匹配
	if config.Platform != "alipay" {
		uc.log.Errorf("Platform mismatch: expected alipay, got %s", config.Platform)
		return false, "failure", err
	}

	// TODO: 使用pkg/payment/alipay验证签名
	// 这里需要调用 alipay client 的 DecodeNotification 方法验证签名

	// 提取订单信息
	orderNo := params.Get("out_trade_no")
	tradeNo := params.Get("trade_no")
	totalAmount := params.Get("total_amount")
	tradeStatus := params.Get("trade_status")

	uc.log.Infof("Alipay notify: orderNo=%s, tradeNo=%s, amount=%s, status=%s",
		orderNo, tradeNo, totalAmount, tradeStatus)

	// 验证交易状态
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		uc.log.Warnf("Alipay trade status not success: %s", tradeStatus)
		return false, "success", nil // 返回success避免重复通知
	}

	// 转换金额（分）
	// amount := int64(float64(totalAmount) * 100)

	// 激活订单
	err = uc.repo.ActivateOrder(ctx, orderNo, "alipay", tradeNo, 0)
	if err != nil {
		uc.log.Errorf("ActivateOrder failed: %v", err)
		return false, "failure", err
	}

	return true, "success", nil
}

// EPayNotify 处理易支付回调
func (uc *PaymentUseCase) EPayNotify(ctx context.Context, token string, params map[string]string) (bool, string, error) {
	// 获取支付配置
	config, err := uc.repo.GetPaymentConfigByToken(ctx, token)
	if err != nil {
		uc.log.Errorf("GetPaymentConfigByToken failed: %v", err)
		return false, "failure", err
	}

	// 验证平台是否匹配
	if config.Platform != "epay" {
		uc.log.Errorf("Platform mismatch: expected epay, got %s", config.Platform)
		return false, "failure", err
	}

	// TODO: 使用pkg/payment/epay验证签名
	// 这里需要调用 epay client 的 VerifySign 方法验证签名

	// 提取订单信息
	orderNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	money := params["money"]
	tradeStatus := params["trade_status"]

	uc.log.Infof("EPay notify: orderNo=%s, tradeNo=%s, amount=%s, status=%s",
		orderNo, tradeNo, money, tradeStatus)

	// 验证交易状态
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "trade_success" {
		uc.log.Warnf("EPay trade status not success: %s", tradeStatus)
		return false, "success", nil // 返回success避免重复通知
	}

	// 激活订单
	err = uc.repo.ActivateOrder(ctx, orderNo, "epay", tradeNo, 0)
	if err != nil {
		uc.log.Errorf("ActivateOrder failed: %v", err)
		return false, "failure", err
	}

	return true, "success", nil
}

// StripeNotify 处理Stripe回调
func (uc *PaymentUseCase) StripeNotify(ctx context.Context, token string, payload []byte, signature string) (bool, error) {
	// 获取支付配置
	config, err := uc.repo.GetPaymentConfigByToken(ctx, token)
	if err != nil {
		uc.log.Errorf("GetPaymentConfigByToken failed: %v", err)
		return false, err
	}

	// 验证平台是否匹配
	if config.Platform != "stripe" {
		uc.log.Errorf("Platform mismatch: expected stripe, got %s", config.Platform)
		return false, err
	}

	// TODO: 使用pkg/payment/stripe验证签名
	// 这里需要调用 stripe client 的 ParseNotify 方法验证签名

	// 临时实现：从payload解析事件类型
	// 实际应该使用 stripe.ParseNotify

	uc.log.Infof("Stripe notify: token=%s, signature=%s", token, signature)

	// TODO: 解析订单信息和金额
	// 然后调用 ActivateOrder

	// err = uc.repo.ActivateOrder(ctx, orderNo, "stripe", tradeNo, amount)
	// if err != nil {
	// 	uc.log.Errorf("ActivateOrder failed: %v", err)
	// 	return false, err
	// }

	return true, nil
}
