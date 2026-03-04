package subscribe

import (
	"context"
)

// SubscribeRepo Public Subscribe数据仓库接口
type SubscribeRepo interface {
	// QuerySubscribeList 查询订阅列表
	QuerySubscribeList(ctx context.Context, language string) ([]*Subscribe, int64, error)
}

// Subscribe 订阅信息
type Subscribe struct {
	ID             int64
	Name           string
	Language       string
	Description    string
	UnitPrice      int64
	UnitTime       string
	Discount       []*SubscribeDiscount
	Replacement    int64
	Inventory      int64
	Traffic        int64
	SpeedLimit     int64
	DeviceLimit    int64
	Quota          int64
	Nodes          []int
	NodeTags       []string
	Show           bool
	Sell           bool
	Sort           int64
	DeductionRatio int64
	AllowDeduction bool
	ResetCycle     int32
	CreatedAt      int64
	UpdatedAt      int64
}

// SubscribeDiscount 订阅折扣
type SubscribeDiscount struct {
	Quantity   int32
	Percentage int32
}

// SubscribeUseCase Public Subscribe用例
type SubscribeUseCase struct {
	repo SubscribeRepo
}

// NewSubscribeUseCase 创建Public Subscribe用例
func NewSubscribeUseCase(repo SubscribeRepo) *SubscribeUseCase {
	return &SubscribeUseCase{repo: repo}
}

// QuerySubscribeList 查询订阅列表
func (uc *SubscribeUseCase) QuerySubscribeList(ctx context.Context, language string) ([]*Subscribe, int64, error) {
	return uc.repo.QuerySubscribeList(ctx, language)
}
