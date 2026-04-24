package order

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
)

// OrderRepo is the interface for order repository
type OrderRepo interface {
	CreateOrder(ctx context.Context, userID int, orderType uint32, quantity, price, amount, discount int,
		coupon string, couponDiscount, commission, feeAmount, paymentID int, tradeNo string,
		status uint32, subscribeID int) error
	UpdateOrderStatus(ctx context.Context, id int, status uint32, paymentID int, tradeNo string) error
	GetOrderList(ctx context.Context, page, size, userID int, status uint32, subscribeID int, search string) ([]*ent.ProxyOrder, int32, error)
}

// OrderUseCase is the use case for order operations
type OrderUseCase struct {
	repo OrderRepo
	log  *log.Helper
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/order")),
	}
}

// CreateOrder creates a new order
func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID int, orderType uint32, quantity, price, amount, discount int,
	coupon string, couponDiscount, commission, feeAmount, paymentID int, tradeNo string,
	status uint32, subscribeID int) error {

	return uc.repo.CreateOrder(ctx, userID, orderType, quantity, price, amount, discount,
		coupon, couponDiscount, commission, feeAmount, paymentID, tradeNo, status, subscribeID)
}

// UpdateOrderStatus updates order status
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id int, status uint32, paymentID int, tradeNo string) error {
	return uc.repo.UpdateOrderStatus(ctx, id, status, paymentID, tradeNo)
}

// GetOrderList gets order list
func (uc *OrderUseCase) GetOrderList(ctx context.Context, page, size, userID int, status uint32, subscribeID int, search string) ([]*ent.ProxyOrder, int32, error) {
	orders, total, err := uc.repo.GetOrderList(ctx, page, size, userID, status, subscribeID, search)
	return orders, total, err
}
