package payment

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/internal/model/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	paymentPkg "github.com/OmnTeam/ppanel-pro/pkg/payment"
	"github.com/OmnTeam/ppanel-pro/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
)

// PaymentMethod 支付方式模型
type PaymentMethod struct {
	ID          int64
	TenantID    int64
	Name        string
	Platform    string
	Description string
	Icon        string
	Domain      string
	Config      string // JSON配置
	FeeMode     int32
	FeePercent  int64
	FeeAmount   int64
	Enable      bool
	Token       string
	NotifyURL   string
}

// PaymentRepo 支付方式仓储接口
type PaymentRepo interface {
	// Create 创建支付方式
	Create(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	// Update 更新支付方式
	Update(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	// Delete 删除支付方式
	Delete(ctx context.Context, id int64) error
	// Get 获取支付方式详情
	Get(ctx context.Context, id int64) (*PaymentMethod, error)
	// List 获取支付方式列表
	List(ctx context.Context, page, size int64, platform, search string, enable *bool) (int64, []*PaymentMethod, error)
}

// PaymentUsecase 支付方式用例
type PaymentUsecase struct {
	repo PaymentRepo
	log  *log.Helper
}

// NewPaymentUsecase 创建支付方式用例
func NewPaymentUsecase(repo PaymentRepo, logger log.Logger) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreatePaymentMethod 创建支付方式
func (uc *PaymentUsecase) CreatePaymentMethod(ctx context.Context, name, platform, description, icon, domain, config string, feeMode int32, feePercent, feeAmount int64, enable *bool) (*PaymentMethod, error) {
	// 验证支付平台是否支持
	if payment.ParsePlatform(platform) == payment.UNSUPPORTED {
		return nil, responsecode.NewUnsupportedPlatformError()
	}

	method := &PaymentMethod{
		Name:        name,
		Platform:    platform,
		Description: description,
		Icon:        icon,
		Domain:      domain,
		Config:      config,
		FeeMode:     feeMode,
		FeePercent:  feePercent,
		FeeAmount:   feeAmount,
	}

	// 处理启用状态
	if enable != nil {
		method.Enable = *enable
	} else {
		method.Enable = false
	}

	return uc.repo.Create(ctx, method)
}

// UpdatePaymentMethod 更新支付方式
func (uc *PaymentUsecase) UpdatePaymentMethod(ctx context.Context, id int64, name, platform, description, icon, domain, config string, feeMode int32, feePercent, feeAmount int64, enable *bool) (*PaymentMethod, error) {
	// 验证支付平台是否支持
	if payment.ParsePlatform(platform) == payment.UNSUPPORTED {
		return nil, responsecode.NewUnsupportedPlatformError()
	}

	// 先获取原有支付方式
	_, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	method := &PaymentMethod{
		ID:          id,
		Name:        name,
		Platform:    platform,
		Description: description,
		Icon:        icon,
		Domain:      domain,
		Config:      config,
		FeeMode:     feeMode,
		FeePercent:  feePercent,
		FeeAmount:   feeAmount,
	}

	// 处理启用状态
	if enable != nil {
		method.Enable = *enable
	} else {
		method.Enable = false
	}

	return uc.repo.Update(ctx, method)
}

// DeletePaymentMethod 删除支付方式
func (uc *PaymentUsecase) DeletePaymentMethod(ctx context.Context, id int64) error {
	// 先验证支付方式是否存在
	_, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}

// GetPaymentMethodList 获取支付方式列表
func (uc *PaymentUsecase) GetPaymentMethodList(ctx context.Context, page, size int64, platform, search string, enable *bool) (int64, []*PaymentMethod, error) {
	return uc.repo.List(ctx, page, size, platform, search, enable)
}

// GetPaymentPlatform 获取支付平台列表
// 完全复刻原项目 server-master/internal/logic/admin/payment/getPaymentPlatformLogic.go
func (uc *PaymentUsecase) GetPaymentPlatform(ctx context.Context) []types.PlatformInfo {
	return paymentPkg.GetSupportedPlatforms()
}
