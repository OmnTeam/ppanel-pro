package data

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxypayment"
	paymentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type publicPaymentRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicPaymentRepo 创建Public Payment仓库
func NewPublicPaymentRepo(data *Data, logger log.Logger) paymentBiz.PaymentRepo {
	return &publicPaymentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (r *publicPaymentRepo) GetAvailablePaymentMethods(ctx context.Context, tenantID int64) ([]*paymentBiz.PaymentMethod, error) {
	// 查询enable=true的支付方式
	methods, err := r.data.db.ProxyPayment.Query().
		Where(
			proxypayment.Enable(true),
		).
		Order(ent.Asc(proxypayment.FieldID)).
		All(ctx)

	if err != nil {
		r.log.Errorf("GetAvailablePaymentMethods query error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	result := make([]*paymentBiz.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		result = append(result, &paymentBiz.PaymentMethod{
			ID:          int64(m.ID),
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     int32(m.FeeMode),
			FeePercent:  int64(m.FeePercent),
			FeeAmount:   m.FeeAmount,
		})
	}

	return result, nil
}
