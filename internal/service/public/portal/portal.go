package portal

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	portalBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/portal"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/errors"
)

// PortalService Portal服务实现
type PortalService struct {
	v1.UnimplementedPortalServer
	uc *portalBiz.PortalUseCase
}

// NewPortalService 创建Portal服务
func NewPortalService(uc *portalBiz.PortalUseCase) *PortalService {
	return &PortalService{
		uc: uc,
	}
}

// GetSubscription 获取订阅列表（未登录）
func (s *PortalService) GetSubscription(ctx context.Context, req *v1.GetSubscriptionRequest) (*v1.GetSubscriptionReply, error) {
	
	// 获取language参数
	language := ""
	if req.Language != nil {
		language = *req.Language
	}

	list, err := s.uc.GetSubscribeList(ctx, 0, language)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.SubscribeInfo, 0, len(list))
	for _, item := range list {
		// 转换Discount数组
		discounts := make([]*v1.SubscribeDiscount, 0, len(item.Discount))
		for _, d := range item.Discount {
			discounts = append(discounts, &v1.SubscribeDiscount{
				Quantity: d.Quantity,
				Discount: d.Discount,
			})
		}

		items = append(items, &v1.SubscribeInfo{
			Id:             item.ID,
			Name:           item.Name,
			Language:       item.Language,
			Description:    item.Description,
			UnitPrice:      item.UnitPrice,
			UnitTime:       item.UnitTime,
			Discount:       discounts,
			Replacement:    item.Replacement,
			Inventory:      item.Inventory,
			Traffic:        item.Traffic,
			SpeedLimit:     item.SpeedLimit,
			DeviceLimit:    item.DeviceLimit,
			Quota:          item.Quota,
			Nodes:          item.Nodes,
			NodeTags:       item.NodeTags,
			Show:           item.Show,
			Sell:           item.Sell,
			Sort:           item.Sort,
			DeductionRatio: item.DeductionRatio,
			AllowDeduction: item.AllowDeduction,
			ResetCycle:     item.ResetCycle,
			RenewalReset:   item.RenewalReset,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}

	return &v1.GetSubscriptionReply{
		Code:    int32(responsecode.GetSubscriptionSuccess),
		Message: responsecode.CodeMessages[responsecode.GetSubscriptionSuccess],
		Data: &v1.GetSubscriptionData{
			List: items,
		},
	}, nil
}

// PrePurchaseOrder 预购买订单（计算价格）
func (s *PortalService) PrePurchaseOrder(ctx context.Context, req *v1.PrePurchaseOrderRequest) (*v1.PrePurchaseOrderReply, error) {
	
	// 构建coupon和paymentID指针
	var coupon *string
	var paymentID *int64
	if req.Coupon != nil {
		coupon = req.Coupon
	}
	if req.PaymentId != nil {
		paymentID = req.PaymentId
	}

	priceInfo, err := s.uc.PrePurchaseOrder(ctx, 0, req.SubscribeId, req.Quantity, coupon, paymentID)
	if err != nil {
		return nil, err
	}

	return &v1.PrePurchaseOrderReply{
		Code:    int32(responsecode.PrePurchaseOrderSuccess),
		Message: responsecode.CodeMessages[responsecode.PrePurchaseOrderSuccess],
		Data: &v1.PrePurchaseOrderData{
			Price:          priceInfo.Price,
			Amount:         priceInfo.Amount, // 实际支付金额（含手续费）
			Discount:       priceInfo.Discount,
			Coupon:         priceInfo.Coupon,
			CouponDiscount: priceInfo.CouponDiscount,
			FeeAmount:      priceInfo.FeeAmount,
		},
	}, nil
}

// Purchase 购买/创建订单（未登录）
func (s *PortalService) Purchase(ctx context.Context, req *v1.PurchaseRequest) (*v1.PurchaseReply, error) {
	
	// 密码加密
	encryptedPassword := tool.EncodePassWord(req.Password)

	// 构建请求
	var coupon, inviteCode *string
	if req.Coupon != nil {
		coupon = req.Coupon
	}
	if req.InviteCode != nil {
		inviteCode = req.InviteCode
	}

	orderReq := &portalBiz.CreateOrderRequest{
		SubscribeID: req.SubscribeId,
		Quantity:    req.Quantity,
		PaymentID:   req.PaymentId,
		Coupon:      coupon,
		Identifier:  req.Identifier,
		AuthType:    req.AuthType,
		Password:    encryptedPassword, // 已加密
		InviteCode:  inviteCode,
	}

	orderNo, err := s.uc.Purchase(ctx, 0, orderReq)
	if err != nil {
		return nil, err
	}

	return &v1.PurchaseReply{
		Code:    int32(responsecode.PortalPurchaseSuccess),
		Message: responsecode.CodeMessages[responsecode.PortalPurchaseSuccess],
		Data: &v1.PurchaseData{
			OrderNo: orderNo,
		},
	}, nil
}

// GetAvailablePaymentMethods 获取可用支付方式
func (s *PortalService) GetAvailablePaymentMethods(ctx context.Context, req *emptypb.Empty) (*v1.GetAvailablePaymentMethodsReply, error) {
	
	methods, err := s.uc.GetAvailablePaymentMethods(ctx, 0)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		items = append(items, &v1.PaymentMethod{
			Id:          m.ID,
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     m.FeeMode,
			FeePercent:  m.FeePercent,
			FeeAmount:   m.FeeAmount,
		})
	}

	return &v1.GetAvailablePaymentMethodsReply{
		Code:    int32(responsecode.GetAvailablePaymentMethodsSuccess),
		Message: responsecode.CodeMessages[responsecode.GetAvailablePaymentMethodsSuccess],
		Data: &v1.GetAvailablePaymentMethodsData{
			Methods: items,
		},
	}, nil
}

