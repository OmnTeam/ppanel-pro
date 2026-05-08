package user

import (
	"context"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// UserRepo 用户仓库接口
type UserRepo interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, req *v1.CreateUserRequest) (int64, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, userID int) error

	// BatchDeleteUser 批量删除用户
	BatchDeleteUser(ctx context.Context, userIDs []int) (int64, error)

	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, userID int) (*ent.ProxyUser, error)

	// GetUserList 获取用户列表
	GetUserList(ctx context.Context, page, size int32, search string, userID, subscribeID, userSubscribeID *int64, unscoped bool, shortCode string) ([]*ent.ProxyUser, int32, error)

	// UpdateUserBasicInfo 更新用户基本信息
	UpdateUserBasicInfo(ctx context.Context, req *v1.UpdateUserBasicInfoRequest) error

	// UpdateUserNotifySettings 更新用户通知设置
	UpdateUserNotifySettings(ctx context.Context, req *v1.UpdateUserNotifySettingsRequest) error

	// GetUserLoginLogs 获取用户登录日志
	GetUserLoginLogs(ctx context.Context, page, size int32, userID *int64, date string) ([]*ent.ProxySystemLog, int32, error)
}

// UserUsecase 用户用例
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUsecase 创建用户用例
func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/user")),
	}
}

// CreateUser 创建用户
func (uc *UserUsecase) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (int64, error) {
	return uc.repo.CreateUser(ctx, req)
}

// DeleteUser 删除用户
func (uc *UserUsecase) DeleteUser(ctx context.Context, userID int) error {
	return uc.repo.DeleteUser(ctx, userID)
}

// BatchDeleteUser 批量删除用户
func (uc *UserUsecase) BatchDeleteUser(ctx context.Context, userIDs []int) (int64, error) {
	return uc.repo.BatchDeleteUser(ctx, userIDs)
}

// CurrentUser 获取当前用户
func (uc *UserUsecase) CurrentUser(ctx context.Context, userID int) (*ent.ProxyUser, error) {
	return uc.repo.GetUserByID(ctx, userID)
}

// GetUserDetail 获取用户详情
func (uc *UserUsecase) GetUserDetail(ctx context.Context, userID int) (*ent.ProxyUser, error) {
	return uc.repo.GetUserByID(ctx, userID)
}

// GetUserList 获取用户列表
func (uc *UserUsecase) GetUserList(ctx context.Context, page, size int32, search string, userID, subscribeID, userSubscribeID *int64, unscoped bool, shortCode string) ([]*ent.ProxyUser, int32, error) {
	return uc.repo.GetUserList(ctx, page, size, search, userID, subscribeID, userSubscribeID, unscoped, shortCode)
}

// UpdateUserBasicInfo 更新用户基本信息
func (uc *UserUsecase) UpdateUserBasicInfo(ctx context.Context, req *v1.UpdateUserBasicInfoRequest) error {
	return uc.repo.UpdateUserBasicInfo(ctx, req)
}

// UpdateUserNotifySettings 更新用户通知设置
func (uc *UserUsecase) UpdateUserNotifySettings(ctx context.Context, req *v1.UpdateUserNotifySettingsRequest) error {
	return uc.repo.UpdateUserNotifySettings(ctx, req)
}

// GetUserLoginLogs 获取用户登录日志
func (uc *UserUsecase) GetUserLoginLogs(ctx context.Context, page, size int32, userID *int64, date string) ([]*ent.ProxySystemLog, int32, error) {
	return uc.repo.GetUserLoginLogs(ctx, page, size, userID, date)
}
