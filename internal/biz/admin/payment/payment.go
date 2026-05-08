package payment

import (
	"context"

	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	paymentPkg "github.com/OmnTeam/npanel-pro/pkg/payment"
	"github.com/OmnTeam/npanel-pro/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
)

// PaymentMethod 支付方式模型
// 业务字段与老项目 payment 模块保持一致。
type PaymentMethod struct {
	ID          int64
	Name        string
	Platform    string
	Description string
	Icon        string
	Domain      string
	Config      string
	FeeMode     int32
	FeePercent  int64
	FeeAmount   int64
	Sort        int64
	Enable      bool
	Token       string
	SiteHost    string
}

type PaymentRepo interface {
	Create(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	Update(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	Delete(ctx context.Context, id int) error
	Get(ctx context.Context, id int) (*PaymentMethod, error)
	List(ctx context.Context, page, size int, platform, search string, enable *bool) (int32, []*PaymentMethod, error)
}

type PaymentUsecase struct {
	repo PaymentRepo
	log  *log.Helper
}

func NewPaymentUsecase(repo PaymentRepo, logger log.Logger) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PaymentUsecase) CreatePaymentMethod(ctx context.Context, name, platform, description, icon, domain, config string, feeMode int32, feePercent, feeAmount, sort int64, enable *bool) (*PaymentMethod, error) {
	if paymentPkg.ParsePlatform(platform) == paymentPkg.UNSUPPORTED {
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
		Sort:        sort,
	}
	if enable != nil {
		method.Enable = *enable
	}
	return uc.repo.Create(ctx, method)
}

func (uc *PaymentUsecase) UpdatePaymentMethod(ctx context.Context, id int, name, platform, description, icon, domain, config string, feeMode int32, feePercent, feeAmount, sort int64, enable *bool) (*PaymentMethod, error) {
	if paymentPkg.ParsePlatform(platform) == paymentPkg.UNSUPPORTED {
		return nil, responsecode.NewUnsupportedPlatformError()
	}

	method := &PaymentMethod{
		ID:          int64(id),
		Name:        name,
		Platform:    platform,
		Description: description,
		Icon:        icon,
		Domain:      domain,
		Config:      config,
		FeeMode:     feeMode,
		FeePercent:  feePercent,
		FeeAmount:   feeAmount,
		Sort:        sort,
	}
	if enable != nil {
		method.Enable = *enable
	}
	return uc.repo.Update(ctx, method)
}

func (uc *PaymentUsecase) DeletePaymentMethod(ctx context.Context, id int) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *PaymentUsecase) GetPaymentMethodList(ctx context.Context, page, size int, platform, search string, enable *bool) (int32, []*PaymentMethod, error) {
	return uc.repo.List(ctx, page, size, platform, search, enable)
}

func (uc *PaymentUsecase) GetPaymentPlatform(ctx context.Context) []types.PlatformInfo {
	return paymentPkg.GetSupportedPlatforms()
}