// PurchaseCheckout 购买结账（获取支付信息）
func (s *PortalService) PurchaseCheckout(ctx context.Context, req *v1.PurchaseCheckoutRequest) (*v1.PurchaseCheckoutReply, error) {
	
	if req.OrderNo == "" {
		return nil, errors.BadRequest("INVALID_PARAMETER", "订单号不能为空")
	}

	// ReturnURL: 支付回调地址（可选）
	returnURL := ""
	if req.ReturnUrl != nil {
		returnURL = *req.ReturnUrl
	}

	paymentInfo, err := s.uc.PurchaseCheckout(ctx, 0, req.OrderNo, returnURL)
	if err != nil {
		return nil, err
	}

	// 构建返回结构（复刻原项目 purchaseCheckoutLogic.go 返回结构）
	data := &v1.PurchaseCheckoutData{
		Type:        paymentInfo.Type,
		CheckoutUrl: nil,
		Stripe:      nil,
	}

	// 根据类型设置相应字段
	if paymentInfo.CheckoutURL != "" {
		data.CheckoutUrl = &paymentInfo.CheckoutURL
	}

	if paymentInfo.Stripe != nil {
		data.Stripe = &v1.StripePayment{
			PublishableKey: paymentInfo.Stripe.PublishableKey,
			ClientSecret:   paymentInfo.Stripe.ClientSecret,
			Method:         paymentInfo.Stripe.Method,
		}
	}

	return &v1.PurchaseCheckoutReply{
		Code:    int32(responsecode.PurchaseCheckoutSuccess),
		Message: responsecode.CodeMessages[responsecode.PurchaseCheckoutSuccess],
		Data:    data,
	}, nil
}

