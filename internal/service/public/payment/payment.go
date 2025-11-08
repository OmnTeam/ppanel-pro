package payment

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/payment/v1"
	paymentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// PaymentService Public Payment服务实现
type PaymentService struct {
	v1.UnimplementedPaymentServer
	uc *paymentBiz.PaymentUseCase
}

// NewPaymentService 创建Public Payment服务
func NewPaymentService(uc *paymentBiz.PaymentUseCase) *PaymentService {
	return &PaymentService{uc: uc}
}

// GetAvailablePaymentMethods 获取可用支付方式
func (s *PaymentService) GetAvailablePaymentMethods(ctx context.Context, req *emptypb.Empty) (*v1.PaymentMethodsReply, error) {
	// 调用业务层
	methods, err := s.uc.GetAvailablePaymentMethods(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.PaymentMethod, 0, len(methods))
	for _, m := range methods {
		list = append(list, &v1.PaymentMethod{
			Id:          m.ID,
			Name:        m.Name,
			Platform:    m.Platform,
			Description: m.Description,
			Icon:        m.Icon,
			FeeMode:     m.FeeMode,
			FeePercent:  m.FeePercent,
			FeeAmount:   m.FeeAmount,
		})
	}

	return &v1.PaymentMethodsReply{
		Code:    int32(responsecode.GetAvailablePaymentMethodsSuccess),
		Message: responsecode.CodeMessages[responsecode.GetAvailablePaymentMethodsSuccess],
		Data: &v1.PaymentMethodsData{
			List: list,
		},
	}, nil
}
