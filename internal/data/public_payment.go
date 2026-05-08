package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyorder"
	"github.com/OmnTeam/npanel-pro/ent/proxypayment"
	paymentBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/payment"
	queueTypes "github.com/OmnTeam/npanel-pro/internal/queue/types"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
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

	config := &paymentBiz.PaymentConfig{
		ID:       int64(payment.ID),
		Platform: normalizePaymentPlatform(payment.Platform),
	}

	switch config.Platform {
	case "stripe":
		var stripeCfg StripeConfig
		if payment.Config != "" {
			if err := stripeCfg.Unmarshal([]byte(payment.Config)); err != nil {
				r.log.Errorf("GetPaymentConfigByToken parse stripe config error: %v", err)
			}
		}
		config.PublicKey = stripeCfg.PublicKey
		config.PrivateKey = stripeCfg.SecretKey
		config.WebhookSecret = stripeCfg.WebhookSecret

	case "alipay":
		var alipayCfg AlipayF2FConfig
		if payment.Config != "" {
			if err := alipayCfg.Unmarshal([]byte(payment.Config)); err != nil {
				r.log.Errorf("GetPaymentConfigByToken parse alipay config error: %v", err)
			}
		}
		config.AppID = alipayCfg.AppId
		config.PrivateKey = alipayCfg.PrivateKey
		config.PublicKey = alipayCfg.PublicKey
		config.InvoiceName = alipayCfg.InvoiceName
		config.Sandbox = alipayCfg.Sandbox

	case "epay":
		if isCryptoSaaSPayment(payment.Platform) {
			var cryptoCfg CryptoSaaSConfig
			if payment.Config != "" {
				if err := cryptoCfg.Unmarshal([]byte(payment.Config)); err != nil {
					r.log.Errorf("GetPaymentConfigByToken parse cryptosaas config error: %v", err)
				}
			}
			config.EPayPid = cryptoCfg.AccountID
			config.EPayKey = cryptoCfg.SecretKey
			config.EPayURL = cryptoCfg.Endpoint
			break
		}

		var epayCfg EPayConfig
		if payment.Config != "" {
			if err := epayCfg.Unmarshal([]byte(payment.Config)); err != nil {
				r.log.Errorf("GetPaymentConfigByToken parse epay config error: %v", err)
			}
		}
		config.EPayPid = epayCfg.Pid
		config.EPayKey = epayCfg.Key
		config.EPayURL = epayCfg.Url

	default:
		var configJSON PaymentConfigJSON
		if payment.Config != "" {
			if err := json.Unmarshal([]byte(payment.Config), &configJSON); err != nil {
				r.log.Errorf("GetPaymentConfigByToken parse fallback config error: %v", err)
			}
		}
		config.Platform = normalizePaymentPlatform(config.Platform)
		config.AppID = configJSON.AppID
		config.PrivateKey = configJSON.PrivateKey
		config.PublicKey = configJSON.PublicKey
		config.WebhookSecret = configJSON.WebhookSecret
		config.InvoiceName = configJSON.InvoiceName
		config.Sandbox = configJSON.Sandbox
		config.EPayPid = configJSON.EPayPid
		config.EPayKey = configJSON.EPayKey
		config.EPayURL = configJSON.EPayURL
		config.NotifyURL = configJSON.NotifyURL
	}

	return config, nil
}

// ActivateOrder 激活订单（支付成功后调用）
func (r *publicPaymentRepo) ActivateOrder(ctx context.Context, orderNo string, platform string, tradeNo string, amount int64) error {
	order, err := r.data.db.ProxyOrder.Query().
		Where(
			proxyorder.OrderNo(orderNo),
		).
		Only(ctx)

	if err != nil {
		r.log.Errorf("ActivateOrder query order error: %v", err)
		return fmt.Errorf("order not found: %s", orderNo)
	}

	if order.Status == 5 {
		r.log.Infof("Order already finished: %s", orderNo)
		return nil
	}

	if order.Status != 2 {
		err = r.data.db.ProxyOrder.UpdateOneID(order.ID).
			SetStatus(2).
			Exec(ctx)
		if err != nil {
			r.log.Errorf("ActivateOrder update order error: %v", err)
			return fmt.Errorf("failed to update order: %v", err)
		}
	}

	payloadBytes, err := json.Marshal(queueTypes.ForthwithActivateOrderPayload{OrderNo: orderNo})
	if err != nil {
		return err
	}

	task := asynq.NewTask(queueTypes.ForthwithActivateOrder, payloadBytes, asynq.MaxRetry(5))
	if _, err := r.data.queue.Enqueue(task); err != nil {
		r.log.Errorf("ActivateOrder enqueue task error: %v", err)
		return err
	}

	return nil
}

func normalizePaymentPlatform(platform string) string {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")

	switch normalized {
	case "stripe":
		return "stripe"
	case "alipay", "alipayf2f":
		return "alipay"
	case "epay", "cryptosaas":
		return "epay"
	default:
		return normalized
	}
}

func isCryptoSaaSPayment(platform string) bool {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return normalized == "cryptosaas"
}
