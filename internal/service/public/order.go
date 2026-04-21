package public

import (
	"context"
	"strconv"

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

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

// CloseOrder closes an order
func (s *PublicOrderService) CloseOrder(ctx context.Context, req *pb.CloseOrderRequest) (*pb.OrderCloseReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	err := s.uc.CloseOrder(ctx, int(userID), req.OrderNo)
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

	order, err := s.uc.QueryOrderDetail(ctx, int(userID), req.OrderNo)
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

	orders, total, err := s.uc.QueryOrderList(ctx, int(userID), int(req.Page), int(req.Size), req.Status, req.Type)
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
			Total: int32(total),
		},
	}, nil
}

// PreCreateOrder validates and calculates order price
func (s *PublicOrderService) PreCreateOrder(ctx context.Context, req *pb.PreCreateOrderRequest) (*pb.OrderPreCreateReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.PreCreateOrderParams{
		UserID:           userID,
		Type:             req.Type,
		SubscribeID:      parseInt64(req.SubscribeId),
		SubscribeGroupID: parseInt64(req.SubscribeGroupId),
		Quantity:         int64(req.Quantity),
		Coupon:           req.Coupon,
		Payment:          parseInt64(req.Payment),
	}

	result, err := s.uc.PreCreateOrder(ctx, params)
	if err != nil {
		return nil, err
	}

	return &pb.OrderPreCreateReply{
		Code:    int32(responsecode.OrderPreCreateSuccess),
		Message: responsecode.CodeMessages[responsecode.OrderPreCreateSuccess],
		Data: &pb.OrderPreCreateData{
			Price:          strconv.FormatInt(result.Price, 10),
			Amount:         strconv.FormatInt(result.Amount, 10),
			Discount:       strconv.FormatInt(result.Discount, 10),
			CouponDiscount: strconv.FormatInt(result.CouponDiscount, 10),
			FeeAmount:      strconv.FormatInt(result.FeeAmount, 10),
			Commission:     strconv.FormatInt(result.Commission, 10),
			GiftAmount:     strconv.FormatInt(result.GiftAmount, 10),
			Valid:          result.Valid,
			ErrorMessage:   result.Message,
		},
	}, nil
}

// Purchase creates a purchase order
func (s *PublicOrderService) Purchase(ctx context.Context, req *pb.PurchaseRequest) (*pb.PurchaseReply, error) {
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.PurchaseParams{
		UserID:      userID,
		SubscribeID: parseInt64(req.SubscribeId),
		Quantity:    int64(req.Quantity),
		Coupon:      req.Coupon,
		Payment:     parseInt64(req.Payment),
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
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.RechargeParams{
		UserID:  userID,
		Amount:  parseInt64(req.Amount),
		Payment: parseInt64(req.Payment),
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
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.RenewalParams{
		UserID:          userID,
		UserSubscribeID: parseInt64(req.UserSubscribeId),
		Quantity:        int64(req.Quantity),
		Coupon:          req.Coupon,
		Payment:         parseInt64(req.Payment),
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
	// Get user ID from context
	userID := middleware.GetUserID(ctx)

	params := &publicBiz.ResetTrafficParams{
		UserID:          userID,
		UserSubscribeID: parseInt64(req.UserSubscribeId),
		Payment:         parseInt64(req.Payment),
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
			Id:          strconv.FormatInt(order.Payment.ID, 10),
			Name:        order.Payment.Name,
			Platform:    order.Payment.Platform,
			Description: order.Payment.Description,
			Icon:        order.Payment.Icon,
			FeeMode:     order.Payment.FeeMode,
			FeePercent:  strconv.FormatInt(order.Payment.FeePercent, 10),
			FeeAmount:   strconv.FormatInt(order.Payment.FeeAmount, 10),
		}
	}

	// 转换 Subscribe 对象
	var subscribe *pb.Subscribe
	if order.Subscribe != nil {
		// 转换折扣列表
		var discounts []*pb.SubscribeDiscount
		for _, d := range order.Subscribe.Discount {
			discounts = append(discounts, &pb.SubscribeDiscount{
				Quantity: int32(d.Quantity),
				Discount: strconv.FormatInt(d.Discount, 10),
			})
		}

		subscribe = &pb.Subscribe{
			Id:             strconv.FormatInt(order.Subscribe.ID, 10),
			Name:           order.Subscribe.Name,
			Language:       order.Subscribe.Language,
			Description:    order.Subscribe.Description,
			UnitPrice:      strconv.FormatInt(order.Subscribe.UnitPrice, 10),
			UnitTime:       order.Subscribe.UnitTime,
			Discount:       discounts,
			Replacement:    strconv.FormatInt(order.Subscribe.Replacement, 10),
			Inventory:      strconv.FormatInt(order.Subscribe.Inventory, 10),
			Traffic:        strconv.FormatInt(order.Subscribe.Traffic, 10),
			SpeedLimit:     strconv.FormatInt(order.Subscribe.SpeedLimit, 10),
			DeviceLimit:    strconv.FormatInt(order.Subscribe.DeviceLimit, 10),
			Quota:          strconv.FormatInt(order.Subscribe.Quota, 10),
			Nodes:          convertIntSliceToInt32Slice(order.Subscribe.Nodes),
			NodeTags:       order.Subscribe.NodeTags,
			Show:           order.Subscribe.Show,
			Sell:           order.Subscribe.Sell,
			Sort:           strconv.FormatInt(order.Subscribe.Sort, 10),
			DeductionRatio: strconv.FormatInt(order.Subscribe.DeductionRatio, 10),
			AllowDeduction: order.Subscribe.AllowDeduction,
		}
	}

	return &pb.OrderDetail{
		Id:             strconv.FormatInt(order.ID, 10),
		ParentId:       strconv.FormatInt(order.ParentID, 10),
		UserId:         strconv.FormatInt(order.UserID, 10),
		OrderNo:        order.OrderNo,
		Type:           order.Type,
		Quantity:       int32(order.Quantity),
		Price:          strconv.FormatInt(order.Price, 10),
		Amount:         strconv.FormatInt(order.Amount, 10),
		GiftAmount:     strconv.FormatInt(order.GiftAmount, 10),
		Discount:       strconv.FormatInt(order.Discount, 10),
		Coupon:         order.Coupon,
		CouponDiscount: strconv.FormatInt(order.CouponDiscount, 10),
		Commission:     "0", // Prevent commission amount leakage (set to 0 for public API)
		Payment:        payment,
		Method:         order.Method,
		FeeAmount:      strconv.FormatInt(order.FeeAmount, 10),
		TradeNo:        order.TradeNo,
		Status:         order.Status,
		SubscribeId:    strconv.FormatInt(order.SubscribeID, 10),
		SubscribeToken: order.SubscribeToken,
		IsNew:          order.IsNew,
		CreatedAt:      strconv.FormatInt(order.CreatedAt, 10),
		UpdatedAt:      strconv.FormatInt(order.UpdatedAt, 10),
		Subscribe:      subscribe,
		SubscribeName:  order.SubscribeName,
		PaymentName:    order.PaymentName,
		StatusText:     order.StatusText,
		TypeText:       order.TypeText,
	}
}

// convertIntSliceToInt32Slice converts []int to []int32
func convertIntSliceToInt32Slice(input []int) []int32 {
	if input == nil {
		return nil
	}
	result := make([]int32, len(input))
	for i, v := range input {
		result[i] = int32(v)
	}
	return result
}
