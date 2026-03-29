package user

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServiceServer

	uc     *userbiz.UserUsecase
	db     *ent.Client
	logger *log.Helper
}

// NewUserService 创建用户服务
func NewUserService(uc *userbiz.UserUsecase, db *ent.Client, logger log.Logger) *UserService {
	return &UserService{
		uc:     uc,
		db:     db,
		logger: log.NewHelper(logger),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserReply, error) {
	userID, err := s.uc.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserReply{
		Code:    responsecode.AdminCreateUserSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminCreateUserSuccess],
		Data: &v1.CreateUserData{
			UserId: strconv.FormatInt(userID, 10),
		},
	}, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserReply, error) {
	err := s.uc.DeleteUser(ctx, int(parseInt64(req.UserId)))
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserReply{
		Code:    responsecode.AdminDeleteUserSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminDeleteUserSuccess],
		Data: &v1.DeleteUserData{
			Success: true,
		},
	}, nil
}

// BatchDeleteUser 批量删除用户
func (s *UserService) BatchDeleteUser(ctx context.Context, req *v1.BatchDeleteUserRequest) (*v1.BatchDeleteUserReply, error) {
	idsInt := make([]int, len(req.UserIds))
	for i, v := range req.UserIds {
		idsInt[i] = int(parseInt64(v))
	}
	deletedCount, err := s.uc.BatchDeleteUser(ctx, idsInt)
	if err != nil {
		return nil, err
	}

	return &v1.BatchDeleteUserReply{
		Code:    responsecode.AdminBatchDeleteUserSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteUserSuccess],
		Data: &v1.BatchDeleteUserData{
			DeletedCount: strconv.FormatInt(int64(deletedCount), 10),
		},
	}, nil
}

// CurrentUser 获取当前用户
func (s *UserService) CurrentUser(ctx context.Context, req *v1.CurrentUserRequest) (*v1.CurrentUserReply, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrMissingAuthToken)
	}
	user, err := s.uc.CurrentUser(ctx, int(userID))
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息
	protoUser, err := s.convertToProto(ctx, user)
	if err != nil {
		return nil, err
	}

	return &v1.CurrentUserReply{
		Code:    responsecode.AdminCurrentUserSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminCurrentUserSuccess],
		Data: &v1.CurrentUserData{
			User: protoUser,
		},
	}, nil
}

// GetUserDetail 获取用户详情
func (s *UserService) GetUserDetail(ctx context.Context, req *v1.GetUserDetailRequest) (*v1.GetUserDetailReply, error) {
	user, err := s.uc.GetUserDetail(ctx, int(parseInt64(req.Id)))
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息
	protoUser, err := s.convertToProto(ctx, user)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserDetailReply{
		Code:    responsecode.AdminGetUserDetailSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminGetUserDetailSuccess],
		Data:    protoUser, // 直接返回用户对象
	}, nil
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(ctx context.Context, req *v1.GetUserListRequest) (*v1.GetUserListReply, error) {

	var userID, subscribeID, userSubscribeID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}
	if req.SubscribeId != "" {
		parsedID := parseInt64(req.SubscribeId)
		subscribeID = &parsedID
	}
	if req.UserSubscribeId != "" {
		parsedID := parseInt64(req.UserSubscribeId)
		userSubscribeID = &parsedID
	}

	users, total, err := s.uc.GetUserList(ctx, req.Page, req.Size, req.Search, userID, subscribeID, userSubscribeID)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoUsers := make([]*v1.User, 0, len(users))
	for _, user := range users {
		protoUser, err := s.convertToProto(ctx, user)
		if err != nil {
			continue
		}
		protoUsers = append(protoUsers, protoUser)
	}

	return &v1.GetUserListReply{
		Code:    responsecode.AdminGetUserListSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminGetUserListSuccess],
		Data: &v1.GetUserListData{
			Total: strconv.FormatInt(total, 10),
			List:  protoUsers,
		},
	}, nil
}

