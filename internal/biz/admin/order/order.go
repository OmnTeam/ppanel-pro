package order

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
)

// OrderRepo is the interface for order repository
type OrderRepo interface {
	CreateOrder(ctx context.Context, tenantID, userID int64, orderType int32, quantity, price, amount, discount int64,
		coupon string, couponDiscount, commission, feeAmount, paymentID int64, tradeNo string,
		status int32, subscribeID int64) error
	UpdateOrderStatus(ctx context.Context, id, tenantID int64, status int32, paymentID int64, tradeNo string) error
	GetOrderList(ctx context.Context, tenantID, page, size, userID int64, status int32, subscribeID int64, search string) ([]*ent.ProxyOrder, int64, error)
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
func (uc *OrderUseCase) CreateOrder(ctx context.Context, tenantID, userID int64, orderType int32, quantity, price, amount, discount int64,
	coupon string, couponDiscount, commission, feeAmount, paymentID int64, tradeNo string,
	status int32, subscribeID int64) error {

	return uc.repo.CreateOrder(ctx, tenantID, userID, orderType, quantity, price, amount, discount,
		coupon, couponDiscount, commission, feeAmount, paymentID, tradeNo, status, subscribeID)
}

// UpdateOrderStatus updates order status
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id, tenantID int64, status int32, paymentID int64, tradeNo string) error {
	return uc.repo.UpdateOrderStatus(ctx, id, tenantID, status, paymentID, tradeNo)
}

// GetOrderList gets order list
func (uc *OrderUseCase) GetOrderList(ctx context.Context, tenantID, page, size, userID int64, status int32, subscribeID int64, search string) ([]*ent.ProxyOrder, int64, error) {
	return uc.repo.GetOrderList(ctx, tenantID, page, size, userID, status, subscribeID, search)
}
