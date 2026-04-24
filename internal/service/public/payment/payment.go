package payment

import (
	"context"
	"net/url"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/payment/v1"
	paymentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/payment"
)

// PaymentService Public Payment服务实现
type PaymentService struct {
	v1.UnimplementedPaymentServer
	uc *paymentBiz.PaymentUseCase
}

// NewPaymentService 创建Public Payment服务
func NewPaymentService(uc *paymentBiz.PaymentUseCase) *PaymentService {
	return &PaymentService{uc: uc}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (s *PaymentService) GetAvailablePaymentMethods(ctx context.Context, req *emptypb.Empty) (*v1.PaymentMethodsReply, error) {
	methods, err := s.uc.GetAvailablePaymentMethods(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*v1.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		list = append(list, &v1.PaymentMethod{
			Id:          m.ID,
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     uint32(m.FeeMode),
			FeePercent:  m.FeePercent,
			FeeAmount:   m.FeeAmount,
		})
	}

	return &v1.PaymentMethodsReply{List: list}, nil
}

// AlipayNotify 支付宝回调
func (s *PaymentService) AlipayNotify(ctx context.Context, req *v1.AlipayNotifyRequest) (*v1.AlipayNotifyReply, error) {
	// 构建url.Values用于签名验证
	params := url.Values{}
	params.Set("trade_no", req.TradeNo)
	params.Set("out_trade_no", req.OutTradeNo)
	params.Set("trade_status", req.TradeStatus)
	params.Set("total_amount", req.TotalAmount)
	params.Set("receipt_amount", req.ReceiptAmount)
	params.Set("buyer_pay_amount", req.BuyerPayAmount)
	params.Set("subject", req.Subject)
	params.Set("body", req.Body)
	params.Set("gmt_create", req.GmtCreate)
	params.Set("gmt_payment", req.GmtPayment)
	params.Set("notify_time", req.NotifyTime)
	params.Set("app_id", req.AppId)
	params.Set("seller_id", req.SellerId)
	params.Set("seller_email", req.SellerEmail)
	params.Set("notify_type", req.NotifyType)
	params.Set("auth_app_id", req.AuthAppId)
	params.Set("charset", req.Charset)
	params.Set("version", req.Version)
	params.Set("sign", req.Sign)

	// 添加额外参数
	for k, v := range req.ExtraParams {
		params.Set(k, v)
	}

	// 调用业务层
	_, response, err := s.uc.AlipayNotify(ctx, req.Token, params)
	if err != nil {
		return nil, err
	}

	return &v1.AlipayNotifyReply{
		Response: response,
	}, nil
}

// EPayNotify 易支付回调
func (s *PaymentService) EPayNotify(ctx context.Context, req *v1.EPayNotifyRequest) (*v1.EPayNotifyReply, error) {
	// 构建参数map
	params := map[string]string{
		"pid":          req.Pid,
		"trade_no":     req.TradeNo,
		"out_trade_no": req.OutTradeNo,
		"type":         req.Type,
		"name":         req.Name,
		"money":        req.Money,
		"trade_status": req.TradeStatus,
		"param":        req.Param,
		"sign":         req.Sign,
	}

	// 添加额外参数
	for k, v := range req.ExtraParams {
		params[k] = v
	}

	// 调用业务层
	_, response, err := s.uc.EPayNotify(ctx, req.Token, params)
	if err != nil {
		return nil, err
	}

	return &v1.EPayNotifyReply{
		Response: response,
	}, nil
}

// StripeNotify Stripe回调
func (s *PaymentService) StripeNotify(ctx context.Context, req *v1.StripeNotifyRequest) (*v1.StripeNotifyReply, error) {
	// 调用业务层
	_, err := s.uc.StripeNotify(ctx, req.Token, req.Payload, req.StripeSignature)
	if err != nil {
		return &v1.StripeNotifyReply{
			Code:    500,
			Message: "Payment notification failed: " + err.Error(),
		}, nil
	}

	return &v1.StripeNotifyReply{
		Code:    0,
		Message: "success",
	}, nil
}
