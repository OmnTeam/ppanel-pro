package service

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	pb.UnimplementedUserServer
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) QueryUserInfo(ctx context.Context, req *emptypb.Empty) (*pb.UserInfoReply, error) {
	return &pb.UserInfoReply{}, nil
}
func (s *UserService) GetLoginLog(ctx context.Context, req *pb.GetLoginLogRequest) (*pb.LoginLogReply, error) {
	return &pb.LoginLogReply{}, nil
}
func (s *UserService) QueryUserBalanceLog(ctx context.Context, req *emptypb.Empty) (*pb.BalanceLogReply, error) {
	return &pb.BalanceLogReply{}, nil
}
func (s *UserService) QueryUserCommissionLog(ctx context.Context, req *pb.QueryUserCommissionLogRequest) (*pb.CommissionLogReply, error) {
	return &pb.CommissionLogReply{}, nil
}
func (s *UserService) QueryUserAffiliate(ctx context.Context, req *emptypb.Empty) (*pb.UserAffiliateReply, error) {
	return &pb.UserAffiliateReply{}, nil
}
func (s *UserService) QueryUserAffiliateList(ctx context.Context, req *pb.QueryUserAffiliateListRequest) (*pb.UserAffiliateListReply, error) {
	return &pb.UserAffiliateListReply{}, nil
}
func (s *UserService) GetOAuthMethods(ctx context.Context, req *emptypb.Empty) (*pb.OAuthMethodsReply, error) {
	return &pb.OAuthMethodsReply{}, nil
}
func (s *UserService) QueryUserSubscribe(ctx context.Context, req *emptypb.Empty) (*pb.UserSubscribeReply, error) {
	return &pb.UserSubscribeReply{}, nil
}
func (s *UserService) GetSubscribeLog(ctx context.Context, req *pb.GetSubscribeLogRequest) (*pb.SubscribeLogReply, error) {
	return &pb.SubscribeLogReply{}, nil
}
func (s *UserService) ResetUserSubscribeToken(ctx context.Context, req *pb.ResetUserSubscribeTokenRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) PreUnsubscribe(ctx context.Context, req *pb.PreUnsubscribeRequest) (*pb.UnsubscribeInfoReply, error) {
	return &pb.UnsubscribeInfoReply{}, nil
}
func (s *UserService) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) UpdateUserNotify(ctx context.Context, req *pb.UpdateUserNotifyRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) UpdateUserPassword(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) BindTelegram(ctx context.Context, req *emptypb.Empty) (*pb.TelegramBindReply, error) {
	return &pb.TelegramBindReply{}, nil
}
func (s *UserService) UnbindTelegram(ctx context.Context, req *emptypb.Empty) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) BindOAuth(ctx context.Context, req *pb.BindOAuthRequest) (*pb.OAuthBindReply, error) {
	return &pb.OAuthBindReply{}, nil
}
func (s *UserService) BindOAuthCallback(ctx context.Context, req *pb.BindOAuthCallbackRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) UnbindOAuth(ctx context.Context, req *pb.UnbindOAuthRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) UpdateBindMobile(ctx context.Context, req *pb.UpdateBindMobileRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
func (s *UserService) UpdateBindEmail(ctx context.Context, req *pb.UpdateBindEmailRequest) (*pb.CommonReply, error) {
	return &pb.CommonReply{}, nil
}
