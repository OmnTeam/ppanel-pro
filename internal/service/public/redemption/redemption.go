package redemption

import (
	"context"
	"strconv"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/redemption/v1"
	redemptionBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/redemption"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	kerrors "github.com/go-kratos/kratos/v2/errors"
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
		code := responsecode.ErrInternalError
		message := err.Error()
		if se := kerrors.FromError(err); se != nil {
			if customCode, ok := se.Metadata["custom_code"]; ok {
				if parsed, parseErr := strconv.Atoi(customCode); parseErr == nil {
					code = parsed
				}
			}
			if se.Message != "" {
				message = se.Message
			}
		}
		return &v1.RedeemCodeReply{
			Code:    int32(code),
			Message: message,
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
