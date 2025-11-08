package service

import (
	"context"

	pb "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
)

type UserServiceService struct {
	pb.UnimplementedUserServiceServer
}

func NewUserServiceService() *UserServiceService {
	return &UserServiceService{}
}

func (s *UserServiceService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	return &pb.CreateUserReply{}, nil
}
func (s *UserServiceService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	return &pb.DeleteUserReply{}, nil
}
func (s *UserServiceService) BatchDeleteUser(ctx context.Context, req *pb.BatchDeleteUserRequest) (*pb.BatchDeleteUserReply, error) {
	return &pb.BatchDeleteUserReply{}, nil
}
func (s *UserServiceService) CurrentUser(ctx context.Context, req *pb.CurrentUserRequest) (*pb.CurrentUserReply, error) {
	return &pb.CurrentUserReply{}, nil
}
func (s *UserServiceService) GetUserDetail(ctx context.Context, req *pb.GetUserDetailRequest) (*pb.GetUserDetailReply, error) {
	return &pb.GetUserDetailReply{}, nil
}
func (s *UserServiceService) GetUserList(ctx context.Context, req *pb.GetUserListRequest) (*pb.GetUserListReply, error) {
	return &pb.GetUserListReply{}, nil
}
func (s *UserServiceService) UpdateUserBasicInfo(ctx context.Context, req *pb.UpdateUserBasicInfoRequest) (*pb.UpdateUserBasicInfoReply, error) {
	return &pb.UpdateUserBasicInfoReply{}, nil
}
func (s *UserServiceService) UpdateUserNotifySettings(ctx context.Context, req *pb.UpdateUserNotifySettingsRequest) (*pb.UpdateUserNotifySettingsReply, error) {
	return &pb.UpdateUserNotifySettingsReply{}, nil
}
func (s *UserServiceService) GetUserLoginLogs(ctx context.Context, req *pb.GetUserLoginLogsRequest) (*pb.GetUserLoginLogsReply, error) {
	return &pb.GetUserLoginLogsReply{}, nil
}
