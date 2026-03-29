package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	paymentbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

type adminPaymentRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminPaymentRepo 创建支付方式仓储
func NewAdminPaymentRepo(data *Data, logger log.Logger) paymentbiz.PaymentRepo {
	return &adminPaymentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create 创建支付方式
func (r *adminPaymentRepo) Create(ctx context.Context, method *paymentbiz.PaymentMethod) (*paymentbiz.PaymentMethod, error) {
	// 生成8位随机token
	token := tool.GenerateRandomString(8)

	// 生成通知URL
	notifyURL := fmt.Sprintf("%s/v1/notify/%s/%s", method.Domain, method.Platform, token)

	created, err := r.data.db.ProxyPayment.
		Create().
		SetName(method.Name).
		SetPlatform(method.Platform).
		SetDescription(method.Description).
		SetIcon(method.Icon).
		SetDomain(method.Domain).
		SetConfig(method.Config).
		SetFeeMode(uint(method.FeeMode)).
		SetFeePercent(method.FeePercent).
		SetFeeAmount(method.FeeAmount).
		SetEnable(method.Enable).
		SetToken(token).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &paymentbiz.PaymentMethod{
		ID:          int64(created.ID),
		Name:        created.Name,
		Platform:    created.Platform,
		Description: created.Description,
		Icon:        created.Icon,
		Domain:      created.Domain,
		Config:      created.Config,
		FeeMode:     int32(created.FeeMode),
		FeePercent:  created.FeePercent,
		FeeAmount:   created.FeeAmount,
		Enable:      created.Enable,
		Token:       created.Token,
		NotifyURL:   notifyURL,
	}, nil
}

// Update 更新支付方式
func (r *adminPaymentRepo) Update(ctx context.Context, method *paymentbiz.PaymentMethod) (*paymentbiz.PaymentMethod, error) {
	// 获取原有支付方式
	original, err := r.data.db.ProxyPayment.
		Query().
		Where(
			proxypayment.ID(method.ID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewPaymentNotFoundError()
		}
		return nil, err
	}

	// 生成通知URL (使用原有token)
	notifyURL := fmt.Sprintf("%s/v1/notify/%s/%s", method.Domain, method.Platform, original.Token)

	updated, err := original.
		Update().
		SetName(method.Name).
		SetPlatform(method.Platform).
		SetDescription(method.Description).
		SetIcon(method.Icon).
		SetDomain(method.Domain).
		SetConfig(method.Config).
		SetFeeMode(uint(method.FeeMode)).
		SetFeePercent(method.FeePercent).
		SetFeeAmount(method.FeeAmount).
		SetEnable(method.Enable).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &paymentbiz.PaymentMethod{
		ID:          int64(updated.ID),
		Name:        updated.Name,
		Platform:    updated.Platform,
		Description: updated.Description,
		Icon:        updated.Icon,
		Domain:      updated.Domain,
		Config:      updated.Config,
		FeeMode:     int32(updated.FeeMode),
		FeePercent:  updated.FeePercent,
		FeeAmount:   updated.FeeAmount,
		Enable:      updated.Enable,
		Token:       updated.Token,
		NotifyURL:   notifyURL,
	}, nil
}

// Delete 删除支付方式
func (r *adminPaymentRepo) Delete(ctx context.Context, id int) error {
	result, err := r.data.db.ProxyPayment.
		Delete().
		Where(
			proxypayment.ID(int64(id)),
		).
		Exec(ctx)

	if err != nil {
		return err
	}

	if result == 0 {
		return responsecode.NewPaymentNotFoundError()
	}

	return nil
}

// Get 获取支付方式详情
func (r *adminPaymentRepo) Get(ctx context.Context, id int) (*paymentbiz.PaymentMethod, error) {
	payment, err := r.data.db.ProxyPayment.
		Query().
		Where(
			proxypayment.ID(int64(id)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, responsecode.NewPaymentNotFoundError()
		}
		return nil, err
	}

	// 生成通知URL
	notifyURL := fmt.Sprintf("%s/v1/notify/%s/%s", payment.Domain, payment.Platform, payment.Token)

	return &paymentbiz.PaymentMethod{
		ID:          int64(payment.ID),
		Name:        payment.Name,
		Platform:    payment.Platform,
		Description: payment.Description,
		Icon:        payment.Icon,
		Domain:      payment.Domain,
		Config:      payment.Config,
		FeeMode:     int32(payment.FeeMode),
		FeePercent:  payment.FeePercent,
		FeeAmount:   payment.FeeAmount,
		Enable:      payment.Enable,
		Token:       payment.Token,
		NotifyURL:   notifyURL,
	}, nil
}

// List 获取支付方式列表
func (r *adminPaymentRepo) List(ctx context.Context, page, size int, platform, search string, enable *bool) (int64, []*paymentbiz.PaymentMethod, error) {
	query := r.data.db.ProxyPayment.
		Query()

	// 平台筛选
	if platform != "" {
		query = query.Where(proxypayment.Platform(platform))
	}

	// 启用状态筛选
	if enable != nil {
		query = query.Where(proxypayment.Enable(*enable))
	}

	// 搜索条件
	if search != "" {
		query = query.Where(
			proxypayment.Or(
				proxypayment.NameContains(search),
				proxypayment.DescriptionContains(search),
			),
		)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	// 分页查询
	payments, err := query.
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Order(ent.Desc(proxypayment.FieldID)).
		All(ctx)

	if err != nil {
		return 0, nil, err
	}

	result := make([]*paymentbiz.PaymentMethod, 0, len(payments))
	for _, p := range payments {
		// 生成通知URL
		notifyURL := fmt.Sprintf("%s/v1/notify/%s/%s", p.Domain, p.Platform, p.Token)

		result = append(result, &paymentbiz.PaymentMethod{
			ID:          int64(p.ID),
			Name:        p.Name,
			Platform:    p.Platform,
			Description: p.Description,
			Icon:        p.Icon,
			Domain:      p.Domain,
			Config:      p.Config,
			FeeMode:     int32(p.FeeMode),
			FeePercent:  p.FeePercent,
			FeeAmount:   p.FeeAmount,
			Enable:      p.Enable,
			Token:       p.Token,
			NotifyURL:   notifyURL,
		})
	}

	return int64(total), result, nil
}
