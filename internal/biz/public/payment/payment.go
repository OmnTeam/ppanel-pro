package payment

import (
	"context"
)

// PaymentRepo Public Payment数据仓库接口
type PaymentRepo interface {
	// GetAvailablePaymentMethods 获取可用支付方式
	GetAvailablePaymentMethods(ctx context.Context, tenantID int64) ([]*PaymentMethod, error)
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID          int64
	Name        string
	Platform    string
	Description string
	Icon        string
	FeeMode     int32
	FeePercent  int64
	FeeAmount   int64
}

// PaymentUseCase Public Payment用例
type PaymentUseCase struct {
	repo PaymentRepo
}

// NewPaymentUseCase 创建Public Payment用例
func NewPaymentUseCase(repo PaymentRepo) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (uc *PaymentUseCase) GetAvailablePaymentMethods(ctx context.Context, tenantID int64) ([]*PaymentMethod, error) {
	return uc.repo.GetAvailablePaymentMethods(ctx, tenantID)
}