// UpdateUserBasicInfo 更新用户基本信息
func (s *UserService) UpdateUserBasicInfo(ctx context.Context, req *v1.UpdateUserBasicInfoRequest) (*v1.UpdateUserBasicInfoReply, error) {
	err := s.uc.UpdateUserBasicInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserBasicInfoReply{
		Code:    responsecode.AdminUpdateUserBasicInfoSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminUpdateUserBasicInfoSuccess],
	}, nil
}

// UpdateUserNotifySettings 更新用户通知设置
func (s *UserService) UpdateUserNotifySettings(ctx context.Context, req *v1.UpdateUserNotifySettingsRequest) (*v1.UpdateUserNotifySettingsReply, error) {
	err := s.uc.UpdateUserNotifySettings(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserNotifySettingsReply{
		Code:    responsecode.AdminUpdateUserNotifySettingsSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminUpdateUserNotifySettingsSuccess],
	}, nil
}

// GetUserLoginLogs 获取用户登录日志
func (s *UserService) GetUserLoginLogs(ctx context.Context, req *v1.GetUserLoginLogsRequest) (*v1.GetUserLoginLogsReply, error) {

	var userID *int64
	if req.UserId != "" {
		parsedID := parseInt64(req.UserId)
		userID = &parsedID
	}

	logs, total, err := s.uc.GetUserLoginLogs(ctx, req.Page, req.Size, userID, req.Date)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoLogs := make([]*v1.LoginLog, 0, len(logs))
	for _, logEntry := range logs {
		// 解析JSON content
		var loginLog logmodel.Login
		if err := loginLog.Unmarshal([]byte(logEntry.Content)); err != nil {
			s.logger.Errorf("Failed to unmarshal login log: %v", err)
			continue
		}

		protoLog := &v1.LoginLog{
			UserId:    strconv.FormatInt(int64(logEntry.ObjectID), 10),
			Method:    loginLog.Method,
			LoginIp:   loginLog.LoginIP,
			UserAgent: loginLog.UserAgent,
			Success:   loginLog.Success,
			Timestamp: strconv.FormatInt(loginLog.Timestamp, 10),
		}

		protoLogs = append(protoLogs, protoLog)
	}

	return &v1.GetUserLoginLogsReply{
		Code:    responsecode.AdminGetUserLoginLogsSuccess,
		Message: responsecode.CodeMessages[responsecode.AdminGetUserLoginLogsSuccess],
		Data: &v1.GetUserLoginLogsData{
			Total: strconv.FormatInt(total, 10),
			List:  protoLogs,
		},
	}, nil
}

