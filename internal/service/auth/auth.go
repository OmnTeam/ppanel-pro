package auth

import (
	"context"
	"strings"

	pb "github.com/OmnTeam/npanel-pro/api/public/auth/v1"
	authbiz "github.com/OmnTeam/npanel-pro/internal/biz/auth"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/OmnTeam/npanel-pro/pkg/constant"
	"github.com/go-kratos/kratos/v2/transport"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type CompatGenerateCaptchaResult struct {
	ID         string
	Image      string
	Type       string
	BlockImage string
}

type CompatSliderVerifyResult struct {
	Token string
}

type CompatDeviceLoginParams struct {
	Identifier string
	ShortCode  string
	Meta       authbiz.RequestMeta
}

type CompatAdminLoginParams struct {
	Email    string
	Password string
	Meta     authbiz.RequestMeta
}

type CompatAdminResetPasswordParams struct {
	Email    string
	Password string
	Code     string
	Meta     authbiz.RequestMeta
}

type AuthCompatProvider interface {
	GenerateCaptcha(ctx context.Context) (*CompatGenerateCaptchaResult, error)
	VerifySliderCaptcha(ctx context.Context, id string, x, y int, trail string) (*CompatSliderVerifyResult, error)
	DeviceLogin(ctx context.Context, params *CompatDeviceLoginParams) (*authbiz.LoginResult, error)
	AdminLogin(ctx context.Context, params *CompatAdminLoginParams) (*authbiz.LoginResult, error)
	AdminResetPassword(ctx context.Context, params *CompatAdminResetPasswordParams) (*authbiz.LoginResult, error)
}

type AuthService struct {
	pb.UnimplementedAuthServer

	uc         *authbiz.AuthUsecase
	authCompat AuthCompatProvider
}

func NewAuthService(uc *authbiz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) SetAuthCompat(authCompat AuthCompatProvider) {
	s.authCompat = authCompat
}

func (s *AuthService) CheckUser(ctx context.Context, req *pb.CheckUserRequest) (*pb.CheckUserReply, error) {
	exist, err := s.uc.CheckUser(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &pb.CheckUserReply{
		Exist: exist,
	}, nil
}

func (s *AuthService) CheckUserTelephone(ctx context.Context, req *pb.CheckUserTelephoneRequest) (*pb.CheckUserTelephoneReply, error) {
	exist, err := s.uc.CheckUserTelephone(ctx, req.TelephoneAreaCode, req.Telephone)
	if err != nil {
		return nil, err
	}

	return &pb.CheckUserTelephoneReply{
		Exist: exist,
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

func (s *AuthService) GenerateCaptcha(ctx context.Context, req *emptypb.Empty) (*pb.GenerateCaptchaReply, error) {
	if s.authCompat == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	result, err := s.authCompat.GenerateCaptcha(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GenerateCaptchaReply{
		Type:       result.Type,
		Id:         result.ID,
		Image:      result.Image,
		BlockImage: result.BlockImage,
	}, nil
}

func (s *AuthService) VerifySliderCaptcha(ctx context.Context, req *pb.VerifySliderCaptchaRequest) (*pb.VerifySliderCaptchaReply, error) {
	if s.authCompat == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	result, err := s.authCompat.VerifySliderCaptcha(ctx, req.Id, int(req.X), int(req.Y), req.Trail)
	if err != nil {
		return nil, err
	}

	return &pb.VerifySliderCaptchaReply{
		Token: result.Token,
	}, nil
}

func (s *AuthService) AdminGenerateCaptcha(ctx context.Context, req *emptypb.Empty) (*pb.GenerateCaptchaReply, error) {
	return s.GenerateCaptcha(ctx, req)
}

func (s *AuthService) AdminVerifySliderCaptcha(ctx context.Context, req *pb.VerifySliderCaptchaRequest) (*pb.VerifySliderCaptchaReply, error) {
	return s.VerifySliderCaptcha(ctx, req)
}

func (s *AuthService) AdminLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.LoginReply, error) {
	if s.authCompat == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	result, err := s.authCompat.AdminLogin(ctx, &CompatAdminLoginParams{
		Email:    req.Email,
		Password: req.Password,
		Meta:     buildRequestMeta(ctx, req.Ip, req.UserAgent, req.LoginType, req.Identifier, req.CfToken, req.CaptchaId, req.CaptchaCode, req.SliderToken),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserLoginSuccess), nil
}

func (s *AuthService) AdminResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.LoginReply, error) {
	if s.authCompat == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	result, err := s.authCompat.AdminResetPassword(ctx, &CompatAdminResetPasswordParams{
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

func (s *AuthService) DeviceLogin(ctx context.Context, req *pb.DeviceLoginRequest) (*pb.LoginReply, error) {
	if s.authCompat == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	result, err := s.authCompat.DeviceLogin(ctx, &CompatDeviceLoginParams{
		Identifier: req.Identifier,
		ShortCode:  req.ShortCode,
		Meta:       buildRequestMeta(ctx, req.Ip, req.UserAgent, "", req.Identifier, req.CfToken, "", "", ""),
	})
	if err != nil {
		return nil, err
	}

	return loginReply(result.Token, responsecode.UserLoginSuccess), nil
}

func loginReply(token string, code int) *pb.LoginReply {
	return &pb.LoginReply{
		Token: token,
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
