package oauth

import (
	"context"
	"strings"

	pb "github.com/OmnTeam/ppanel-pro/api/auth/oauth/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/auth/oauth"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

// OAuthService OAuth服务实现
type OAuthService struct {
	pb.UnimplementedOAuthServer

	uc     oauth.OAuthUseCase
	logger *log.Helper
}

// NewOAuthService 创建OAuth服务实例
func NewOAuthService(uc oauth.OAuthUseCase, logger log.Logger) *OAuthService {
	return &OAuthService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// OAuthLogin OAuth登录 - 获取OAuth提供商的授权URL
func (s *OAuthService) OAuthLogin(ctx context.Context, req *pb.OAuthLoginRequest) (*pb.LoginRedirectReply, error) {
	s.logger.Infof("[OAuthLogin] method: %s, redirect: %s", req.Method, req.Redirect)

	params := &oauth.OAuthParams{
		Method:   req.Method,
		Redirect: req.Redirect,
	}

	result, err := s.uc.OAuthLogin(ctx, params)
	if err != nil {
		s.logger.Errorf("[OAuthLogin] failed: %v", err)
		return nil, err
	}

	return &pb.LoginRedirectReply{
		Code:    int32(responsecode.OAuthLoginSuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthLoginSuccess],
		Data: &pb.LoginRedirectData{
			Redirect: result.Redirect,
		},
	}, nil
}

// OAuthLoginGetToken OAuth登录获取令牌 - 处理OAuth回调并返回JWT token
func (s *OAuthService) OAuthLoginGetToken(ctx context.Context, req *pb.OAuthLoginGetTokenRequest) (*pb.LoginTokenReply, error) {
	ip, userAgent := buildOAuthMeta(ctx, req.Ip, req.UserAgent)
	s.logger.Infof("[OAuthLoginGetToken] method: %s, ip: %s", req.Method, ip)

	params := &oauth.OAuthTokenParams{
		Method:    req.Method,
		Callback:  req.Callback,
		IP:        ip,
		UserAgent: userAgent,
	}

	result, err := s.uc.OAuthLoginGetToken(ctx, params)
	if err != nil {
		s.logger.Errorf("[OAuthLoginGetToken] failed: %v", err)
		return nil, err
	}

	return &pb.LoginTokenReply{
		Code:    int32(responsecode.OAuthTokenGetSuccess),
		Message: responsecode.CodeMessages[responsecode.OAuthTokenGetSuccess],
		Data: &pb.LoginTokenData{
			Token: result.Token,
		},
	}, nil
}

// AppleLoginCallback Apple登录回调处理
func (s *OAuthService) AppleLoginCallback(ctx context.Context, req *pb.AppleLoginCallbackRequest) (*pb.CallbackReply, error) {
	s.logger.Infof("[AppleLoginCallback] code: %s, state: %s", req.Code, req.State)

	params := &oauth.AppleCallbackParams{
		Code:    req.Code,
		IDToken: req.IdToken,
		State:   req.State,
	}

	err := s.uc.AppleLoginCallback(ctx, params)
	if err != nil {
		s.logger.Errorf("[AppleLoginCallback] failed: %v", err)
		return nil, err
	}

	return &pb.CallbackReply{
		Code:    int32(responsecode.AppleCallbackSuccess),
		Message: responsecode.CodeMessages[responsecode.AppleCallbackSuccess],
	}, nil
}

func buildOAuthMeta(ctx context.Context, fallbackIP, fallbackUserAgent string) (string, string) {
	ip := fallbackIP
	userAgent := fallbackUserAgent

	if tr, ok := transport.FromServerContext(ctx); ok {
		if value := strings.TrimSpace(tr.RequestHeader().Get("User-Agent")); value != "" {
			userAgent = value
		}
		for _, key := range []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"} {
			if value := firstForwardedIP(strings.TrimSpace(tr.RequestHeader().Get(key))); value != "" {
				ip = value
				break
			}
		}
	}

	return ip, userAgent
}

func firstForwardedIP(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	return strings.TrimSpace(parts[0])
}