// convertToProto 将Ent实体转换为Proto消息
func (s *UserService) convertToProto(ctx context.Context, user *ent.ProxyUser) (*v1.User, error) {
	// 查询用户的认证方法
	authMethods, err := s.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(user.ID),
		).
		All(ctx)
	if err != nil {
		s.logger.Errorf("Failed to query auth methods for user %d: %v", user.ID, err)
	}

	// 查询用户设备
	userDevices, err := s.db.ProxyUserDevice.Query().
		Where(
			proxyuserdevice.UserIDEQ(user.ID),
		).
		All(ctx)
	if err != nil {
		s.logger.Errorf("Failed to query user devices for user %d: %v", user.ID, err)
	}

	// 提取邮箱、手机号和 Telegram
	var email, telephone, telephoneAreaCode string
	var telegram int64
	for _, am := range authMethods {
		if am.AuthType == "email" {
			email = am.AuthIdentifier
		} else if am.AuthType == "mobile" {
			// 解析手机号格式: {area_code}-{phone}
			telephone = am.AuthIdentifier
			// 手机号格式化处理，与原项目保持一致
		} else if am.AuthType == "telegram" {
			telegram, _ = strconv.ParseInt(am.AuthIdentifier, 10, 64)
		}
	}
	if user.Telegram != nil && *user.Telegram > 0 {
		telegram = *user.Telegram
	}

	// 转换认证方法为 proto
	protoAuthMethods := make([]*v1.UserAuthMethod, 0, len(authMethods))
	for _, am := range authMethods {
		protoAuthMethods = append(protoAuthMethods, &v1.UserAuthMethod{
			Id:             strconv.FormatInt(int64(am.ID), 10),
			UserId:         strconv.FormatInt(int64(am.UserID), 10),
			AuthType:       am.AuthType,
			AuthIdentifier: am.AuthIdentifier,
			Verified:       am.Verified,
			CreatedAt:      strconv.FormatInt(am.CreatedAt.UnixMilli(), 10),
			UpdatedAt:      strconv.FormatInt(am.UpdatedAt.UnixMilli(), 10),
		})
	}

	// 转换用户设备为 proto
	protoUserDevices := make([]*v1.UserDevice, 0, len(userDevices))
	for _, ud := range userDevices {
		// 处理设备的指针字段
		subscribeID := ""
		if ud.SubscribeID != nil {
			subscribeID = strconv.FormatInt(int64(*ud.SubscribeID), 10)
		}
		ip := ""
		if ud.IP != nil {
			ip = *ud.IP
		}
		identifier := ""
		if ud.Identifier != nil {
			identifier = *ud.Identifier
		}
		userAgent := ""
		if ud.UserAgent != nil {
			userAgent = *ud.UserAgent
		}

		protoUserDevices = append(protoUserDevices, &v1.UserDevice{
			Id:          strconv.FormatInt(int64(ud.ID), 10),
			UserId:      strconv.FormatInt(int64(ud.UserID), 10),
			SubscribeId: subscribeID,
			Ip:          ip,
			Identifier:  identifier,
			UserAgent:   userAgent,
			Online:      ud.Online,
			Enabled:     ud.Enabled,
			CreatedAt:   strconv.FormatInt(ud.CreatedAt.UnixMilli(), 10),
			UpdatedAt:   strconv.FormatInt(ud.UpdatedAt.UnixMilli(), 10),
		})
	}

	// 处理指针字段
	balance := "0"
	if user.Balance != nil {
		balance = strconv.FormatInt(int64(*user.Balance), 10)
	}
	referCode := ""
	if user.ReferCode != nil {
		referCode = *user.ReferCode
	}
	refererID := ""
	if user.RefererID != nil {
		refererID = strconv.FormatInt(int64(*user.RefererID), 10)
	}
	commission := "0"
	if user.Commission != nil {
		commission = strconv.FormatInt(int64(*user.Commission), 10)
	}
	giftAmount := "0"
	if user.GiftAmount != nil {
		giftAmount = strconv.FormatInt(int64(*user.GiftAmount), 10)
	}
	avatar := ""
	if user.Avatar != nil {
		avatar = *user.Avatar
	}

	protoUser := &v1.User{
		Id:                    strconv.FormatInt(int64(user.ID), 10),
		Email:                 email,
		Telephone:             telephone,
		TelephoneAreaCode:     telephoneAreaCode,
		Balance:               balance,
		ReferCode:             referCode,
		RefererId:             refererID,
		Commission:            commission,
		ReferralPercentage:    int32(user.ReferralPercentage),
		OnlyFirstPurchase:     user.OnlyFirstPurchase,
		GiftAmount:            giftAmount,
		Enable:                user.Enable,
		IsAdmin:               user.IsAdmin,
		EnableBalanceNotify:   user.EnableBalanceNotify,
		EnableLoginNotify:     user.EnableLoginNotify,
		EnableSubscribeNotify: user.EnableSubscribeNotify,
		EnableTradeNotify:     user.EnableTradeNotify,
		Avatar:                avatar,
		CreatedAt:             strconv.FormatInt(user.CreatedAt.UnixMilli(), 10),
		UpdatedAt:             strconv.FormatInt(user.UpdatedAt.UnixMilli(), 10),
		Telegram:              strconv.FormatInt(telegram, 10),
		AuthMethods:           protoAuthMethods,
		UserDevices:           protoUserDevices,
	}

	return protoUser, nil
}
