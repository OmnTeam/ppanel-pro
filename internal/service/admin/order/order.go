package order

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/order/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/order"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

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
	if req.UserId == 0 {
		return nil, responsecode.ErrUserIDRequired()
	}

	// 创建订单
	err := s.uc.CreateOrder(ctx, 0,
		req.UserId,
		req.Type,
		req.Quantity,
		req.Price,
		req.Amount,
		req.Discount,
		req.Coupon,
		req.CouponDiscount,
		req.Commission,
		req.FeeAmount,
		req.PaymentId,
		req.TradeNo,
		req.Status,
		req.SubscribeId,
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
	if req.Id == 0 {
		return nil, responsecode.ErrOrderIDRequired()
	}

	// 更新订单状态
	err := s.uc.UpdateOrderStatus(ctx, 0, req.Id, req.Status, req.PaymentId, req.TradeNo)
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

	// 获取订单列表
	list, total, err := s.uc.GetOrderList(ctx, 0, req.Page, req.Size, req.UserId, req.Status, req.SubscribeId, req.Search)
	if err != nil {
		s.log.Errorw("msg", "get order list failed", "error", err)
		return nil, responsecode.ErrOrderListFailed()
	}

	// 转换为响应格式
	var items []*v1.OrderItem
	for _, o := range list {
		item := &v1.OrderItem{
			Id:             o.ID,
			ParentId:       o.ParentID,
			UserId:         o.UserID,
			OrderNo:        o.OrderNo,
			Type:           int32(o.Type),
			Quantity:       o.Quantity,
			Price:          o.Price,
			Amount:         o.Amount,
			GiftAmount:     0, // Field doesn't exist in schema
			Discount:       o.Discount,
			Coupon:         o.Coupon,
			CouponDiscount: o.CouponDiscount,
			Commission:     o.Commission,
			PaymentId:      o.PaymentID,
			Method:         o.Method,
			FeeAmount:      o.FeeAmount,
			TradeNo:        o.TradeNo,
			Status:         int32(o.Status),
			SubscribeId:    o.SubscribeID,
			SubscribeToken: "", // Field doesn't exist in schema
			IsNew:          false, // Field doesn't exist in schema
			CreatedAt:      o.CreatedAt.Unix(),
			UpdatedAt:      o.UpdatedAt.Unix(),
		}
		items = append(items, item)
	}

	return &v1.GetOrderListReply{
		Code:    int32(responsecode.AdminGetOrderListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetOrderListSuccess],
		Data: &v1.GetOrderListData{
			List:  items,
			Total: total,
		},
	}, nil
}
