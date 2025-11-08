package public

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/public/order/v1"
	publicBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type PublicOrderService struct {
	pb.UnimplementedPublicOrderServer

	uc *publicBiz.OrderUsecase
}

func NewPublicOrderService(uc *publicBiz.OrderUsecase) *PublicOrderService {
	return &PublicOrderService{
		uc: uc,
	}
}

// CloseOrder closes an order
func (s *PublicOrderService) CloseOrder(ctx context.Context, req *pb.CloseOrderRequest) (*pb.OrderCloseReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	err := s.uc.CloseOrder(ctx, userID, req.OrderNo)
	if err != nil {
		return nil, err
	}

	return &pb.OrderCloseReply{
		Code:    int32(responsecode.OrderCloseSuccess),
		Message: responsecode.CodeMessages[responsecode.OrderCloseSuccess],
		Data: &pb.OrderCloseData{
			Success: true,
		},
	}, nil
}

// QueryOrderDetail queries order detail
func (s *PublicOrderService) QueryOrderDetail(ctx context.Context, req *pb.QueryOrderDetailRequest) (*pb.OrderDetailReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	order, err := s.uc.QueryOrderDetail(ctx, userID, req.OrderNo)
	if err != nil {
		return nil, err
	}

	return &pb.OrderDetailReply{
		Code:    int32(responsecode.OrderDetailQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.OrderDetailQuerySuccess],
		Data:    s.convertToProtoOrderDetail(order),
	}, nil
}

// QueryOrderList queries order list
func (s *PublicOrderService) QueryOrderList(ctx context.Context, req *pb.QueryOrderListRequest) (*pb.OrderListReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	orders, total, err := s.uc.QueryOrderList(ctx, userID, req.Page, req.Size, req.Status, req.Type)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*pb.OrderDetail, len(orders))
	for i, order := range orders {
		protoOrders[i] = s.convertToProtoOrderDetail(order)
	}

	return &pb.OrderListReply{
		Code:    int32(responsecode.OrderListQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.OrderListQuerySuccess],
		Data: &pb.OrderListData{
			List:  protoOrders,
			Total: total,
		},
	}, nil
}

