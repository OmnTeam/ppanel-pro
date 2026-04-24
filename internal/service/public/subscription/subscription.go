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
	userAgent := middleware.GetUserAgent(ctx)
	if userAgent == "" {
		userAgent = req.GetUa()
	}
	clientIP := middleware.GetClientIP(ctx)

	requestURI := middleware.GetRequestURI(ctx)
	requestHost := middleware.GetRequestHost(ctx)
	gatewayMode := middleware.GetGatewayMode(ctx)
	queryParams := middleware.GetQueryParams(ctx)

	return s.uc.GetSubscribeConfig(ctx, req, userAgent, clientIP, requestURI, requestHost, gatewayMode, queryParams)
}

func (s *PublicSubscriptionService) ResolveDownloadMeta(ctx context.Context, req *pb.GetSubscribeConfigRequest) (string, string, error) {
	userAgent := middleware.GetUserAgent(ctx)
	if userAgent == "" {
		userAgent = req.GetUa()
	}
	return s.uc.ResolveDownloadMeta(ctx, req, userAgent)
}
