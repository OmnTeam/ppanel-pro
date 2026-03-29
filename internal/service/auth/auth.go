package auth

import (
	"context"
	"strings"

	pb "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	authbiz "github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/constant"
	"github.com/go-kratos/kratos/v2/transport"
)

type AuthService struct {
	pb.UnimplementedAuthServer

	uc *authbiz.AuthUsecase
}

func NewAuthService(uc *authbiz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) CheckUser(ctx context.Context, req *pb.CheckUserRequest) (*pb.CheckUserReply, error) {
	exist, err := s.uc.CheckUser(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &pb.CheckUserReply{
		Code:    int32(responsecode.UserCheckSuccess),
		Message: responsecode.CodeMessages[responsecode.UserCheckSuccess],
		Data: &pb.CheckUserData{
			Exist: exist,
		},
	}, nil
}

func (s *AuthService) CheckUserTelephone(ctx context.Context, req *pb.CheckUserTelephoneRequest) (*pb.CheckUserTelephoneReply, error) {
	exist, err := s.uc.CheckUserTelephone(ctx, req.TelephoneAreaCode, req.Telephone)
	if err != nil {
		return nil, err
	}

	return &pb.CheckUserTelephoneReply{
		Code:    int32(responsecode.UserCheckSuccess),
		Message: responsecode.CodeMessages[responsecode.UserCheckSuccess],
		Data: &pb.CheckUserTelephoneData{
			Exist: exist,
		},
	}, nil
}

func (s *AuthService) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.UserLogin(ctx, &authbiz.UserLoginParams{
		Email:    req.Email,
		Password: req.Password,
		Meta:     buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserLoginSuccess), nil
}

func (s *AuthService) TelephoneLogin(ctx context.Context, req *pb.TelephoneLoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneLogin(ctx, &authbiz.TelephoneLoginParams{
		TelephoneAreaCode: req.TelephoneAreaCode,
		Telephone:         req.Telephone,
		Password:          req.Password,
		TelephoneCode:     req.TelephoneCode,
		Meta:              buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserLoginSuccess), nil
}

func (s *AuthService) UserRegister(ctx context.Context, req *pb.UserRegisterRequest) (*pb.LoginReply, error) {
	result, err := s.uc.UserRegister(ctx, &authbiz.UserRegisterParams{
		Email:    req.Email,
		Password: req.Password,
		Invite:   req.Invite,
		Code:     req.Code,
		Meta:     buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserRegisterSuccess), nil
}

func (s *AuthService) TelephoneRegister(ctx context.Context, req *pb.TelephoneRegisterRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneRegister(ctx, &authbiz.TelephoneRegisterParams{
		TelephoneAreaCode: req.TelephoneAreaCode,
		Telephone:         req.Telephone,
		Password:          req.Password,
		Invite:            req.Invite,
		Code:              req.Code,
		Meta:              buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserRegisterSuccess), nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.LoginReply, error) {
	result, err := s.uc.ResetPassword(ctx, &authbiz.ResetPasswordParams{
		Email:    req.Email,
		Password: req.Password,
		Code:     req.Code,
		Meta:     buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.PasswordResetSuccess), nil
}

func (s *AuthService) TelephoneResetPassword(ctx context.Context, req *pb.TelephoneResetPasswordRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneResetPassword(ctx, &authbiz.TelephoneResetPasswordParams{
		TelephoneAreaCode: req.TelephoneAreaCode,
		Telephone:         req.Telephone,
		Password:          req.Password,
		Code:              req.Code,
		Meta:              buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.PasswordResetSuccess), nil
}

func loginReply(token string, code int) *pb.LoginReply {
	return &pb.LoginReply{
		Code:    int32(code),
		Message: responsecode.CodeMessages[code],
		Data: &pb.LoginData{
			Token: token,
		},
	}
}

func buildRequestMeta(ctx context.Context, fallbackIP, fallbackUserAgent, fallbackLoginType, identifier, cfToken, captchaID, captchaCode, sliderToken string) authbiz.RequestMeta {
	meta := authbiz.RequestMeta{
		Identifier:  identifier,
		LoginType:   fallbackLoginType,
		IP:          fallbackIP,
		UserAgent:   fallbackUserAgent,
		CfToken:     cfToken,
		CaptchaID:   captchaID,
		CaptchaCode: captchaCode,
		SliderToken: sliderToken,
	}

	if tr, ok := transport.FromServerContext(ctx); ok {
		if loginType := firstHeader(tr, "Login-Type"); loginType != "" {
			meta.LoginType = loginType
		}
		if ip := firstHeader(tr, "X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"); ip != "" {
			meta.IP = firstForwardedIP(ip)
		}
		if userAgent := firstHeader(tr, "User-Agent"); userAgent != "" {
			meta.UserAgent = userAgent
		}
	}

	if meta.LoginType == "" {
		if value, ok := ctx.Value(constant.LoginType).(string); ok {
			meta.LoginType = value
		}
	}
	if meta.Identifier == "" {
		if value, ok := ctx.Value(constant.CtxKeyIdentifier).(string); ok {
			meta.Identifier = value
		}
	}

	return meta
}

func firstHeader(tr transport.Transporter, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(tr.RequestHeader().Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstForwardedIP(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	return strings.TrimSpace(parts[0])
}
