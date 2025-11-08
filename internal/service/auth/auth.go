package auth

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type AuthService struct {
	pb.UnimplementedAuthServer

	uc *auth.AuthUsecase
}

func NewAuthService(uc *auth.AuthUsecase) *AuthService {
	return &AuthService{
		uc: uc,
	}
}

// CheckUser checks if user exists by email
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

// CheckUserTelephone checks if user exists by telephone
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

// UserLogin handles user login
func (s *AuthService) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.UserLogin(ctx, req.Email, req.Password, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.UserLoginSuccess),
		Message: responsecode.CodeMessages[responsecode.UserLoginSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}

// TelephoneLogin handles telephone user login
func (s *AuthService) TelephoneLogin(ctx context.Context, req *pb.TelephoneLoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneLogin(ctx, req.TelephoneAreaCode, req.Telephone, req.Password, req.TelephoneCode, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.UserLoginSuccess),
		Message: responsecode.CodeMessages[responsecode.UserLoginSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}

// UserRegister handles user registration
func (s *AuthService) UserRegister(ctx context.Context, req *pb.UserRegisterRequest) (*pb.LoginReply, error) {
	result, err := s.uc.UserRegister(ctx, req.Email, req.Password, req.Invite, req.Code, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.UserRegisterSuccess),
		Message: responsecode.CodeMessages[responsecode.UserRegisterSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}

// TelephoneRegister handles telephone user registration
func (s *AuthService) TelephoneRegister(ctx context.Context, req *pb.TelephoneRegisterRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneRegister(ctx, req.TelephoneAreaCode, req.Telephone, req.Password, req.Invite, req.Code, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.UserRegisterSuccess),
		Message: responsecode.CodeMessages[responsecode.UserRegisterSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}

// ResetPassword handles password reset
func (s *AuthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.LoginReply, error) {
	result, err := s.uc.ResetPassword(ctx, req.Email, req.Password, req.Code, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.PasswordResetSuccess),
		Message: responsecode.CodeMessages[responsecode.PasswordResetSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}

// TelephoneResetPassword handles telephone password reset
func (s *AuthService) TelephoneResetPassword(ctx context.Context, req *pb.TelephoneResetPasswordRequest) (*pb.LoginReply, error) {
	result, err := s.uc.TelephoneResetPassword(ctx, req.TelephoneAreaCode, req.Telephone, req.Password, req.Code, req.Ip, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:    int32(responsecode.PasswordResetSuccess),
		Message: responsecode.CodeMessages[responsecode.PasswordResetSuccess],
		Data: &pb.LoginData{
			Token: result.Token,
		},
	}, nil
}
