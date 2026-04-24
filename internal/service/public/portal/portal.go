package portal

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	portalBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/portal"
)

func convertPortalTrafficLimits(items []portalBiz.TrafficLimit) []*v1.TrafficLimit {
	result := make([]*v1.TrafficLimit, 0, len(items))
	for _, item := range items {
		result = append(result, &v1.TrafficLimit{
			StatType:     item.StatType,
			StatValue:    item.StatValue,
			TrafficUsage: item.TrafficUsage,
			SpeedLimit:   int32(item.SpeedLimit),
		})
	}
	return result
}

func convertPortalSubscribe(item *portalBiz.SubscribeInfo) *v1.SubscribeInfo {
	if item == nil {
		return nil
	}

	description := ""
	if item.Description != nil {
		description = *item.Description
	}

	discounts := make([]*v1.SubscribeDiscount, 0, len(item.Discount))
	for _, discount := range item.Discount {
		discounts = append(discounts, &v1.SubscribeDiscount{
			Quantity: int64(discount.Quantity),
			Discount: int64(discount.Discount),
		})
	}

	return &v1.SubscribeInfo{
		Id:                item.ID,
		Name:              item.Name,
		Language:          item.Language,
		Description:       description,
		UnitPrice:         item.UnitPrice,
		UnitTime:          item.UnitTime,
		Discount:          discounts,
		Replacement:       item.Replacement,
		Inventory:         int32(item.Inventory),
		Traffic:           item.Traffic,
		SpeedLimit:        int32(item.SpeedLimit),
		DeviceLimit:       int32(item.DeviceLimit),
		Quota:             int32(item.Quota),
		Nodes:             convertIntSliceToInt64Slice(item.Nodes),
		NodeTags:          item.NodeTags,
		NodeGroupIds:      convertStringSliceToInt64Slice(item.NodeGroupIds),
		NodeGroupId:       parseStringInt64(item.NodeGroupId),
		TrafficLimit:      convertPortalTrafficLimits(item.TrafficLimit),
		Show:              item.Show,
		Sell:              item.Sell,
		Sort:              int32(item.Sort),
		DeductionRatio:    int32(item.DeductionRatio),
		AllowDeduction:    item.AllowDeduction,
		ResetCycle:        int32(item.ResetCycle),
		RenewalReset:      item.RenewalReset,
		ShowOriginalPrice: item.ShowOriginalPrice,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

type PortalService struct {
	v1.UnimplementedPortalServer
	uc *portalBiz.PortalUseCase
}

func NewPortalService(uc *portalBiz.PortalUseCase) *PortalService {
	return &PortalService{uc: uc}
}

func (s *PortalService) GetSubscription(ctx context.Context, req *v1.GetSubscriptionRequest) (*v1.GetSubscriptionReply, error) {
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

	return &v1.GetSubscriptionReply{List: items}, nil
}

func (s *PortalService) PrePurchaseOrder(ctx context.Context, req *v1.PrePurchaseOrderRequest) (*v1.PrePurchaseOrderReply, error) {
	var coupon *string
	if req.Coupon != nil {
		coupon = req.Coupon
	}

	var paymentID *int64
	if req.Payment > 0 {
		paymentID = &req.Payment
	}

	priceInfo, err := s.uc.PrePurchaseOrder(ctx, req.SubscribeId, req.Quantity, coupon, paymentID)
	if err != nil {
		return nil, err
	}

	return &v1.PrePurchaseOrderReply{
		Price:          int64(priceInfo.Price),
		Amount:         int64(priceInfo.Amount),
		Discount:       int64(priceInfo.Discount),
		Coupon:         priceInfo.Coupon,
		CouponDiscount: int64(priceInfo.CouponDiscount),
		FeeAmount:      int64(priceInfo.FeeAmount),
	}, nil
}

func (s *PortalService) Purchase(ctx context.Context, req *v1.PurchaseRequest) (*v1.PurchaseReply, error) {
	var coupon *string
	if req.Coupon != nil {
		coupon = req.Coupon
	}

	var inviteCode *string
	if req.InviteCode != nil {
		inviteCode = req.InviteCode
	}

	orderNo, err := s.uc.Purchase(ctx, &portalBiz.CreateOrderRequest{
		SubscribeID: req.SubscribeId,
		Quantity:    req.Quantity,
		PaymentID:   int(req.Payment),
		Coupon:      coupon,
		Identifier:  req.Identifier,
		AuthType:    req.AuthType,
		Password:    req.Password,
		InviteCode:  inviteCode,
	})
	if err != nil {
		return nil, err
	}

	return &v1.PurchaseReply{OrderNo: orderNo}, nil
}

func (s *PortalService) GetAvailablePaymentMethods(ctx context.Context, req *emptypb.Empty) (*v1.GetAvailablePaymentMethodsReply, error) {
	methods, err := s.uc.GetAvailablePaymentMethods(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PaymentMethod, 0, len(methods))
	for _, method := range methods {
		items = append(items, &v1.PaymentMethod{
			Id:          method.ID,
			Name:        method.Name,
			Platform:    method.Platform,
			Description: method.Description,
			Icon:        method.Icon,
			FeeMode:     uint32(method.FeeMode),
			FeePercent:  int64(method.FeePercent),
			FeeAmount:   int64(method.FeeAmount),
		})
	}

	return &v1.GetAvailablePaymentMethodsReply{Methods: items}, nil
}

func (s *PortalService) PurchaseCheckout(ctx context.Context, req *v1.PurchaseCheckoutRequest) (*v1.PurchaseCheckoutReply, error) {
	returnURL := ""
	if req.ReturnUrl != nil {
		returnURL = *req.ReturnUrl
	}

	paymentInfo, err := s.uc.PurchaseCheckout(ctx, req.OrderNo, returnURL)
	if err != nil {
		return nil, err
	}

	reply := &v1.PurchaseCheckoutReply{Type: paymentInfo.Type}
	if paymentInfo.CheckoutURL != "" {
		reply.CheckoutUrl = &paymentInfo.CheckoutURL
	}
	if paymentInfo.Stripe != nil {
		reply.Stripe = &v1.StripePayment{
			PublishableKey: paymentInfo.Stripe.PublishableKey,
			ClientSecret:   paymentInfo.Stripe.ClientSecret,
			Method:         paymentInfo.Stripe.Method,
		}
	}

	return reply, nil
}

func (s *PortalService) QueryPurchaseOrder(ctx context.Context, req *v1.QueryPurchaseOrderRequest) (*v1.QueryPurchaseOrderReply, error) {
	statusInfo, token, err := s.uc.QueryPurchaseOrder(ctx, req.OrderNo, req.AuthType, req.Identifier)
	if err != nil {
		return nil, err
	}

	var paymentInfo *v1.PaymentMethod
	if statusInfo.Payment != nil {
		paymentInfo = &v1.PaymentMethod{
			Id:          statusInfo.Payment.ID,
			Name:        statusInfo.Payment.Name,
			Platform:    statusInfo.Payment.Platform,
			Description: statusInfo.Payment.Description,
			Icon:        statusInfo.Payment.Icon,
			FeeMode:     uint32(statusInfo.Payment.FeeMode),
			FeePercent:  int64(statusInfo.Payment.FeePercent),
			FeeAmount:   int64(statusInfo.Payment.FeeAmount),
		}
	}

	return &v1.QueryPurchaseOrderReply{
		OrderNo:        statusInfo.OrderNo,
		Subscribe:      convertPortalSubscribe(statusInfo.Subscribe),
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
	}, nil
}

func convertIntSliceToInt64Slice(input []int) []int64 {
	if len(input) == 0 {
		return []int64{}
	}
	result := make([]int64, 0, len(input))
	for _, item := range input {
		result = append(result, int64(item))
	}
	return result
}

func convertStringSliceToInt64Slice(input []string) []int64 {
	if len(input) == 0 {
		return []int64{}
	}
	result := make([]int64, 0, len(input))
	for _, item := range input {
		result = append(result, parseStringInt64(item))
	}
	return result
}

func parseStringInt64(value string) int64 {
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
