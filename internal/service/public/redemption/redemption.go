package redemption

import (
	"context"

	v1 "github.com/OmnTeam/npanel-pro/api/public/redemption/v1"
	redemptionBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/redemption"
	"github.com/OmnTeam/npanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
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
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidAccess)
	}

	result, err := s.uc.RedeemCode(ctx, userID, req.Code)
	if err != nil {
		return nil, err
	}

	return &v1.RedeemCodeReply{
		Message: result.Message,
	}, nil
}
