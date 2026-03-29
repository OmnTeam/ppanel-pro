package redemption

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/redemption/v1"
	redemptionBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/redemption"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// RedemptionService 兑换码服务
type RedemptionService struct {
	v1.UnimplementedRedemptionServiceServer
	uc *redemptionBiz.RedemptionUseCase
}

// NewRedemptionService 创建兑换码服务
func NewRedemptionService(uc *redemptionBiz.RedemptionUseCase) *RedemptionService {
	return &RedemptionService{
		uc: uc,
	}
}

// RedeemCode 兑换兑换码
func (s *RedemptionService) RedeemCode(ctx context.Context, req *v1.RedeemCodeRequest) (*v1.RedeemCodeReply, error) {
	// 获取当前用户
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &v1.RedeemCodeReply{
			Code:    responsecode.ErrInvalidAccess,
			Message: "未授权访问",
		}, nil
	}

	// 调用业务逻辑
	result, err := s.uc.RedeemCode(ctx, userID, req.Code)
	if err != nil {
		return &v1.RedeemCodeReply{
			Code:    responsecode.ErrInternalError,
			Message: err.Error(),
		}, nil
	}

	return &v1.RedeemCodeReply{
		Code:    0,
		Message: "兑换成功",
		Data: &v1.RedeemCodeData{
			OrderNo: result.OrderNo,
			Message: result.Message,
		},
	}, nil
}
