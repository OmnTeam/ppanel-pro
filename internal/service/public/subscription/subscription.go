package subscription

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/public/subscription/v1"
	subscriptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscription"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
)

type PublicSubscriptionService struct {
	pb.UnimplementedSubscriptionServer

	uc *subscriptionbiz.SubscriptionUseCase
}

func NewPublicSubscriptionService(uc *subscriptionbiz.SubscriptionUseCase) *PublicSubscriptionService {
	return &PublicSubscriptionService{
		uc: uc,
	}
}

func (s *PublicSubscriptionService) ValidateLegacyRequest(ctx context.Context, token, requestHost, userAgent string) error {
	if s == nil || s.uc == nil {
		return nil
	}

	clients, err := s.uc.GetSubscribeApplications(ctx)
	if err != nil {
		return err
	}

	return s.uc.ValidateLegacyRequest(ctx, token, requestHost, userAgent, clients)
}

// GetSubscribeConfig 获取订阅配置
func (s *PublicSubscriptionService) GetSubscribeConfig(ctx context.Context, req *pb.GetSubscribeConfigRequest) (*pb.GetSubscribeConfigReply, error) {
	// 获取User-Agent和客户端IP（用于日志记录）
	userAgent := middleware.GetUserAgent(ctx)
	clientIP := middleware.GetClientIP(ctx)

	// 获取请求URI和Host（用于生成订阅URL）
	requestURI := middleware.GetRequestURI(ctx)
	requestHost := middleware.GetRequestHost(ctx)
	gatewayMode := middleware.GetGatewayMode(ctx)
	queryParams := middleware.GetQueryParams(ctx)

	// 调用业务层处理
	return s.uc.GetSubscribeConfig(ctx, req, userAgent, clientIP, requestURI, requestHost, gatewayMode, queryParams)
}