// QueryPurchaseOrder 查询购买订单状态
func (s *PortalService) QueryPurchaseOrder(ctx context.Context, req *v1.QueryPurchaseOrderRequest) (*v1.QueryPurchaseOrderReply, error) {
	
	if req.OrderNo == "" {
		return nil, errors.BadRequest("INVALID_PARAMETER", "订单号不能为空")
	}
	if req.AuthType == "" {
		return nil, errors.BadRequest("INVALID_PARAMETER", "认证类型不能为空")
	}
	if req.Identifier == "" {
		return nil, errors.BadRequest("INVALID_PARAMETER", "认证标识符不能为空")
	}

	statusInfo, token, err := s.uc.QueryPurchaseOrder(ctx, 0, req.OrderNo, req.AuthType, req.Identifier)
	if err != nil {
		return nil, err
	}

	// 构建Subscribe对象
	var subscribeInfo *v1.SubscribeInfo
	if statusInfo.Subscribe != nil {
		// 转换Discount数组
		discounts := make([]*v1.SubscribeDiscount, 0, len(statusInfo.Subscribe.Discount))
		for _, d := range statusInfo.Subscribe.Discount {
			discounts = append(discounts, &v1.SubscribeDiscount{
				Quantity: d.Quantity,
				Discount: d.Discount,
			})
		}

		subscribeInfo = &v1.SubscribeInfo{
			Id:             statusInfo.Subscribe.ID,
			Name:           statusInfo.Subscribe.Name,
			Language:       statusInfo.Subscribe.Language,
			Description:    statusInfo.Subscribe.Description,
			UnitPrice:      statusInfo.Subscribe.UnitPrice,
			UnitTime:       statusInfo.Subscribe.UnitTime,
			Discount:       discounts,
			Replacement:    statusInfo.Subscribe.Replacement,
			Inventory:      statusInfo.Subscribe.Inventory,
			Traffic:        statusInfo.Subscribe.Traffic,
			SpeedLimit:     statusInfo.Subscribe.SpeedLimit,
			DeviceLimit:    statusInfo.Subscribe.DeviceLimit,
			Quota:          statusInfo.Subscribe.Quota,
			Nodes:          statusInfo.Subscribe.Nodes,
			NodeTags:       statusInfo.Subscribe.NodeTags,
			Show:           statusInfo.Subscribe.Show,
			Sell:           statusInfo.Subscribe.Sell,
			Sort:           statusInfo.Subscribe.Sort,
			DeductionRatio: statusInfo.Subscribe.DeductionRatio,
			AllowDeduction: statusInfo.Subscribe.AllowDeduction,
			ResetCycle:     statusInfo.Subscribe.ResetCycle,
			RenewalReset:   statusInfo.Subscribe.RenewalReset,
			CreatedAt:      statusInfo.Subscribe.CreatedAt,
			UpdatedAt:      statusInfo.Subscribe.UpdatedAt,
		}
	}

	// 构建Payment对象
	var paymentInfo *v1.PaymentMethod
	if statusInfo.Payment != nil {
		paymentInfo = &v1.PaymentMethod{
			Id:          statusInfo.Payment.ID,
			Name:        statusInfo.Payment.Name,
			Platform:    statusInfo.Payment.Platform,
			Description: statusInfo.Payment.Description,
			Icon:        statusInfo.Payment.Icon,
			FeeMode:     statusInfo.Payment.FeeMode,
			FeePercent:  statusInfo.Payment.FeePercent,
			FeeAmount:   statusInfo.Payment.FeeAmount,
		}
	}

	return &v1.QueryPurchaseOrderReply{
		Code:    int32(responsecode.QueryPurchaseOrderSuccess),
		Message: responsecode.CodeMessages[responsecode.QueryPurchaseOrderSuccess],
		Data: &v1.QueryPurchaseOrderData{
			OrderNo:        statusInfo.OrderNo,
			Subscribe:      subscribeInfo,
			Quantity:       statusInfo.Quantity,
			Price:          statusInfo.Price,
			Amount:         statusInfo.Amount,
			Discount:       statusInfo.Discount,
			Coupon:         statusInfo.Coupon,
			CouponDiscount: statusInfo.CouponDiscount,
			FeeAmount:      statusInfo.FeeAmount,
			Payment:        paymentInfo,
			Status:         statusInfo.Status,
			CreatedAt:      statusInfo.CreatedAt.UnixMilli(),
			Token:          token,
		},
	}, nil
}
