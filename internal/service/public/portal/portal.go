package portal

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	portalBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/portal"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func convertPortalTrafficLimits(items []portalBiz.TrafficLimit) []*v1.TrafficLimit {
	result := make([]*v1.TrafficLimit, 0, len(items))
	for _, item := range items {
		result = append(result, &v1.TrafficLimit{
			StatType:     item.StatType,
			StatValue:    strconv.FormatInt(item.StatValue, 10),
			TrafficUsage: strconv.FormatInt(item.TrafficUsage, 10),
			SpeedLimit:   strconv.FormatInt(item.SpeedLimit, 10),
		})
	}
	return result
}

func convertPortalSubscribe(item *portalBiz.SubscribeInfo) *v1.SubscribeInfo {
	if item == nil {
		return nil
	}
	discounts := make([]*v1.SubscribeDiscount, 0, len(item.Discount))
	for _, d := range item.Discount {
		discounts = append(discounts, &v1.SubscribeDiscount{
			Quantity: strconv.FormatInt(int64(d.Quantity), 10),
			Discount: strconv.FormatInt(int64(d.Discount), 10),
		})
	}
	return &v1.SubscribeInfo{
		Id:                strconv.FormatInt(item.ID, 10),
		Name:              item.Name,
		Language:          item.Language,
		Description:       item.Description,
		UnitPrice:         strconv.FormatInt(item.UnitPrice, 10),
		UnitTime:          item.UnitTime,
		Discount:          discounts,
		Replacement:       strconv.FormatInt(item.Replacement, 10),
		Inventory:         strconv.FormatInt(item.Inventory, 10),
		Traffic:           strconv.FormatInt(item.Traffic, 10),
		SpeedLimit:        strconv.FormatInt(item.SpeedLimit, 10),
		DeviceLimit:       strconv.FormatInt(item.DeviceLimit, 10),
		Quota:             strconv.FormatInt(item.Quota, 10),
		Nodes:             convertIntSliceToStringSlice(item.Nodes),
		NodeTags:          item.NodeTags,
		NodeGroupIds:      item.NodeGroupIds,
		NodeGroupId:       item.NodeGroupId,
		TrafficLimit:      convertPortalTrafficLimits(item.TrafficLimit),
		Show:              item.Show,
		Sell:              item.Sell,
		Sort:              strconv.FormatInt(item.Sort, 10),
		DeductionRatio:    strconv.FormatInt(item.DeductionRatio, 10),
		AllowDeduction:    item.AllowDeduction,
		ResetCycle:        strconv.FormatInt(item.ResetCycle, 10),
		RenewalReset:      item.RenewalReset,
		ShowOriginalPrice: item.ShowOriginalPrice,
		CreatedAt:         strconv.FormatInt(item.CreatedAt, 10),
		UpdatedAt:         strconv.FormatInt(item.UpdatedAt, 10),
	}
}

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

	list, err := s.uc.GetSubscribeList(ctx, language)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.SubscribeInfo, 0, len(list))
	for _, item := range list {
		items = append(items, convertPortalSubscribe(item))
	}

	return &v1.GetSubscriptionReply{
		Code:    int32(responsecode.GetSubscriptionSuccess),
		Message: responsecode.CodeMessages[responsecode.GetSubscriptionSuccess],
		Data: &v1.GetSubscriptionData{
			List: items,
		},
	}, nil
}

// convertIntSliceToStringSlice converts []int to []string
func convertIntSliceToStringSlice(input []int) []string {
	if input == nil {
		return nil
	}
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = strconv.FormatInt(int64(v), 10)
	}
	return result
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
		parsedID := parseInt64(*req.PaymentId)
		paymentID = &parsedID
	}

	priceInfo, err := s.uc.PrePurchaseOrder(ctx, parseInt64(req.SubscribeId), parseInt64(req.Quantity), coupon, paymentID)
	if err != nil {
		return nil, err
	}

	return &v1.PrePurchaseOrderReply{
		Code:    int32(responsecode.PrePurchaseOrderSuccess),
		Message: responsecode.CodeMessages[responsecode.PrePurchaseOrderSuccess],
		Data: &v1.PrePurchaseOrderData{
			Price:          strconv.FormatInt(int64(priceInfo.Price), 10),
			Amount:         strconv.FormatInt(int64(priceInfo.Amount), 10), // 实际支付金额（含手续费）
			Discount:       strconv.FormatInt(int64(priceInfo.Discount), 10),
			Coupon:         priceInfo.Coupon,
			CouponDiscount: strconv.FormatInt(int64(priceInfo.CouponDiscount), 10),
			FeeAmount:      strconv.FormatInt(int64(priceInfo.FeeAmount), 10),
		},
	}, nil
}