// PreCreateOrder validates and calculates order price
func (s *PublicOrderService) PreCreateOrder(ctx context.Context, req *pb.PreCreateOrderRequest) (*pb.OrderPreCreateReply, error) {
	// Get tenant ID and user ID from context
	tenantID := int64(0) // 默认租户ID，与原项目保持一致
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.PreCreateOrderParams{
		TenantID:         tenantID,
		UserID:           userID,
		Type:             req.Type,
		SubscribeID:      req.SubscribeId,
		SubscribeGroupID: req.SubscribeGroupId,
		Quantity:         req.Quantity,
		Coupon:           req.Coupon,
		Payment:          req.Payment,
	}

	result, err := s.uc.PreCreateOrder(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.OrderPreCreateReply{
		Code:    int32(responsecode.OrderPreCreateSuccess),
		Message: responsecode.CodeMessages[responsecode.OrderPreCreateSuccess],
		Data: &pb.OrderPreCreateData{
			Price:          result.Price,
			Amount:         result.Amount,
			Discount:       result.Discount,
			CouponDiscount: result.CouponDiscount,
			FeeAmount:      result.FeeAmount,
			Commission:     result.Commission,
			GiftAmount:     result.GiftAmount,
			Valid:          result.Valid,
			ErrorMessage:   result.Message,
		},
	}, nil
}

// Purchase creates a purchase order
func (s *PublicOrderService) Purchase(ctx context.Context, req *pb.PurchaseRequest) (*pb.PurchaseReply, error) {
	// Get tenant ID and user ID from context
	tenantID := int64(0) // 默认租户ID，与原项目保持一致
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.PurchaseParams{
		TenantID:    tenantID,
		UserID:      userID,
		SubscribeID: req.SubscribeId,
		Quantity:    req.Quantity,
		Coupon:      req.Coupon,
		Payment:     req.Payment,
	}

	result, err := s.uc.Purchase(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.PurchaseReply{
		Code:    int32(responsecode.PurchaseSuccess),
		Message: responsecode.CodeMessages[responsecode.PurchaseSuccess],
		Data: &pb.PurchaseData{
			OrderNo: result.OrderNo,
		},
	}, nil
}

// Recharge creates a recharge order
func (s *PublicOrderService) Recharge(ctx context.Context, req *pb.RechargeRequest) (*pb.RechargeReply, error) {
	// Get tenant ID and user ID from context
	tenantID := int64(0) // 默认租户ID，与原项目保持一致
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.RechargeParams{
		TenantID: tenantID,
		UserID:   userID,
		Amount:   req.Amount,
		Payment:  req.Payment,
	}

	result, err := s.uc.Recharge(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.RechargeReply{
		Code:    int32(responsecode.RechargeSuccess),
		Message: responsecode.CodeMessages[responsecode.RechargeSuccess],
		Data: &pb.RechargeData{
			OrderNo: result.OrderNo,
		},
	}, nil
}

// Renewal creates a renewal order
func (s *PublicOrderService) Renewal(ctx context.Context, req *pb.RenewalRequest) (*pb.RenewalReply, error) {
	// Get tenant ID and user ID from context
	tenantID := int64(0) // 默认租户ID，与原项目保持一致
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.RenewalParams{
		TenantID:        tenantID,
		UserID:          userID,
		UserSubscribeID: req.UserSubscribeId,
		Quantity:        req.Quantity,
		Coupon:          req.Coupon,
		Payment:         req.Payment,
	}

	result, err := s.uc.Renewal(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.RenewalReply{
		Code:    int32(responsecode.RenewalSuccess),
		Message: responsecode.CodeMessages[responsecode.RenewalSuccess],
		Data: &pb.RenewalData{
			OrderNo: result.OrderNo,
		},
	}, nil
}

// ResetTraffic creates a reset traffic order
func (s *PublicOrderService) ResetTraffic(ctx context.Context, req *pb.ResetTrafficRequest) (*pb.TrafficResetReply, error) {
	// Get tenant ID and user ID from context
	tenantID := int64(0) // 默认租户ID，与原项目保持一致
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.ResetTrafficParams{
		TenantID:        tenantID,
		UserID:          userID,
		UserSubscribeID: req.UserSubscribeId,
		Payment:         req.Payment,
	}

	result, err := s.uc.ResetTraffic(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.TrafficResetReply{
		Code:    int32(responsecode.TrafficResetSuccess),
		Message: responsecode.CodeMessages[responsecode.TrafficResetSuccess],
		Data: &pb.TrafficResetData{
			OrderNo: result.OrderNo,
		},
	}, nil
}

// convertToProtoOrderDetail converts biz OrderDetail to proto OrderDetail
func (s *PublicOrderService) convertToProtoOrderDetail(order *publicBiz.OrderDetail) *pb.OrderDetail {
	// 转换 Payment 对象
	var payment *pb.PaymentMethod
	if order.Payment != nil {
		payment = &pb.PaymentMethod{
			Id:          order.Payment.ID,
			Name:        order.Payment.Name,
			Platform:    order.Payment.Platform,
			Description: order.Payment.Description,
			Icon:        order.Payment.Icon,
			FeeMode:     order.Payment.FeeMode,
			FeePercent:  order.Payment.FeePercent,
			FeeAmount:   order.Payment.FeeAmount,
		}
	}

	// 转换 Subscribe 对象
	var subscribe *pb.Subscribe
	if order.Subscribe != nil {
		// 转换折扣列表
		var discounts []*pb.SubscribeDiscount
		for _, d := range order.Subscribe.Discount {
			discounts = append(discounts, &pb.SubscribeDiscount{
				Quantity: d.Quantity,
				Discount: d.Discount,
			})
		}

		subscribe = &pb.Subscribe{
			Id:             order.Subscribe.ID,
			Name:           order.Subscribe.Name,
			Language:       order.Subscribe.Language,
			Description:    order.Subscribe.Description,
			UnitPrice:      order.Subscribe.UnitPrice,
			UnitTime:       order.Subscribe.UnitTime,
			Discount:       discounts,
			Replacement:    order.Subscribe.Replacement,
			Inventory:      order.Subscribe.Inventory,
			Traffic:        order.Subscribe.Traffic,
			SpeedLimit:     order.Subscribe.SpeedLimit,
			DeviceLimit:    order.Subscribe.DeviceLimit,
			Quota:          order.Subscribe.Quota,
			Nodes:          order.Subscribe.Nodes,
			NodeTags:       order.Subscribe.NodeTags,
			Show:           order.Subscribe.Show,
			Sell:           order.Subscribe.Sell,
			Sort:           order.Subscribe.Sort,
			DeductionRatio: order.Subscribe.DeductionRatio,
			AllowDeduction: order.Subscribe.AllowDeduction,
		}
	}

	return &pb.OrderDetail{
		Id:             order.ID,
		ParentId:       order.ParentID,
		UserId:         order.UserID,
		OrderNo:        order.OrderNo,
		Type:           order.Type,
		Quantity:       order.Quantity,
		Price:          order.Price,
		Amount:         order.Amount,
		GiftAmount:     order.GiftAmount,
		Discount:       order.Discount,
		Coupon:         order.Coupon,
		CouponDiscount: order.CouponDiscount,
		Commission:     0, // Prevent commission amount leakage (set to 0 for public API)
		Payment:        payment,
		Method:         order.Method,
		FeeAmount:      order.FeeAmount,
		TradeNo:        order.TradeNo,
		Status:         order.Status,
		SubscribeId:    order.SubscribeID,
		SubscribeToken: order.SubscribeToken,
		IsNew:          order.IsNew,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		Subscribe:      subscribe,
		SubscribeName:  order.SubscribeName,
		PaymentName:    order.PaymentName,
		StatusText:     order.StatusText,
		TypeText:       order.TypeText,
	}
}
