package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyorder"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	paymentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// PaymentConfigJSON 支付配置JSON结构
type PaymentConfigJSON struct {
	AppID         string `json:"app_id"`
	PrivateKey    string `json:"private_key"`
	PublicKey     string `json:"public_key"`
	WebhookSecret string `json:"webhook_secret"`
	InvoiceName   string `json:"invoice_name"`
	Sandbox       bool   `json:"sandbox"`
	EPayPid       string `json:"epay_pid"`
	EPayKey       string `json:"epay_key"`
	EPayURL       string `json:"epay_url"`
	NotifyURL     string `json:"notify_url"`
}

type publicPaymentRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicPaymentRepo 创建Public Payment仓库
func NewPublicPaymentRepo(data *Data, logger log.Logger) paymentBiz.PaymentRepo {
	return &publicPaymentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (r *publicPaymentRepo) GetAvailablePaymentMethods(ctx context.Context) ([]*paymentBiz.PaymentMethod, error) {
	// 查询enable=true的支付方式
	methods, err := r.data.db.ProxyPayment.Query().
		Where(
			proxypayment.Enable(true),
		).
		Order(ent.Asc(proxypayment.FieldID)).
		All(ctx)

	if err != nil {
		r.log.Errorf("GetAvailablePaymentMethods query error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	result := make([]*paymentBiz.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		result = append(result, &paymentBiz.PaymentMethod{
			ID:          int64(m.ID),
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     int32(m.FeeMode),
			FeePercent:  int64(m.FeePercent),
			FeeAmount:   int64(m.FeeAmount),
		})
	}

	return result, nil
}

// GetPaymentConfigByToken 根据token获取支付配置
func (r *publicPaymentRepo) GetPaymentConfigByToken(ctx context.Context, token string) (*paymentBiz.PaymentConfig, error) {
	// 根据token查询支付配置
	payment, err := r.data.db.ProxyPayment.Query().
		Where(
			proxypayment.Token(token),
		).
		Only(ctx)

	if err != nil {
		r.log.Errorf("GetPaymentConfigByToken query error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrPaymentNotFound)
	}

	// 解析config JSON字段
	var configJSON PaymentConfigJSON
	if payment.Config != "" {
		err = json.Unmarshal([]byte(payment.Config), &configJSON)
		if err != nil {
			r.log.Errorf("GetPaymentConfigByToken parse config error: %v", err)
			// 返回空配置而不是错误，保持向后兼容
		}
	}

	return &paymentBiz.PaymentConfig{
		ID:            int64(payment.ID),
		Platform:      payment.Platform,
		AppID:         configJSON.AppID,
		PrivateKey:    configJSON.PrivateKey,
		PublicKey:     configJSON.PublicKey,
		WebhookSecret: configJSON.WebhookSecret,
		InvoiceName:   configJSON.InvoiceName,
		Sandbox:       configJSON.Sandbox,
		EPayPid:       configJSON.EPayPid,
		EPayKey:       configJSON.EPayKey,
		EPayURL:       configJSON.EPayURL,
		NotifyURL:     configJSON.NotifyURL,
	}, nil
}

// ActivateOrder 激活订单（支付成功后调用）
func (r *publicPaymentRepo) ActivateOrder(ctx context.Context, orderNo string, platform string, tradeNo string, amount int64) error {
	// 查询订单
	order, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.OrderNo(orderNo),
		).
		Only(ctx)

	if err != nil {
		r.log.Errorf("ActivateOrder query order error: %v", err)
		return fmt.Errorf("order not found: %s", orderNo)
	}

	// 检查订单状态，避免重复激活
	if order.Status == 1 { // 已支付
		r.log.Infof("Order already paid: %s", orderNo)
		return nil
	}

	// TODO: 验证金额是否匹配（这里需要订单金额信息）

	// 更新订单状态
	err = r.data.db.ProxyOrder.UpdateOneID(order.ID).
		SetStatus(1). // 已支付
		SetTradeNo(tradeNo).
		SetMethod(platform).
		Exec(ctx)

	if err != nil {
		r.log.Errorf("ActivateOrder update order error: %v", err)
		return fmt.Errorf("failed to update order: %v", err)
	}

	// TODO: 激活订阅或充值到余额
	// 这里需要根据订单类型进行处理
	// 1. 如果是新购订阅：激活订阅
	// 2. 如果是续费：延期订阅
	// 3. 如果是充值：增加余额
	// 4. 如果是流量重置：重置流量

	r.log.Infof("Order activated successfully: orderNo=%s, tradeNo=%s, platform=%s",
		orderNo, tradeNo, platform)

	return nil
}
