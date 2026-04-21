package order

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/order/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/order"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

func parseOrderStringInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return val, nil
}

type OrderService struct {
	v1.UnimplementedOrderServiceServer

	uc  *order.OrderUseCase
	log *log.Helper
}

func NewOrderService(uc *order.OrderUseCase, logger log.Logger) *OrderService {
	return &OrderService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/admin/order")),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
	// 验证用户ID
	if req.UserId == "" {
		return nil, responsecode.ErrUserIDRequired()
	}

	// 转换参数类型
	userId, err := parseOrderStringInt(req.UserId)
	if err != nil {
		return nil, err
	}
	subscribeId := 0
	if req.SubscribeId != "" {
		subscribeId, err = parseOrderStringInt(req.SubscribeId)
		if err != nil {
			return nil, err
		}
	}
	paymentId := 0
	if req.PaymentId != "" {
		paymentId, err = parseOrderStringInt(req.PaymentId)
		if err != nil {
			return nil, err
		}
	}

	// 创建订单
	err = s.uc.CreateOrder(ctx,
		userId,
		req.Type,
		int(req.Quantity),
		int(req.Price),
		int(req.Amount),
		int(req.Discount),
		req.Coupon,
		int(req.CouponDiscount),
		int(req.Commission),
		int(req.FeeAmount),
		paymentId,
		req.TradeNo,
		req.Status,
		subscribeId,
	)
	if err != nil {
		s.log.Errorw("msg", "create order failed", "error", err)
		return nil, responsecode.ErrOrderCreateFailed()
	}

	return &v1.CreateOrderReply{
		Code:    int32(responsecode.AdminCreateOrderSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateOrderSuccess],
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.UpdateOrderStatusReply, error) {
	// 验证订单ID
	if req.Id == "" {
		return nil, responsecode.ErrOrderIDRequired()
	}

	// 转换参数类型
	id, err := parseOrderStringInt(req.Id)
	if err != nil {
		return nil, err
	}
	paymentId := 0
	if req.PaymentId != "" {
		paymentId, err = parseOrderStringInt(req.PaymentId)
		if err != nil {
			return nil, err
		}
	}

	// 更新订单状态
	err = s.uc.UpdateOrderStatus(ctx, id, req.Status, paymentId, req.TradeNo)
	if err != nil {
		s.log.Errorw("msg", "update order status failed", "error", err)
		return nil, responsecode.ErrOrderUpdateFailed()
	}

	return &v1.UpdateOrderStatusReply{
		Code:    int32(responsecode.AdminUpdateOrderStatusSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateOrderStatusSuccess],
	}, nil
}

// GetOrderList 获取订单列表
func (s *OrderService) GetOrderList(ctx context.Context, req *v1.GetOrderListRequest) (*v1.GetOrderListReply, error) {
	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	// 转换参数类型
	var err error
	userId := 0
	if req.UserId != "" {
		userId, err = parseOrderStringInt(req.UserId)
		if err != nil {
			return nil, err
		}
	}
	subscribeId := 0
	if req.SubscribeId != "" {
		subscribeId, err = parseOrderStringInt(req.SubscribeId)
		if err != nil {
			return nil, err
		}
	}

	// 获取订单列表
	list, total, err := s.uc.GetOrderList(ctx, int(req.Page), int(req.Size), userId, req.Status, subscribeId, req.Search)
	if err != nil {
		s.log.Errorw("msg", "get order list failed", "error", err)
		return nil, responsecode.ErrOrderListFailed()
	}

	// 转换为响应格式
	var items []*v1.OrderItem
	for _, o := range list {
		item := &v1.OrderItem{
			Id:             strconv.Itoa(int(o.ID)),
			ParentId:       strconv.Itoa(int(o.ParentID)),
			UserId:         strconv.Itoa(int(o.UserID)),
			OrderNo:        o.OrderNo,
			Type:           int32(o.Type),
			Quantity:       int64(o.Quantity),
			Price:          int64(o.Price),
			Amount:         int64(o.Amount),
			GiftAmount:     0, // Field doesn't exist in schema
			Discount:       int64(o.Discount),
			Coupon:         o.Coupon,
			CouponDiscount: int64(o.CouponDiscount),
			Commission:     int64(o.Commission),
			PaymentId:      strconv.FormatInt(int64(o.PaymentID), 10),
			Method:         o.Method,
			FeeAmount:      int64(o.FeeAmount),
			TradeNo:        o.TradeNo,
			Status:         int32(o.Status),
			SubscribeId:    strconv.Itoa(int(o.SubscribeID)),
			SubscribeToken: "",    // Field doesn't exist in schema
			IsNew:          false, // Field doesn't exist in schema
			CreatedAt:      o.CreatedAt.UnixMilli(),
			UpdatedAt:      o.UpdatedAt.UnixMilli(),
		}
		items = append(items, item)
	}

	return &v1.GetOrderListReply{
		Code:    int32(responsecode.AdminGetOrderListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetOrderListSuccess],
		Data: &v1.GetOrderListData{
			List:  items,
			Total: int64(total),
		},
	}, nil
}