// Purchase 购买/创建订单（未登录）
func (s *PortalService) Purchase(ctx context.Context, req *v1.PurchaseRequest) (*v1.PurchaseReply, error) {

	// 构建请求
	var coupon, inviteCode *string
	if req.Coupon != nil {
		coupon = req.Coupon
	}
	if req.InviteCode != nil {
		inviteCode = req.InviteCode
	}

	orderReq := &portalBiz.CreateOrderRequest{
		SubscribeID: parseInt64(req.SubscribeId),
		Quantity:    parseInt64(req.Quantity),
		PaymentID:   int(parseInt64(req.PaymentId)),
		Coupon:      coupon,
		Identifier:  req.Identifier,
		AuthType:    req.AuthType,
		Password:    req.Password,
		InviteCode:  inviteCode,
	}

	orderNo, err := s.uc.Purchase(ctx, orderReq)
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

	methods, err := s.uc.GetAvailablePaymentMethods(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		items = append(items, &v1.PaymentMethod{
			Id:          strconv.FormatInt(m.ID, 10),
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     m.FeeMode,
			FeePercent:  strconv.FormatInt(int64(m.FeePercent), 10),
			FeeAmount:   strconv.FormatInt(int64(m.FeeAmount), 10),
		})
	}

	return &v1.GetAvailablePaymentMethodsReply{
		Code:    int32(responsecode.GetAvailablePaymentMethodsSuccess),
		Message: responsecode.CodeMessages[responsecode.GetAvailablePaymentMethodsSuccess],
		Data: &v1.GetAvailablePaymentMethodsData{
			List: items,
		},
	}, nil
}

// PurchaseCheckout 购买结账（获取支付信息）
func (s *PortalService) PurchaseCheckout(ctx context.Context, req *v1.PurchaseCheckoutRequest) (*v1.PurchaseCheckoutReply, error) {
	// ReturnURL: 支付回调地址（可选）
	returnURL := ""
	if req.ReturnUrl != nil {
		returnURL = *req.ReturnUrl
	}

	paymentInfo, err := s.uc.PurchaseCheckout(ctx, req.OrderNo, returnURL)
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
	statusInfo, token, err := s.uc.QueryPurchaseOrder(ctx, req.OrderNo, req.AuthType, req.Identifier)
	if err != nil {
		return nil, err
	}

	// 构建Subscribe对象
	var subscribeInfo *v1.SubscribeInfo
	if statusInfo.Subscribe != nil {
		subscribeInfo = convertPortalSubscribe(statusInfo.Subscribe)
	}

	// 构建Payment对象
	var paymentInfo *v1.PaymentMethod
	if statusInfo.Payment != nil {
		paymentInfo = &v1.PaymentMethod{
			Id:          strconv.FormatInt(statusInfo.Payment.ID, 10),
			Name:        statusInfo.Payment.Name,
			Platform:    statusInfo.Payment.Platform,
			Description: statusInfo.Payment.Description,
			Icon:        statusInfo.Payment.Icon,
			FeeMode:     statusInfo.Payment.FeeMode,
			FeePercent:  strconv.FormatInt(int64(statusInfo.Payment.FeePercent), 10),
			FeeAmount:   strconv.FormatInt(int64(statusInfo.Payment.FeeAmount), 10),
		}
	}

	return &v1.QueryPurchaseOrderReply{
		Code:    int32(responsecode.QueryPurchaseOrderSuccess),
		Message: responsecode.CodeMessages[responsecode.QueryPurchaseOrderSuccess],
		Data: &v1.QueryPurchaseOrderData{
			OrderNo:        statusInfo.OrderNo,
			Subscribe:      subscribeInfo,
			Quantity:       strconv.FormatInt(statusInfo.Quantity, 10),
			Price:          strconv.FormatInt(statusInfo.Price, 10),
			Amount:         strconv.FormatInt(statusInfo.Amount, 10),
			Discount:       strconv.FormatInt(statusInfo.Discount, 10),
			Coupon:         statusInfo.Coupon,
			CouponDiscount: strconv.FormatInt(statusInfo.CouponDiscount, 10),
			FeeAmount:      strconv.FormatInt(statusInfo.FeeAmount, 10),
			Payment:        paymentInfo,
			Status:         statusInfo.Status,
			CreatedAt:      strconv.FormatInt(statusInfo.CreatedAt.UnixMilli(), 10),
			Token:          token,
		},
	}, nil
}

// convertIntSliceToInt64Slice converts []int to []int64
func convertIntSliceToInt64Slice(input []int) []int64 {
	if input == nil {
		return nil
	}
	result := make([]int64, len(input))
	for i, v := range input {
		result[i] = int64(v)
	}
	return result
}
